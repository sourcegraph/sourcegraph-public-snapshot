import { defineConfig } from 'vite'
import { resolve } from 'path'
import react from '@vitejs/plugin-react'

// https://vitejs.dev/config/
export default defineConfig({
    plugins: [react()],
    base: '/kodj/',
    build: {
        emptyOutDir: false,
        sourcemap: true,
        outDir: 'dist',
        rollupOptions: {
            external: [/^vscode/],
            watch: {
                // https://rollupjs.org/configuration-options/#watch
                include: ['src/**', 'src/extension.ts'],
                exclude: ['node_modules/**', '../../node_modules/**'],
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
