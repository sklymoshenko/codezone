import type { Language } from '../types'

type LocalStorageKeys = {
  selectedLanguage: string
  panelSizes: number[]
} & {
  [K in Language as `code-${K}`]: string
}

export const locStorage = {
  get<T = string>(key: keyof LocalStorageKeys): T | null {
    try {
      const item = localStorage.getItem(key)
      if (item === null) return null

      // Try to parse as JSON first, fallback to string
      try {
        return JSON.parse(item) as T
      } catch {
        return item as T
      }
    } catch (error) {
      console.warn(`Failed to get localStorage item "${key}":`, error)
      return null
    }
  },

  set<T>(key: keyof LocalStorageKeys, value: T): void {
    try {
      const stringValue = typeof value === 'string' ? value : JSON.stringify(value)
      localStorage.setItem(key, stringValue)
    } catch (error) {
      console.warn(`Failed to set localStorage item "${key}":`, error)
    }
  },

  remove(key: keyof LocalStorageKeys): void {
    try {
      localStorage.removeItem(key)
    } catch (error) {
      console.warn(`Failed to remove localStorage item "${key}":`, error)
    }
  }
}

export const isValidLanguage = (lang: string): lang is Language => {
  return ['javascript', 'go', 'postgres'].includes(lang)
}

// Panel size utilities
export const getStoredPanelSizes = (): number[] => {
  const stored = locStorage.get<number[]>('panelSizes')
  // Default to 60% editor, 40% output
  return stored && Array.isArray(stored) && stored.length === 2 ? stored : [0.6, 0.4]
}

export const storePanelSizes = (sizes: number[]): void => {
  if (sizes.length === 2) {
    locStorage.set('panelSizes', sizes)
  }
}
