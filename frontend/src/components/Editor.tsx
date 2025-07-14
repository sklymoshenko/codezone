import { Component, createEffect, createSignal, Show, createResource } from 'solid-js'
import { SiJavascript, SiGo, SiPostgresql } from 'solid-icons/si'
import LangButton from './LangButton'
import TitleBar from './TitleBar'
import hljs from 'highlight.js/lib/core'
import javascript from 'highlight.js/lib/languages/javascript'
import { locStorage, isValidLanguage } from '../utils/locStorage'
import { Environment } from 'wailsjs/runtime'
import type { EnvironmentInfo } from 'wailsjs/runtime'
import type { Language } from '../types'

hljs.registerLanguage('javascript', javascript)

const defaultCode: Record<Language, string> = {
  javascript:
    '// Write your code here\n\nfunction hello() {\n  console.log("Hello, World!");\n}\n\nhello()\n',
  go: `package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")
}`,
  postgres: 'SELECT * FROM tables;',
}

const Editor: Component = () => {
  const [env] = createResource<EnvironmentInfo>(Environment);
  const storedLang = locStorage.get('selectedLanguage')
  const initialLang =
    storedLang && isValidLanguage(storedLang) ? storedLang : 'javascript'

  const getCodeForLang = (lang: Language): string => {
    return locStorage.get<string>(`code-${lang}`) ?? defaultCode[lang]
  }

  const [language, setLanguage] = createSignal<Language>(initialLang)
  const [code, setCode] = createSignal<string>(getCodeForLang(initialLang))
  let codeRef: HTMLElement | undefined
  let preRef: HTMLPreElement | undefined

  createEffect(async () => {
    const lang = language()
    const currentCode = code()

    try {
      if (lang === 'go' && !hljs.getLanguage('go')) {
        const module = await import('highlight.js/lib/languages/go')
        hljs.registerLanguage('go', module.default)
      } 
      
      if (lang === 'postgres' && !hljs.getLanguage('postgres')) {
        const module = await import('highlight.js/lib/languages/sql')
        hljs.registerLanguage('postgres', module.default)
      }

      if (codeRef) {
        const highlighted = hljs.highlight(currentCode, {
          language: lang,
          ignoreIllegals: true,
        }).value
        codeRef.innerHTML = highlighted
      }
    } catch (e) {
      console.error(e)
      if (codeRef) {
        codeRef.textContent = currentCode
      }
    }
  })

  const handleLanguageChange = (lang: Language) => {
    setLanguage(lang)
    setCode(getCodeForLang(lang))
    locStorage.set('selectedLanguage', lang)
  }

  const handleInput = (e: Event) => {
    const target = e.target as HTMLTextAreaElement
    const currentCode = target.value
    setCode(currentCode)
    locStorage.set(`code-${language()}`, currentCode)
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
    <div class="flex h-full w-full flex-col">
      <Show when={!env.loading && env()?.platform === 'linux'}>
        <TitleBar />
      </Show>
      <div class="relative flex-grow overflow-hidden">
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
          class="font-mono text-base leading-normal m-0 h-full w-full overflow-auto"
          style={{
            'tab-size': '2',
            '-moz-tab-size': '2',
          }}
        >
          <code
            ref={codeRef}
            class={`language-${language()} block h-full w-full p-4`}
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
      </div>
    </div>
  )
}

export default Editor 