import {
  Component,
  createResource,
  createSignal,
  onCleanup,
  onMount,
  Show
} from 'solid-js'
import { GetPostgreSQLConnectionStatus } from 'wailsjs/go/main/App'
import { executor } from 'wailsjs/go/models'
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
import Output from './Output'
import LangSwitch from './ui/LangSwitch'
import { Resizable, ResizableHandle, ResizablePanel } from './ui/resizable'
import TitleBar, { TITLE_BAR_HEIGHT } from './ui/TitleBar'

const Main: Component = () => {
  const [env] = createResource<EnvironmentInfo>(Environment)
  const [executionResult, setExecutionResult] =
    createSignal<executor.ExecutionResult | null>(null)
  const [isExecuting, setIsExecuting] = createSignal(false)
  const [postgresConnectionStatus, setPostgresConnectionStatus] =
    createSignal<PostgresConnectionStatus>('disconnected')

  const storedLang = locStorage.get('selectedLanguage')
  const initialLang =
    storedLang && isValidLanguage(storedLang) ? storedLang : 'typescript'
  const [language, setLanguage] = createSignal<Language>(initialLang)

  const [panelSizes, setPanelSizes] = createSignal<number[]>(getStoredPanelSizes())

  let statusInterval: ReturnType<typeof setInterval> | undefined

  const checkConnectionStatus = async () => {
    try {
      const isConnected = await GetPostgreSQLConnectionStatus()
      setPostgresConnectionStatus(isConnected ? 'connected' : 'disconnected')
    } catch (err) {
      console.error('Error checking PostgreSQL connection status:', err)
      setPostgresConnectionStatus('disconnected')
    }
  }

  onMount(() => {
    // Initial status check
    void checkConnectionStatus()

    // Set up periodic status monitoring (every 5 seconds)
    statusInterval = setInterval(() => void checkConnectionStatus(), 5000)
  })

  onCleanup(() => {
    if (statusInterval) {
      clearInterval(statusInterval)
    }
  })

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

  // Provide connection status refresh function to child components
  const refreshConnectionStatus = () => {
    void checkConnectionStatus()
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
          onConnectionChange={refreshConnectionStatus}
        />
      </div>

      <div
        class={`flex-grow`}
        style={{ 'max-height': `calc(100vh - ${TITLE_BAR_HEIGHT}px)` }}
      >
        <Resizable
          orientation="horizontal"
          class="h-full w-full"
          sizes={panelSizes()}
          onSizesChange={handlePanelSizesChange}
        >
          <ResizablePanel class="min-w-[400px] h-full">
            <Editor
              language={language()}
              onLanguageChange={handleLanguageChange}
              onExecutionResult={setExecutionResult}
              onExecutionStart={() => setIsExecuting(true)}
              onExecutionEnd={() => setIsExecuting(false)}
              postgresConnectionStatus={postgresConnectionStatus()}
            />
          </ResizablePanel>
          <ResizableHandle withHandle />
          <ResizablePanel class="min-w-[200px] h-full flex flex-col">
            <Show when={executionResult() || language() === 'postgres'}>
              <Output
                isExecuting={isExecuting()}
                executionResult={executionResult()}
                language={language()}
                postgresConnectionStatus={postgresConnectionStatus()}
              />
            </Show>
          </ResizablePanel>
        </Resizable>
      </div>
    </main>
  )
}

export default Main
