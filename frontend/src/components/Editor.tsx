import hljs from 'highlight.js/lib/core'
import javascript from 'highlight.js/lib/languages/javascript'
import { Component, createEffect, createSignal } from 'solid-js'
import { ExecuteCode, RefreshExecutor } from 'wailsjs/go/main/App'
import type { executor } from 'wailsjs/go/models'
import type { Language } from '../types'
import { debounce } from '../utils/debounce'
import { locStorage } from '../utils/locStorage'
import { useUndo } from '../utils/useUndo'
import { showErrorToast } from './ui/ErrorToast'

hljs.registerLanguage('javascript', javascript)

// Helper function to show Go not installed toast
const showGoNotInstalledToast = () => {
  showErrorToast({
    title: 'Golang Not Installed',
    description: 'Go compiler is not installed. Please install Golang to run Go code.',
    actionLabel: 'Download Golang',
    actionUrl: 'https://golang.org/dl/',
    duration: 8000
  })
}

const defaultCode: Record<Language, string> = {
  javascript:
    '// Write your code here\n\nfunction hello() {\n  console.log("Hello, World!");\n}\n\nhello()\n',
  go: `package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")
}`,
  postgres: 'SELECT * FROM tables;'
}

type EditorProps = {
  language: Language
  onLanguageChange: (lang: Language) => void
  onExecutionResult: (result: executor.ExecutionResult | null) => void
  onExecutionStart: () => void
  onExecutionEnd: () => void
}

