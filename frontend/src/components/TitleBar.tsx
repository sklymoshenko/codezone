import {
  FaSolidWindowMaximize,
  FaSolidWindowMinimize,
  FaSolidXmark
} from 'solid-icons/fa'
import { Component } from 'solid-js'
import { Quit, WindowMinimise, WindowToggleMaximise } from 'wailsjs/runtime'

const TitleBar: Component = () => {
  return (
    <div
      style={{ '--wails-draggable': 'drag' }}
      class="flex h-fit flex-shrink-0 cursor-move items-center justify-end bg-muted border-b border-border"
    >
      <div class="flex gap-4 py-1 mr-2" style={{ '--wails-draggable': 'none' }}>
        <button
          onClick={WindowMinimise}
          class="flex h-4 w-4 items-center justify-center rounded-md text-foreground transition-colors hover:text-foreground/70 cursor-pointer p-2"
        >
          <FaSolidWindowMinimize />
        </button>
        <button
          onClick={WindowToggleMaximise}
          class="flex h-4 w-4 items-center justify-center rounded-md text-foreground transition-colors hover:text-foreground/70 cursor-pointer p-2"
        >
          <FaSolidWindowMaximize />
        </button>
        <button
          onClick={Quit}
          class="flex h-4 w-4 items-center justify-center rounded-md text-foreground transition-colors hover:text-destructive cursor-pointer p-2"
        >
          <FaSolidXmark />
        </button>
      </div>
    </div>
  )
}

export default TitleBar
