import type { Component } from 'solid-js'
import Main from './components/Main'
import { Toaster } from './components/ui/toast'

const App: Component = () => {
  return (
    <div class="min-h-screen flex flex-col items-center justify-center bg-gray-900 text-white overflow-hidden">
      <Main />
      <Toaster />
    </div>
  )
}

export default App
