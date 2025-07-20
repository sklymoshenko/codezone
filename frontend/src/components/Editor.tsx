import { validateTsCode } from '~/utils/validate'
import hljs from 'highlight.js/lib/core'
import typescript from 'highlight.js/lib/languages/typescript'
import { Parser } from 'node-sql-parser/build/postgresql'
import { FaSolidPlay } from 'solid-icons/fa'
import { Component, createEffect, createMemo, createSignal, Show } from 'solid-js'
import { ExecuteCode, RefreshExecutor } from 'wailsjs/go/main/App'
import { executor } from 'wailsjs/go/models'
import type { Language, PostgresConnectionStatus } from '../types'
import { debounce } from '../utils/debounce'
import { locStorage } from '../utils/locStorage'
import { useUndo } from '../utils/useUndo'
import { showErrorToast } from './ui/ErrorToast'

hljs.registerLanguage('typescript', typescript)

// Helper function to show Go not installed toast
const showGoNotInstalledToast = () => {
  showErrorToast({
    title: 'Golang Not Installed',
    description: 'Go compiler is not installed or not in the PATH. Please install Golang and add it to the PATH to run Go code.',
    actionLabel: 'Download Golang',
    actionUrl: 'https://golang.org/dl/',
    duration: 8000
  })
}

const showNodeNotInstalledToast = () => {
  showErrorToast({
    title: "Can't execute TypeScript code",
    description: "The internal JavaScript runner doesn't fully support ES2016+ features (arrow functions, template literals, etc.).\n\n" +
			"Node.js is required as a fallback for modern JavaScript features.\n\n" +
			"Please install Node.js to further execute this code.",
    actionLabel: 'Download Node.js',
    actionUrl: 'https://nodejs.org/en/download/', 
    duration: 8000
  })
}

const defaultCode: Record<Language, string> = {
  typescript:
    '// TypeScript example\n\nfunction greet(name: string): string {\n  return `Hello, ${name}!`\n}\n\nconsole.log(greet("World"))\n',
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
  postgresConnectionStatus: PostgresConnectionStatus
}

