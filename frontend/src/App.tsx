/*
 * Copyright (c) 2024-2025 Stanislav Klymoshenko
 * Licensed under the MIT License. See LICENSE file for details.
 */

import {
  ColorModeProvider,
  ColorModeScript,
  createLocalStorageManager
} from '@kobalte/core'
import type { Component } from 'solid-js'
import Main from './components/Main'
import { Toaster } from './components/ui/toast'

const App: Component = () => {
  // Create storage manager with dark mode as default
  const storageManager = createLocalStorageManager('codezone-theme')

  // Set default theme to dark if no preference is stored
  if (!localStorage.getItem('codezone-theme')) {
    localStorage.setItem('codezone-theme', 'dark')
  }

  return (
    <>
      <ColorModeScript storageType={storageManager.type} />
      <ColorModeProvider storageManager={storageManager}>
        <div class="min-h-screen flex flex-col items-center justify-center text-foreground overflow-hidden">
          <Main />
          <Toaster />
        </div>
      </ColorModeProvider>
    </>
  )
}

export default App
