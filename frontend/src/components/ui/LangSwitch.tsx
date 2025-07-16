import { SiGo, SiPostgresql, SiTypescript } from 'solid-icons/si'
import { Component, createEffect, createSignal } from 'solid-js'
import { GetGoVersion } from '../../../wailsjs/go/main/App'
import type { Language, PostgresConnectionStatus } from '../../types'
import LangButton from './LangButton'
import PostgresConnectionDialog from './PostgresConnectionDialog'
import PostgresDisconnectDialog from './PostgresDisconnectDialog'
import { Tooltip, TooltipContent, TooltipTrigger } from './tooltip'

type LangSwitchProps = {
  currentLanguage: Language
  onLanguageChange: (lang: Language) => void
  postgresConnectionStatus: PostgresConnectionStatus
  onConnectionChange?: () => void
}

const LangSwitch: Component<LangSwitchProps> = props => {
  const [showConnectionDialog, setShowConnectionDialog] = createSignal(false)
  const [showDisconnectDialog, setShowDisconnectDialog] = createSignal(false)
  const [goVersion, setGoVersion] = createSignal('Loading Go version...')
  const getGoVersion = async () => {
    const version = await GetGoVersion()
    setGoVersion(version)
  }
  createEffect(() => {
    void getGoVersion()
  })

  const handlePostgresClick = () => {
    if (props.postgresConnectionStatus === 'disconnected') {
      setShowConnectionDialog(true)
    } else {
      if (props.currentLanguage === 'postgres') {
        setShowDisconnectDialog(true)
      } else {
        props.onLanguageChange('postgres')
      }
    }
  }

  const handleConnectionSuccess = (connected: boolean) => {
    if (connected && props.onConnectionChange) {
      props.onConnectionChange()
    }

    if (connected) {
      props.onLanguageChange('postgres')
    }
  }

  const handleDisconnectionSuccess = (disconnected: boolean) => {
    if (disconnected && props.onConnectionChange) {
      props.onConnectionChange()
    }
  }

  return (
    <>
      <div class="flex space-x-2">
        <Tooltip>
          <TooltipTrigger>
            <LangButton
              onClick={() => props.onLanguageChange('typescript')}
              isActive={props.currentLanguage === 'typescript'}
              icon={<SiTypescript size={16} />}
              activeClasses="bg-blue-500 text-white"
              hoverClasses="hover:bg-blue-500/80"
            />
          </TooltipTrigger>
          <TooltipContent>TypeScript 5.8.3</TooltipContent>
        </Tooltip>

        <Tooltip>
          <TooltipTrigger>
            <LangButton
              onClick={() => props.onLanguageChange('go')}
              isActive={props.currentLanguage === 'go'}
              icon={<SiGo size={16} />}
              activeClasses="bg-cyan-400 text-white"
              hoverClasses="hover:bg-cyan-400"
            />
          </TooltipTrigger>
          <TooltipContent>{goVersion()}</TooltipContent>
        </Tooltip>

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
