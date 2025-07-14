import Main from './components/Main'
import type { Component } from 'solid-js'

const App: Component = () => {
  return (
    <div class="min-h-screen flex flex-col items-center justify-center bg-gray-900 text-white">
      <Main />
    </div>
  )
}

export default App
