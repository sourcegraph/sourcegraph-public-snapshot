import { dirname, join } from 'path'
import { fileURLToPath } from 'url'

import { sveltekit } from '@sveltejs/kit/vite'
import { defineConfig, mergeConfig, type Plugin, type UserConfig } from 'vite'
import inspect from 'vite-plugin-inspect'
import graphql from '@rollup/plugin-graphql'

const __dirname = dirname(fileURLToPath(import.meta.url))

async function generateGraphQLOperations(): Promise<Plugin> {
    const documents = ['src/lib/**/*.{ts,graphql,gql}', 'src/routes/**/*.{ts,graphql,gql}', '!src/lib/graphql-{types,operations}.ts' ]

    // We have to dynamically import this module to not make it a dependency when using
    // Bazel
    const codegen = (await import('vite-plugin-graphql-codegen')).default

    return codegen({
        // Keep in sync with client/shared/dev/generateGraphQlOperations.ts
        config: {
            generates: {
                // "legacy" graphql operations file
                './src/lib/graphql-operations.ts' : {
                    config: {
                        onlyOperationTypes: true,
                        enumValues: '@sourcegraph/shared/src/graphql-operations',
                        interfaceNameForOperations: 'SvelteKitGraphQlOperations',
                    },
                    plugins: [
                        '../shared/dev/extractGraphQlOperationCodegenPlugin.js',
                        'typescript',
                        'typescript-operations',
                    ],
                },
                // All graphql types
                "src/lib/graphql-types.ts": {
                    plugins: ['typescript'],
                },
                // GraphQL operations colocated with their source files.
                // Generates typed documents and operation types.
                "src/": {
                    preset: 'near-operation-file',
                    presetConfig: {
                        extension: '.gql.ts',
                        baseTypesPath: 'lib/graphql-types.ts',
                    },
                    plugins: ['typescript-operations', 'typed-document-node'],
                },
            },
            schema: '../../cmd/frontend/graphqlbackend/*.graphql',
            errorsOnly: true,
            silent: true,
            config: {
                // https://the-guild.dev/graphql/codegen/plugins/typescript/typescript-operations#config-api-reference
                useTypeImports: true,
                arrayInputCoercion: false,
                preResolveTypes: true,
                operationResultSuffix: 'Result',
                omitOperationSuffix: true,
                namingConvention: {
                    typeNames: 'keep',
                    enumValues: 'keep',
                    transformUnderscore: true,
                },
                declarationKind: 'interface',
                avoidOptionals: {
                    field: true,
                    inputValue: false,
                    object: true,
                },
                scalars: {
                    DateTime: 'string',
                    JSON: 'object',
                    JSONValue: 'unknown',
                    GitObjectID: 'string',
                    JSONCString: 'string',
                    PublishedValue: "boolean | 'draft'",
                    BigInt: 'string',
                },
            },
            // Top-level documents needs to be expliclity configured, otherwise vite-plugin-graphql-codgen
            // won't regenerate on change.
            documents,
        },
    })
}

export default defineConfig(({ mode }) => {
    let config: UserConfig = {
        plugins: [
            sveltekit(),
            // When using bazel the graphql operations fiel is generated
            // by bazel targets
            process.env.BAZEL ? null : generateGraphQLOperations(),
            inspect(),
            // Necessary to enable fragment imports in graphql files
            graphql(),
        ],
        define:
            mode === 'test'
                ? {}
                : {
                      'process.platform': '"browser"',
                      'process.env.VITEST': 'undefined',
                      'process.env': '({})',
                  },
        css: {
            preprocessorOptions: {
                scss: {
                    includePaths: [
                        // Our scss files and scss files in client/* often import global styles via @import 'wildcard/src/...'
                        // Adding '..' as load path causes scss to look for these imports in the client folder.
                        // (without it scss @import paths are always relative to the importing file)
                        join(__dirname, '..'),
                    ],
                },
            },
            modules: {
                localsConvention: 'camelCase',
            },
        },
        server: {
            // Allow setting the port via env variables to make it easier to integrate with
            // our existing caddy setup (which proxies requests to a specific port).
            port: process.env.SK_PORT ? +process.env.SK_PORT : undefined,
            strictPort: !!process.env.SV_PORT,
            proxy: {
                // Proxy requests to specific endpoints to a real Sourcegraph
                // instance.
                '^(/sign-in|/.assets|/-|/.api|/search/stream|/users|/notebooks|/insights)': {
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
    }

    if (process.env.BAZEL) {
        // Merge settings necessary to make the build work with bazel
        config = mergeConfig(config, {
            resolve: {
                alias: [
                    // When using Bazel, @sourcegraph/* dependencies will refer to the built packages.
                    // These do not contain the source *.module.scss files but still contain import statements
                    // that reference *.scss files. Processing them with vite throws an error unless we
                    // update the imports to reference the corresponding *.css files instead.
                    // Additionally our own source files might reference *.module.scss files, which we also want
                    // to rewrite.
                    {
                        find: /^(.+)\.module\.scss$/,
                        replacement: '$1.module.css',
                        customResolver(source, importer, options) {
                            // The this.resolve(...) part is taken from the @rollup/plugin-alias implementation. Without
                            // it it appears the bundler tries to resolve relative module IDs to the current working
                            // directory.
                            return source.includes('@sourcegraph') || importer?.includes('@sourcegraph/')
                                ? this.resolve(source, importer, { skipSelf: true, ...options }).then(
                                      resolved => resolved || { id: source }
                                  )
                                : null
                        },
                    },
                ],
            },
            ssr: {
                // By default vite treats dependencies that are links to other packages in the monorepo as source code
                // and processes them as well.
                // In a bazel sandbox however all @sourcegraph/* dependencies are built packages and thus not processed
                // by vite without this additional setting.
                // We have to process those files to apply certain "fixes", such as aliases defined in svelte.config.js.
                noExternal: [/@sourcegraph\/.*/],
            },
        } satisfies UserConfig)
    }

    return config
})
