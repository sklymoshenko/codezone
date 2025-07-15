import { SiGo, SiJavascript, SiPostgresql } from 'solid-icons/si'
import { Component, createSignal } from 'solid-js'
import type { Language, PostgresConnectionStatus } from '../types'
import LangButton from './LangButton'
import PostgresConnectionDialog from './ui/PostgresConnectionDialog'
import PostgresDisconnectDialog from './ui/PostgresDisconnectDialog'

type LangSwitchProps = {
  currentLanguage: Language
  onLanguageChange: (lang: Language) => void
  postgresConnectionStatus: PostgresConnectionStatus
  onConnectionChange?: () => void
}

const LangSwitch: Component<LangSwitchProps> = props => {
  const [showConnectionDialog, setShowConnectionDialog] = createSignal(false)
  const [showDisconnectDialog, setShowDisconnectDialog] = createSignal(false)

  const handlePostgresClick = () => {
    if (props.postgresConnectionStatus === 'disconnected') {
      // Show connection dialog if not connected
      setShowConnectionDialog(true)
    } else {
      // Show disconnect dialog if connected, or switch to postgres if already selected
      if (props.currentLanguage === 'postgres') {
        setShowDisconnectDialog(true)
      } else {
        props.onLanguageChange('postgres')
      }
    }
  }

  const handleConnectionSuccess = (connected: boolean) => {
    // Trigger status refresh in parent component
    if (connected && props.onConnectionChange) {
      props.onConnectionChange()
    }

    if (connected) {
      // Automatically switch to postgres language after successful connection
      props.onLanguageChange('postgres')
    }
  }

  const handleDisconnectionSuccess = (disconnected: boolean) => {
    // Trigger status refresh in parent component
    if (disconnected && props.onConnectionChange) {
      props.onConnectionChange()
    }

    if (disconnected) {
      // Switch away from postgres language after disconnection
      props.onLanguageChange('javascript')
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
          activeClasses="bg-[#336791] text-white"
          hoverClasses="hover:bg-[#336791]/80"
          connectionStatus={props.postgresConnectionStatus}
        />
      </div>

      <PostgresConnectionDialog
        open={showConnectionDialog()}
        onOpenChange={setShowConnectionDialog}
        onConnect={handleConnectionSuccess}
      />

      <PostgresDisconnectDialog
        open={showDisconnectDialog()}
        onOpenChange={setShowDisconnectDialog}
        onDisconnect={handleDisconnectionSuccess}
      />
    </>
  )
}

export default LangSwitch
