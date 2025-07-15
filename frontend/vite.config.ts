import path from 'path'
import { defineConfig } from 'vite'
import solid from 'vite-plugin-solid'
import tsconfigPaths from 'vite-tsconfig-paths'

export default defineConfig({
  plugins: [solid(), tsconfigPaths()],
  server: {
    port: 3000,
    strictPort: true
  },
  build: {
    target: 'esnext'
  },
  resolve: {
    alias: {
      '~': path.resolve(__dirname, './src')
    }
  }
})
