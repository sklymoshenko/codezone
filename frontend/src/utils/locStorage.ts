import type { Language } from '../types'

type StorageKey = 'selectedLanguage' | `code-${Language}`

export const locStorage = {
  get: <T>(key: StorageKey): T | null => {
    if (typeof window === 'undefined') {
      return null
    }

    try {
      const item = window.localStorage.getItem(key)
      return item ? (JSON.parse(item) as T) : null
    } catch (error) {
      console.error(`Error getting item ${key} from localStorage`, error)
      return null
    }
  },
  set: <T>(key: StorageKey, value: T): void => {
    if (typeof window === 'undefined') {
      return
    }

    try {
      window.localStorage.setItem(key, JSON.stringify(value))
    } catch (error) {
      console.error(`Error setting item ${key} in localStorage`, error)
    }
  },
}

export const isValidLanguage = (lang: unknown): lang is Language => {
  return typeof lang === 'string' && ['javascript', 'go', 'postgres'].includes(lang)
} 