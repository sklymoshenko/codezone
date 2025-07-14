import hljs from 'highlight.js/lib/core'
import javascript from 'highlight.js/lib/languages/javascript'
import { SiGo, SiJavascript, SiPostgresql } from 'solid-icons/si'
import { Component, createEffect, createResource, createSignal, Show } from 'solid-js'
import { ExecuteCode, RefreshExecutor } from 'wailsjs/go/main/App'
import type { executor } from 'wailsjs/go/models'
import { Environment, type EnvironmentInfo } from 'wailsjs/runtime'
import type { Language } from '../types'
import { debounce } from '../utils/debounce'
import { isValidLanguage, locStorage } from '../utils/locStorage'
import { useUndo } from '../utils/useUndo'
import LangButton from './LangButton'
import TitleBar from './TitleBar'

hljs.registerLanguage('javascript', javascript)

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

const Editor: Component = () => {
  const [env] = createResource<EnvironmentInfo>(Environment)
  const storedLang = locStorage.get('selectedLanguage')
  const initialLang =
    storedLang && isValidLanguage(storedLang) ? storedLang : 'javascript'

  const getCodeForLang = (lang: Language): string => {
    return locStorage.get<string>(`code-${lang}`) ?? defaultCode[lang]
  }

  const [language, setLanguage] = createSignal<Language>(initialLang)
  const [code, setCode] = createSignal<string>(getCodeForLang(initialLang))
  const [executionResult, setExecutionResult] =
    createSignal<executor.ExecutionResult | null>(null)
  const [isExecuting, setIsExecuting] = createSignal(false)

  // Initialize undo system
  const { undo, redo, recordChange, clear } = useUndo(code, setCode)

  let codeRef: HTMLElement | undefined
  let preRef: HTMLPreElement | undefined
  let textareaRef: HTMLTextAreaElement | undefined

  // Execute initial code on component mount
  setTimeout(() => void execute(code(), language()), 0)

  // This function sends the code to the backend for execution.
  const execute = async (codeToExecute: string, lang: Language) => {
    // Don't execute if code is empty or language is not JS
    if (!codeToExecute.trim() || lang !== 'javascript') {
      setExecutionResult(null)
      return
    }

    setIsExecuting(true)
    try {
      const result = await ExecuteCode({
        code: codeToExecute,
        language: lang,
        timeout: 0 // 0 tells the backend to use its default timeout.
      })
      setExecutionResult(result)
    } catch (e: unknown) {
      // We can safely ignore the "executor is busy" error.
      // For other errors, we should display them.
      const errorMessage = e instanceof Error ? e.message : String(e)
      if (errorMessage.includes('executor is busy')) {
        // It's expected that some requests will be discarded. Do nothing.
      } else {
        setExecutionResult({
          output: '',
          error: errorMessage,
          exitCode: 1,
          duration: 0,
          durationString: '0s',
          language: lang
        })
      }
    } finally {
      setIsExecuting(false)
    }
  }

  // Create a debounced version of the execute function.
  const debouncedExecute = debounce((codeToExecute: string, lang: Language) => {
    void execute(codeToExecute, lang)
  }, 100)

  createEffect(() => {
    const lang = language()
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

  const handleLanguageChange = (lang: Language) => {
    setLanguage(lang)
    const newCode = getCodeForLang(lang)
    setCode(newCode)
    locStorage.set('selectedLanguage', lang)
    // Clear undo history when switching languages
    clear()
    // Execute the code for the new language immediately
    void execute(newCode, lang)
  }

  const handleInput = (e: Event) => {
    const target = e.target as HTMLTextAreaElement
    const oldCode = code()
    const newCode = target.value
    const cursorPos = target.selectionStart

    // Record the change for undo system
    recordChange(oldCode, newCode, cursorPos)

    // Update state and storage immediately on input
    setCode(newCode)
    locStorage.set(`code-${language()}`, newCode)
    // Trigger the debounced execution
    debouncedExecute(newCode, language())
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
      locStorage.set(`code-${language()}`, newCode)

      // FIXME: we need this because otherwise cursor will be at the end of the line
      target.selectionStart = target.selectionEnd = start + 2
    }
  }

  const _handleReset = async () => {
    try {
      await RefreshExecutor(language())
      setExecutionResult(null) // Clear the output panel
      // Re-run the current code in the new clean environment
      void execute(code(), language())
    } catch (e) {
      console.error('Failed to reset environment:', e)
    }
  }

  const hasResult = () => executionResult() !== null

  return (
    <div class="flex h-screen w-full flex-col bg-[#1f2937]">
      <Show when={!env.loading && env()?.platform === 'linux'}>
        <TitleBar />
      </Show>

      <div
        class="flex-grow flex flex-col overflow-y-auto"
        style={{ height: 'calc(100vh - 2rem)' }}
      >
        {/* Top Panel: Editor */}
        <div
          class="relative flex-grow overflow-hidden"
          style={{ height: hasResult() ? '60%' : '100%' }}
        >
          <div class="absolute top-4 right-4 z-10 flex space-x-2">
            <LangButton
              onClick={() => handleLanguageChange('javascript')}
              isActive={language() === 'javascript'}
              icon={<SiJavascript size={24} />}
              activeClasses="bg-yellow-400 text-black"
              hoverClasses="hover:bg-yellow-400 hover:text-black"
            />
            <LangButton
              onClick={() => handleLanguageChange('go')}
              isActive={language() === 'go'}
              icon={<SiGo size={24} />}
              activeClasses="bg-cyan-400 text-white"
              hoverClasses="hover:bg-cyan-400"
            />
            <LangButton
              onClick={() => handleLanguageChange('postgres')}
              isActive={language() === 'postgres'}
              icon={<SiPostgresql size={24} />}
              activeClasses="bg-blue-600 text-white"
              hoverClasses="hover:bg-blue-600"
            />
          </div>
          <pre
            ref={preRef}
            class="font-mono text-base leading-normal m-0 h-full w-full overflow-auto bg-[#111827]"
            style={{
              'tab-size': '2',
              '-moz-tab-size': '2'
            }}
          >
            <code
              ref={codeRef}
              class={`language-${language()} block h-full w-full p-4`}
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
        {/* Bottom Panel: Output */}
        <Show when={hasResult()}>
          <div
            class="flex-shrink-0 bg-[#0d1117] border-t-2 border-gray-700"
            style={{ height: '40%' }}
          >
            <div class="p-2 text-sm text-gray-400 border-b border-gray-700">
              {isExecuting()
                ? 'Executing...'
                : `Output (took ${executionResult()?.durationString})`}
            </div>
            <div class="font-mono text-sm overflow-auto px-2">
              <Show when={executionResult()?.output}>
                <pre class="text-gray-200 whitespace-pre-wrap">
                  {executionResult()?.output}
                </pre>
              </Show>
              <Show when={executionResult()?.error}>
                <pre class="text-red-400 whitespace-pre-wrap">
                  {executionResult()?.error}
                </pre>
              </Show>
            </div>
          </div>
        </Show>
      </div>
    </div>
  )
}

export default Editor
