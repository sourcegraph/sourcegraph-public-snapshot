import { resolve } from 'path'

import react from '@vitejs/plugin-react'
import { defineConfig } from 'vite'

import copyContentStyle from './utils/plugins/copy-content-style'
import makeManifest from './utils/plugins/make-manifest'

const root = resolve(__dirname, 'src')
const pagesDir = resolve(root, 'pages')
const assetsDir = resolve(root, 'assets')
const outDir = resolve(__dirname, 'dist')
const publicDir = resolve(__dirname, 'public')

export default defineConfig({
    resolve: {
        alias: {
            '@src': root,
            '@assets': assetsDir,
            '@pages': pagesDir,
        },
    },
    plugins: [react(), makeManifest(), copyContentStyle()],
    publicDir,
    build: {
        outDir,
        sourcemap: process.env.__DEV__ === 'true',
        rollupOptions: {
            external: [/^chrome/],
            watch: {
                // https://rollupjs.org/configuration-options/#watch
                include: ['src/**'],
                exclude: ['node_modules'],
            },
            input: {
                content: resolve(pagesDir, 'content', 'index.ts'),
                background: resolve(pagesDir, 'background', 'index.ts'),
                popup: resolve(pagesDir, 'popup', 'index.html'),
                newtab: resolve(pagesDir, 'newtab', 'index.html'),
                options: resolve(pagesDir, 'options', 'index.html'),
            },
            output: {
                entryFileNames: chunk => `src/pages/${chunk.name}/index.js`,
            },
        },
    },
})
