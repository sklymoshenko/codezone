import type { Component, JSX } from 'solid-js'

interface LangButtonProps {
  onClick: () => void
  isActive: boolean
  icon: JSX.Element
  activeClasses: string
  hoverClasses: string
}

const LangButton: Component<LangButtonProps> = props => {
  return (
    <button
      onClick={props.onClick}
      class={`flex h-7 w-7 items-center justify-center rounded-full transition-all duration-300 ${props.hoverClasses} cursor-pointer rounded-md`}
      classList={{
        'bg-gray-700 text-white': !props.isActive,
        [props.activeClasses]: props.isActive
      }}
    >
      {props.icon}
    </button>
  )
}

export default LangButton