const Editor: Component<EditorProps> = props => {
  const getCodeForLang = (lang: Language): string => {
    return locStorage.get<string>(`code-${lang}`) ?? defaultCode[lang]
  }

  const [code, setCode] = createSignal<string>(getCodeForLang(props.language))

  // Initialize undo system
  const { undo, redo, recordChange, clear } = useUndo(code, setCode)

  let codeRef: HTMLElement | undefined
  let preRef: HTMLPreElement | undefined
  let textareaRef: HTMLTextAreaElement | undefined

  // Watch for language changes from parent (also handles initial execution)
  createEffect(() => {
    const newLang = props.language
    const newCode = getCodeForLang(newLang)
    setCode(newCode)
    // Clear undo history when switching languages
    clear()
    // Execute the code for the new language immediately
    void execute(newCode, newLang)
  })

  // This function sends the code to the backend for execution.
  const execute = async (codeToExecute: string, lang: Language) => {
    // Don't execute if code is empty
    if (!codeToExecute.trim()) {
      props.onExecutionResult(null)
      return
    }

    // Only execute for supported languages (javascript and go)
    if (lang !== 'javascript' && lang !== 'go') {
      props.onExecutionResult(null)
      return
    }

    props.onExecutionStart()
    try {
      const result = await ExecuteCode({
        code: codeToExecute,
        language: lang,
        timeout: 0 // 0 tells the backend to use its default timeout.
      })

      // // Handle Go not installed error (exit code 150)
      if (!result.error && result.exitCode === 150 && lang === 'go') {
        showGoNotInstalledToast()
        // Don't show output for Go not installed error
        return
      }

      props.onExecutionResult(result)
    } catch (e: unknown) {
      // We can safely ignore the "executor is busy" error.
      // For other errors, we should display them.
      const errorMessage = e instanceof Error ? e.message : String(e)
      if (errorMessage.includes('executor is busy')) {
        // It's expected that some requests will be discarded. Do nothing.
      } else {
        const errorResult = {
          output: '',
          error: errorMessage,
          exitCode: 1,
          duration: 0,
          durationString: '0s',
          language: lang
        }

        props.onExecutionResult(errorResult)
      }
    } finally {
      props.onExecutionEnd()
    }
  }

  // Create a debounced version of the execute function.
  const debouncedExecute = debounce((codeToExecute: string, lang: Language) => {
    void execute(codeToExecute, lang)
  }, 100)

  createEffect(() => {
    const lang = props.language
    const currentCode = code()

    // Dynamically load language highlighters and highlight code
    const loadAndHighlight = async () => {
      try {
        if (lang === 'go' && !hljs.getLanguage('go')) {
          const module = await import('highlight.js/lib/languages/go')
          hljs.registerLanguage('go', module.default)
        }

        if (lang === 'postgres' && !hljs.getLanguage('postgres')) {
          const module = await import('highlight.js/lib/languages/sql')
          hljs.registerLanguage('postgres', module.default)
        }

        // Highlight the code
        if (codeRef) {
          const highlighted = hljs.highlight(currentCode, {
            language: lang,
            ignoreIllegals: true
          }).value
          codeRef.innerHTML = highlighted
        }
      } catch (e) {
        console.error(e)
        if (codeRef) {
          codeRef.textContent = currentCode
        }
      }
    }

    void loadAndHighlight()
  })

  const handleInput = (e: Event) => {
    const target = e.target as HTMLTextAreaElement
    const oldCode = code()
    const newCode = target.value
    const cursorPos = target.selectionStart

    // Record the change for undo system
    recordChange(oldCode, newCode, cursorPos)

    // Update state and storage immediately on input
    setCode(newCode)
    locStorage.set(`code-${props.language}`, newCode)
    // Trigger the debounced execution
    debouncedExecute(newCode, props.language)
  }

  const handleScroll = (e: Event) => {
    const target = e.target as HTMLTextAreaElement
    if (preRef) {
      preRef.scrollTop = target.scrollTop
      preRef.scrollLeft = target.scrollLeft
    }
  }

  const handleKeyDown = (e: KeyboardEvent) => {
    // Handle undo/redo shortcuts
    if (e.ctrlKey || e.metaKey) {
      if (e.key === 'z' && !e.shiftKey) {
        e.preventDefault()
        undo()
        return
      }
      if ((e.key === 'z' && e.shiftKey) || e.key === 'y') {
        e.preventDefault()
        redo()
        return
      }
    }

    if (e.key === 'Tab') {
      e.preventDefault()
      const target = e.target as HTMLTextAreaElement
      const start = target.selectionStart
      const end = target.selectionEnd
      const oldCode = code()
      const newCode = `${oldCode.substring(0, start)}  ${oldCode.substring(end)}`

      // Record the change for undo
      recordChange(oldCode, newCode, start + 2)

      setCode(newCode)
      locStorage.set(`code-${props.language}`, newCode)

      // FIXME: we need this because otherwise cursor will be at the end of the line
      target.selectionStart = target.selectionEnd = start + 2
    }
  }

  const _handleReset = async () => {
    try {
      await RefreshExecutor(props.language)
      props.onExecutionResult(null) // Clear the output panel
      // Re-run the current code in the new clean environment
      void execute(code(), props.language)
    } catch (e) {
      console.error('Failed to reset environment:', e)
    }
  }

  return (
    <div class="flex h-full w-full flex-col">
      <div class="flex-grow flex flex-col overflow-y-auto">
        <div class="relative flex-grow overflow-hidden">
          <pre
            ref={preRef}
            class="font-mono text-base leading-normal m-0 h-full w-full overflow-auto"
            style={{
              'tab-size': '2',
              '-moz-tab-size': '2'
            }}
          >
            <code
              ref={codeRef}
              class={`language-${props.language} block h-full w-full p-4`}
              style={{ 'white-space': 'pre-wrap' }}
            />
          </pre>
          <textarea
            ref={textareaRef}
            value={code()}
            onInput={handleInput}
            onScroll={handleScroll}
            onKeyDown={handleKeyDown}
            spellcheck={false}
            class="hide-scrollbar absolute top-0 left-0 h-full w-full resize-none border-none bg-transparent p-4 font-mono text-base leading-normal text-transparent caret-white outline-none"
            style={{
              'white-space': 'pre-wrap',
              'tab-size': '2',
              '-moz-tab-size': '2'
            }}
          />
        </div>
      </div>
    </div>
  )
}

export default Editor
