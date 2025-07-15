import { Component, createResource, createSignal, Show } from 'solid-js'
import type { executor } from 'wailsjs/go/models'
import { Environment, type EnvironmentInfo } from 'wailsjs/runtime'
import type { Language, PostgresConnectionStatus } from '../types'
import { debounce } from '../utils/debounce'
import {
  getStoredPanelSizes,
  isValidLanguage,
  locStorage,
  storePanelSizes
} from '../utils/locStorage'
import Editor from './Editor'
import LangSwitch from './LangSwitch'
import Output from './Output'
import TitleBar from './TitleBar'
import { Resizable, ResizableHandle, ResizablePanel } from './ui/resizable'

const Main: Component = () => {
  const [env] = createResource<EnvironmentInfo>(Environment)
  const [executionResult, setExecutionResult] =
    createSignal<executor.ExecutionResult | null>(null)
  const [isExecuting, setIsExecuting] = createSignal(false)
  const [postgresConnectionStatus] =
    createSignal<PostgresConnectionStatus>('disconnected')

  // Language state management
  const storedLang = locStorage.get('selectedLanguage')
  const initialLang =
    storedLang && isValidLanguage(storedLang) ? storedLang : 'javascript'
  const [language, setLanguage] = createSignal<Language>(initialLang)

  const [panelSizes, setPanelSizes] = createSignal<number[]>(getStoredPanelSizes())

  const handleLanguageChange = (lang: Language) => {
    setLanguage(lang)
    locStorage.set('selectedLanguage', lang)
  }

  // Create debounced save function to prevent conflicts during dragging
  const debouncedSavePanelSizes = debounce(storePanelSizes, 100)

  const handlePanelSizesChange = (sizes: number[]) => {
    if (sizes.length === 2 && sizes.every(size => size > 0 && size <= 1)) {
      setPanelSizes(sizes)
      debouncedSavePanelSizes(sizes)
    }
  }

  return (
    <main class="h-screen w-screen flex flex-col relative">
      <Show when={!env.loading && env()?.platform === 'linux'}>
        <TitleBar />
      </Show>

      {/* Language switcher in top-right corner */}
      <div class="absolute top-11 right-1 z-50">
        <LangSwitch
          currentLanguage={language()}
          onLanguageChange={handleLanguageChange}
          postgresConnectionStatus={postgresConnectionStatus()}
        />
      </div>

      <div class="flex-grow">
        <Resizable
          orientation="horizontal"
          class="h-full w-full"
          sizes={panelSizes()}
          onSizesChange={handlePanelSizesChange}
        >
          <ResizablePanel class="min-w-[400px]">
            <Editor
              language={language()}
              onLanguageChange={handleLanguageChange}
              onExecutionResult={setExecutionResult}
              onExecutionStart={() => setIsExecuting(true)}
              onExecutionEnd={() => setIsExecuting(false)}
            />
          </ResizablePanel>
          <ResizableHandle withHandle />
          <ResizablePanel class="min-w-[200px]">
            <Show when={executionResult()}>
              <Output isExecuting={isExecuting()} executionResult={executionResult()!} />
            </Show>
          </ResizablePanel>
        </Resizable>
      </div>
    </main>
  )
}

export default Main
