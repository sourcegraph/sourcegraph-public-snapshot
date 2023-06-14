import { resolve } from 'path'

import react from '@vitejs/plugin-react'
import { defineConfig } from 'vite'
import { viteStaticCopy } from 'vite-plugin-static-copy'

// https://vitejs.dev/config/
// eslint-disable-next-line import/no-default-export
export default defineConfig({
    plugins: [
        viteStaticCopy({
            targets: [
                {
                    src: 'node_modules/cody-icons-font/font/cody-icons.woff',
                    dest: 'assets',
                },
            ],
        }),
        react(),
    ],
    publicDir: 'resources',
    base: './',
    css: {
        modules: {
            localsConvention: 'camelCaseOnly',
        },
    },
    logLevel: 'warn',
    build: {
        emptyOutDir: false,
        outDir: 'dist',
        target: 'esnext',
        minify: false,
        sourcemap: 'inline',
        reportCompressedSize: false,
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
