
import { Component, createSignal, createEffect } from 'solid-js'
import hljs from 'highlight.js/lib/core'
import javascript from 'highlight.js/lib/languages/javascript'

hljs.registerLanguage('javascript', javascript)

const Main: Component = () => {
  const [code, setCode] = createSignal(
    '// Write your code here\n\nfunction hello() {\n  console.log("Hello, World!");\n}',
  )
  let codeRef: HTMLElement | undefined
  let preRef: HTMLPreElement | undefined

  createEffect(() => {
    if (codeRef) {
      const highlighted = hljs.highlight(code(), {
        language: 'javascript',
        ignoreIllegals: true,
      }).value
      codeRef.innerHTML = highlighted
    }
  })

  const handleInput = (e: Event) => {
    const target = e.target as HTMLTextAreaElement
    setCode(target.value)
  }

  const handleScroll = (e: Event) => {
    const target = e.target as HTMLTextAreaElement
    if (preRef) {
      preRef.scrollTop = target.scrollTop
      preRef.scrollLeft = target.scrollLeft
    }
  }

  const handleKeyDown = (e: KeyboardEvent) => {
    if (e.key === 'Tab') {
      e.preventDefault()
      const target = e.target as HTMLTextAreaElement
      const start = target.selectionStart
      const end = target.selectionEnd
      const newCode = `${code().substring(0, start)}  ${code().substring(end)}`
      setCode(newCode)

      //
      // FIXME: we need this because otherwise cursor will be at the end of the line
      //
      target.selectionStart = target.selectionEnd = start + 2
    }
  }

  return (
    <main class="relative h-screen w-screen overflow-hidden">
      <pre
        ref={preRef}
        class="font-mono text-base leading-normal m-0 h-full w-full overflow-auto"
        style={{
          'tab-size': '2',
          '-moz-tab-size': '2',
        }}
      >
        <code
          ref={codeRef}
          class="language-javascript block h-full w-full p-4"
          style={{ 'white-space': 'pre-wrap' }}
        />
      </pre>
      <textarea
        value={code()}
        onInput={handleInput}
        onScroll={handleScroll}
        onKeyDown={handleKeyDown}
        spellcheck={false}
        class="hide-scrollbar absolute top-0 left-0 h-full w-full resize-none border-none bg-transparent p-4 font-mono text-base leading-normal text-transparent caret-white outline-none"
        style={{
          'white-space': 'pre-wrap',
          'tab-size': '2',
          '-moz-tab-size': '2',
        }}
      />
    </main>
  )
}

export default Main
