import { sveltekit } from '@sveltejs/kit/vite'
import { defineConfig, type Plugin } from 'vite'
import codegen from 'vite-plugin-graphql-codegen'
import inspect from 'vite-plugin-inspect'

import operations from '@sourcegraph/shared/dev/generateGraphQlOperations'

function generateGraphQLOperations(): Plugin {
    const outputPath = './src/lib/graphql-operations.ts'
    const interfaceNameForOperations = 'SvelteKitGraphQlOperations'
    const documents = ['src/lib/**/*.ts', '!src/lib/graphql-operations.ts']

    return codegen({
        config: {
            ...operations.createCodegenConfig([{ interfaceNameForOperations, outputPath }]),
            // Top-level documents needs to be expliclity configured, otherwise vite-plugin-graphql-codgen
            // won't regenerate on change.
            documents,
        },
    })
}

const config = defineConfig(({ mode }) => ({
    plugins: [sveltekit(), generateGraphQLOperations(), inspect()],
    define:
        mode === 'test'
            ? {}
            : {
                  'process.platform': '"browser"',
                  'process.env.VITEST': 'undefined',
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
            '^(/sign-in|/.assets|/-|/.api|/search/stream|/users)': {
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

    test: {
        setupFiles: './src/testing/setup.ts',
    },
}))

export default config
