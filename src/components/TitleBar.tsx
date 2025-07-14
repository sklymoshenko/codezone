import { Component } from 'solid-js'
import {
  FaSolidWindowMinimize,
  FaSolidWindowMaximize,
  FaSolidXmark,
} from 'solid-icons/fa'
import { invoke } from '@tauri-apps/api/core'

const TitleBar: Component = () => {
  const handleMouseDown = (e: MouseEvent) => {
    // prevent dragging when clicking buttons
    if (e.target !== e.currentTarget) return
    void invoke('start_dragging').catch(console.error)
  }

  return (
    <div
      onMouseDown={handleMouseDown}
      class="flex h-9 flex-shrink-0 cursor-move items-center justify-end bg-gray-800"
    >
      <div class="flex space-x-2 py-2">
        <button
          onClick={() => void invoke('window_minimize').catch(console.error)}
          class="flex h-8 w-8 items-center justify-center rounded-md text-white transition-colors hover:bg-gray-700 cursor-pointer"
        >
          <FaSolidWindowMinimize />
        </button>
        <button
          onClick={() =>
            void invoke('window_toggle_maximize').catch(console.error)
          }
          class="flex h-8 w-8 items-center justify-center rounded-md text-white transition-colors hover:bg-gray-700 cursor-pointer"
        >
          <FaSolidWindowMaximize />
        </button>
        <button
          onClick={() => void invoke('window_close').catch(console.error)}
          class="flex h-8 w-8 items-center justify-center rounded-md text-white transition-colors hover:bg-red-600 cursor-pointer"
        >
          <FaSolidXmark />
        </button>
      </div>
    </div>
  )
}

export default TitleBar 