const Editor: Component<EditorProps> = props => {
  const getCodeForLang = (lang: Language): string => {
    return locStorage.get<string>(`code-${lang}`) ?? defaultCode[lang]
  }

  const [code, setCode] = createSignal<string>(getCodeForLang(props.language))
  const [cursorPosition, setCursorPosition] = createSignal<number>(0)
  const [buttonPosition, setButtonPosition] = createSignal({
    display: false,
    top: 0,
    left: 0
  })

  // Add selection tracking to detect when user is actively selecting
  const [isSelecting, setIsSelecting] = createSignal<boolean>(false)

  // Initialize undo system
  const { undo, redo, recordChange, clear } = useUndo(code, setCode)

  // Initialize SQL parser for PostgreSQL
  const sqlParser = new Parser()

  // Declare refs first
  let codeRef: HTMLElement | undefined
  let preRef: HTMLPreElement | undefined
  let textareaRef: HTMLTextAreaElement | undefined

  // Validate SQL using node-sql-parser and return detailed error info
  const validateSQL = (sqlCode: string): { isValid: boolean; error: string | null } => {
    if (!sqlCode.trim()) {
      return { isValid: false, error: null }
    }

    try {
      // Try to parse as PostgreSQL
      sqlParser.astify(sqlCode)
      return { isValid: true, error: null }
    } catch (error) {
      // Parser throws error for invalid SQL
      const errorMessage = error instanceof Error ? error.message : String(error)

      // Clean up the error message for better user experience
      //@ts-expect-error - error is not typed
      const friendlyError = `${error.name}: ${errorMessage}`

      return { isValid: false, error: friendlyError }
    }
  }

  // Check if SQL operations should be enabled
  const shouldEnableSQLOperations = createMemo(() => {
    return (
      props.language === 'postgres' &&
      !isSelecting() &&
      props.postgresConnectionStatus === 'connected'
    )
  })

  // Get the nearest complete statement (with semicolon) - but only when SQL operations enabled
  const getNearestCompleteStatement = createMemo(() => {
    if (!shouldEnableSQLOperations()) {
      return null
    }

    const currentCode = code()
    const cursor = cursorPosition()

    // Look backwards for the nearest semicolon
    let nearestSemicolon = -1
    for (let i = cursor; i >= 0; i--) {
      if (currentCode[i] === ';') {
        nearestSemicolon = i
        break
      }
    }

    if (nearestSemicolon === -1) return null // No semicolon found before cursor

    // Find the start of this statement (previous semicolon or beginning)
    let start = currentCode.lastIndexOf(';', nearestSemicolon - 1)
    start = start === -1 ? 0 : start + 1

    const end = nearestSemicolon + 1 // Include the semicolon

    const statement = currentCode.substring(start, end).trim()
    return { statement, start, end }
  })

  // Check if there's meaningful and VALID SQL content to show execute button
  const hasExecutableSQL = createMemo(() => {
    if (!shouldEnableSQLOperations()) {
      if (props.language !== 'postgres') {
        props.onExecutionResult(null)
      }
      return false
    }

    const nearestStatement = getNearestCompleteStatement()

    if (!nearestStatement) {
      props.onExecutionResult(null)
      return false
    }

    // Remove comments from the statement
    const cleanStatement = nearestStatement.statement
      .split('\n')
      .map(line => line.trim())
      .filter(line => line && !line.startsWith('--'))
      .join('\n')
      .trim()

    if (cleanStatement.length === 0) {
      props.onExecutionResult(null)
      return false
    }

    // Validate the statement
    const validation = validateSQL(cleanStatement)

    if (!validation.isValid && validation.error) {
      const sqlErrorResult = new executor.ExecutionResult({
        output: '',
        error: validation.error,
        exitCode: 1,
        duration: 0,
        durationString: '0s',
        language: 'postgres'
      })

      props.onExecutionResult(sqlErrorResult)
      return false
    } else if (validation.isValid) {
      props.onExecutionResult(null)
      return true
    }

    return false
  })

  const currentStatementEnd = createMemo(() => {
    const statement = getNearestCompleteStatement()
    return statement?.end ?? -1
  })

  createEffect(() => {
    if (props.language !== 'postgres') {
      return
    }

    if (!hasExecutableSQL() || !textareaRef) {
      setButtonPosition({ display: false, top: 0, left: 0 })
      return
    }

    const endPos = currentStatementEnd()
    if (endPos === -1) {
      setButtonPosition({ display: false, top: 0, left: 0 })
      return
    }

    // Get actual computed styles from the textarea
    const computedStyle = getComputedStyle(textareaRef)
    const fontSize = parseFloat(computedStyle.fontSize)
    const lineHeight = parseFloat(computedStyle.lineHeight) || fontSize * 1.2 // fallback if 'normal'
    const fontFamily = computedStyle.fontFamily
    const paddingTop = parseFloat(computedStyle.paddingTop)
    const paddingLeft = parseFloat(computedStyle.paddingLeft)

    // Create a canvas context to measure text accurately
    const canvas = document.createElement('canvas')
    const ctx = canvas.getContext('2d')!
    ctx.font = `${fontSize}px ${fontFamily}`

    // Split text into lines up to the statement end
    const textUpToEnd = code().substring(0, endPos)
    const lines = textUpToEnd.split('\n')

    const lineNumber = lines.length - 1
    const lastLine = lines[lines.length - 1] || ''

    // Measure the actual width of the last line
    const textWidth = ctx.measureText(lastLine).width

    // Button dimensions
    const buttonHeight = 30

    // Calculate precise position - center button vertically on the line
    const top = paddingTop + lineNumber * lineHeight + lineHeight / 2 - buttonHeight / 2
    const left = paddingLeft + textWidth + 8 // Small gap after text

    // Clean up canvas
    canvas.remove()

    setButtonPosition({
      display: true,
      top: top,
      left: left
    })
  })

  const validateTs = (codeToValidate: string): boolean => {
    const errors = validateTsCode(codeToValidate)
    if (errors.length === 0) {
      return true
    }

    const errorResult = new executor.ExecutionResult({
      error: errors.join('\n'),
      exitCode: 1,
      duration: 0,
      durationString: '0s',
      language: 'typescript'
    })
    props.onExecutionResult(errorResult)
    return false
  }

  createEffect(() => {
    const newLang = props.language
    const newCode = getCodeForLang(newLang)
    setCode(newCode)
    clear()
    if (newLang === 'typescript') {
      const isValid = validateTs(newCode)
      if (!isValid) {
        return
      }
    }

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
    if (lang !== 'typescript' && lang !== 'go') {
      return
    }

    props.onExecutionStart()
    try {
      const result = await ExecuteCode(
        new executor.ExecutionConfig({
          code: codeToExecute,
          language: lang,
          timeout: 0 // 0 tells the backend to use its default timeout.
        })
      )

      // // Handle Go not installed error (exit code 150)
      if (!result.error && result.exitCode === 150 && lang === 'go') {
        showGoNotInstalledToast()
        // Don't show output for Go not installed error
        return
      }

      if (!result.error && result.exitCode === 160 && lang === 'typescript') {
        showNodeNotInstalledToast()
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
        const errorResult = new executor.ExecutionResult({
          output: '',
          error: errorMessage,
          exitCode: 1,
          duration: 0,
          durationString: '0s',
          language: lang
        })

        props.onExecutionResult(errorResult)
      }
    } finally {
      props.onExecutionEnd()
    }
  }

  // Create a debounced version of the execute function.
  const debouncedExecute = debounce((codeToExecute: string, lang: Language) => {
    if (lang === 'typescript') {
      const isValid = validateTs(codeToExecute)
      if (!isValid) {
        return
      }
    }
    void execute(codeToExecute, lang)
  }, 500)

  // Simplified highlighting effect (no validation logic needed here)
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

        // Just highlight the code
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

  // Update the textarea event handlers
  const handleInput = (e: Event) => {
    const target = e.target as HTMLTextAreaElement
    const oldCode = code()
    const newCode = target.value
    const cursorPos = target.selectionStart

    // Track cursor position
    setCursorPosition(cursorPos)
    setIsSelecting(target.selectionStart !== target.selectionEnd)

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

  // Add cursor tracking to existing events
  const handleTextareaClick = (e: MouseEvent) => {
    const target = e.target as HTMLTextAreaElement
    setCursorPosition(target.selectionStart)
    setIsSelecting(target.selectionStart !== target.selectionEnd)
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

      // Update cursor position
      setCursorPosition(start + 2)

      // FIXME: we need this because otherwise cursor will be at the end of the line
      target.selectionStart = target.selectionEnd = start + 2
    }
  }

  const handleKeyUp = (e: KeyboardEvent) => {
    const target = e.target as HTMLTextAreaElement
    setCursorPosition(target.selectionStart)
    setIsSelecting(target.selectionStart !== target.selectionEnd)
  }

  const handleMouseUp = (e: MouseEvent) => {
    const target = e.target as HTMLTextAreaElement
    setCursorPosition(target.selectionStart)
    setIsSelecting(target.selectionStart !== target.selectionEnd)
  }

  // Add selection change handler
  const handleSelectionChange = () => {
    if (textareaRef) {
      setCursorPosition(textareaRef.selectionStart)
      setIsSelecting(textareaRef.selectionStart !== textareaRef.selectionEnd)
    }
  }

  const _handleReset = async () => {
    try {
      await RefreshExecutor(props.language)
      props.onExecutionResult(null)
      void execute(code(), props.language)
    } catch (e) {
      console.error('Failed to reset environment:', e)
    }
  }

  const handleExecutePostgreSQL = async () => {
    if (!hasExecutableSQL()) {
      return
    }

    const nearestStatement = getNearestCompleteStatement()
    const statement = nearestStatement?.statement
    if (!statement) {
      return
    }

    props.onExecutionStart()

    try {
      const config = new executor.ExecutionConfig({
        code: statement,
        language: 'postgres',
        timeout: 0 // Use default timeout
      })

      const result = await ExecuteCode(config)

      if (result.sqlResult) {
        if (result.sqlResult?.columns && result.sqlResult?.rows) {
          if (result.sqlResult.rows.length > 0) {
            console.table(
              result.sqlResult.rows.map(row => {
                const rowObj: Record<string, unknown> = {}
                result.sqlResult!.columns.forEach((col, index) => {
                  rowObj[col] = row[index]
                })
                return rowObj
              })
            )
          }
        }
      }

      if (result.error) {
        console.error('SQL Error:', result.error)
      }

      props.onExecutionResult(result)
    } catch (error) {
      console.error('Failed to execute SQL:', error)
      const errorResult = new executor.ExecutionResult({
        output: '',
        error: error instanceof Error ? error.message : String(error),
        exitCode: 1,
        duration: 0,
        durationString: '0s',
        language: 'postgres'
      })

      props.onExecutionResult(errorResult)
    } finally {
      props.onExecutionEnd()
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
            onClick={handleTextareaClick}
            onKeyUp={handleKeyUp}
            onMouseUp={handleMouseUp}
            onSelectionChange={handleSelectionChange}
            spellcheck={false}
            class="hide-scrollbar absolute top-0 left-0 h-full w-full resize-none border-none bg-transparent p-4 font-mono text-base leading-normal text-transparent caret-white outline-none"
            style={{
              'white-space': 'pre-wrap',
              'tab-size': '2',
              '-moz-tab-size': '2'
            }}
          />

          {/* Floating Execute Button for PostgreSQL - only show when SQL operations enabled */}
          <Show when={buttonPosition().display && shouldEnableSQLOperations()}>
            <button
              class="absolute bg-success hover:bg-success/90 text-white rounded-md shadow-md hover:shadow-lg transition-all duration-150 hover:scale-110 z-20 flex items-center justify-center"
              style={{
                top: `${buttonPosition().top}px`,
                left: `${buttonPosition().left}px`,
                width: '28px',
                height: '28px'
              }}
              onClick={() => void handleExecutePostgreSQL()}
              title="Execute Query"
            >
              <FaSolidPlay size={16} />
            </button>
          </Show>
        </div>
      </div>
    </div>
  )
}

export default Editor
