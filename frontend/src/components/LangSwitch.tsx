import { SiGo, SiJavascript, SiPostgresql } from 'solid-icons/si'
import { Component, createSignal } from 'solid-js'
import type { Language, PostgresConnectionStatus } from '../types'
import LangButton from './LangButton'
import PostgresConnectionDialog from './ui/PostgresConnectionDialog'

type LangSwitchProps = {
  currentLanguage: Language
  onLanguageChange: (lang: Language) => void
  postgresConnectionStatus: PostgresConnectionStatus
}

const LangSwitch: Component<LangSwitchProps> = props => {
  // PostgreSQL connection state - default disconnected
  const [postgresConnectionStatus, setPostgresConnectionStatus] = createSignal<
    'connected' | 'disconnected'
  >('disconnected')
  const [showConnectionDialog, setShowConnectionDialog] = createSignal(false)

  const handlePostgresClick = () => {
    if (postgresConnectionStatus() === 'disconnected') {
      // Show connection dialog if not connected
      setShowConnectionDialog(true)
    } else {
      // Switch to postgres language if already connected
      props.onLanguageChange('postgres')
    }
  }

  const handleConnectionSuccess = (connected: boolean) => {
    setPostgresConnectionStatus(connected ? 'connected' : 'disconnected')
    if (connected) {
      // Automatically switch to postgres language after successful connection
      props.onLanguageChange('postgres')
    }
  }

  return (
    <>
      <div class="flex space-x-2">
        <LangButton
          onClick={() => props.onLanguageChange('javascript')}
          isActive={props.currentLanguage === 'javascript'}
          icon={<SiJavascript size={16} />}
          activeClasses="bg-yellow-400 text-black"
          hoverClasses="hover:bg-yellow-400 hover:text-black"
        />
        <LangButton
          onClick={() => props.onLanguageChange('go')}
          isActive={props.currentLanguage === 'go'}
          icon={<SiGo size={16} />}
          activeClasses="bg-cyan-400 text-white"
          hoverClasses="hover:bg-cyan-400"
        />
        <LangButton
          onClick={handlePostgresClick}
          isActive={props.currentLanguage === 'postgres'}
          icon={<SiPostgresql size={16} />}
          activeClasses="bg-primary text-primary-foreground"
          hoverClasses="hover:bg-primary/80"
          connectionStatus={postgresConnectionStatus()} // Add connection status
        />
      </div>

      <PostgresConnectionDialog
        open={showConnectionDialog()}
        onOpenChange={setShowConnectionDialog}
        onConnect={handleConnectionSuccess}
      />
    </>
  )
}

export default LangSwitch
