import { Show, type Component, type JSX } from 'solid-js'
import type { PostgresConnectionStatus } from '../../types'

interface LangButtonProps {
  onClick: () => void
  isActive: boolean
  icon: JSX.Element
  activeClasses: string
  hoverClasses: string
  connectionStatus?: PostgresConnectionStatus
}

const LangButton: Component<LangButtonProps> = props => {
  return (
    <button
      onClick={props.onClick}
      class={`relative flex h-7 w-7 items-center justify-center rounded-full transition-all duration-300 ${props.hoverClasses} cursor-pointer rounded-md`}
      classList={{
        'bg-muted text-muted-foreground': !props.isActive,
        [props.activeClasses]: props.isActive
      }}
    >
      {props.icon}

      <Show when={props.connectionStatus}>
        <div
          class={`absolute -top-0.5 -right-0.5 w-2 h-2 rounded-full border border-border transition-colors`}
          classList={{
            'bg-success animate-pulse': props.connectionStatus === 'connected',
            'bg-destructive': props.connectionStatus === 'disconnected'
          }}
        />
      </Show>
    </button>
  )
}

export default LangButton
