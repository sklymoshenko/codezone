import { Component } from 'solid-js'
import {
  FaSolidWindowMinimize,
  FaSolidWindowMaximize,
  FaSolidXmark,
} from 'solid-icons/fa'
import {
  WindowMinimise,
  WindowToggleMaximise,
  Quit,
} from 'wailsjs/runtime'

const TitleBar: Component = () => {
  return (
    <div
      style={{ '--wails-draggable': 'drag' }}
      class="flex h-9 flex-shrink-0 cursor-move items-center justify-end bg-gray-800"
    >
      <div class="flex space-x-2 py-2">
        <button
          onClick={WindowMinimise}
          class="flex h-8 w-8 items-center justify-center rounded-md text-white transition-colors hover:bg-gray-700 cursor-pointer"
        >
          <FaSolidWindowMinimize />
        </button>
        <button
          onClick={WindowToggleMaximise}
          class="flex h-8 w-8 items-center justify-center rounded-md text-white transition-colors hover:bg-gray-700 cursor-pointer"
        >
          <FaSolidWindowMaximize />
        </button>
        <button
          onClick={Quit}
          class="flex h-8 w-8 items-center justify-center rounded-md text-white transition-colors hover:bg-red-600 cursor-pointer"
        >
          <FaSolidXmark />
        </button>
      </div>
    </div>
  )
}

export default TitleBar 