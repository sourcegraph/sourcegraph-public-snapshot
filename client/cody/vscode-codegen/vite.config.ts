import { resolve } from 'path'

import react from '@vitejs/plugin-react'
import { defineConfig } from 'vite'

// https://vitejs.dev/config/
export default defineConfig({
    plugins: [react()],
    publicDir: 'resources',
    base: './',
    build: {
        emptyOutDir: false,
        outDir: 'dist',
        rollupOptions: {
            external: [/^vscode/],
            watch: {
                // https://rollupjs.org/configuration-options/#watch
                include: ['webviews/**'],
                exclude: ['node_modules', 'src'],
            },
            input: {
                main: resolve(__dirname, 'index.html'),
            },
            output: {
                entryFileNames: '[name].js',
            },
        },
    },
})
