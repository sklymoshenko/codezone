import 'typescript'

export type Language = 'typescript' | 'go' | 'postgres'
export type PostgresConnectionStatus = 'connected' | 'disconnected'

declare global {
  interface Window {
    ts: typeof import('typescript')
  }
}
