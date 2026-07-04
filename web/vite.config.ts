import { defineConfig } from 'vite'

export default defineConfig({
  root: '.',
  base: '/dist/',
  build: {
    outDir: 'dist',
    emptyOutDir: true,
    rollupOptions: {
      input: {
        main: 'index.html',
        sw: 'src/sw.ts',
      },
      output: {
        entryFileNames: (chunk) => {
          return chunk.name === 'sw' ? 'sw.js' : 'assets/[name]-[hash].js'
        },
      },
    },
  },
  server: {
    port: 5173,
  },
})
