import { sveltekit } from '@sveltejs/kit/vite'
import { defineConfig } from 'vite'

const config = defineConfig(({ mode }) => ({
    plugins: [sveltekit()],
    define:
        mode === 'test'
            ? {}
            : {
                  'process.platform': '"browser"',
                  'process.env': '{}',
              },
    css: {
        modules: {
            localsConvention: 'camelCase',
        },
    },
    server: {
        proxy: {
            // Proxy requests to specific endpoints to a real Sourcegraph
            // instance.
            '^(/sign-in|/.assets|/-|/.api|/search/stream)': {
                target: process.env.SOURCEGRAPH_API_URL || 'https://sourcegraph.com',
                changeOrigin: true,
                secure: false,
            },
        },
    },
    optimizeDeps: {
        exclude: [
            // Without addings this Vite throws an error
            'linguist-languages',
        ],
    },
}))

export default config
