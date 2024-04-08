import { generate, type CodegenConfig } from '@graphql-codegen/cli'
import type { Plugin } from 'vite'

const codgegenConfig: CodegenConfig = {
    hooks: {
        // This hook removes the 'import type * as Types from ...' import from generated files if it's not used.
        // The near-operation-file preset generates this import for every file, even if it's not used. This
        // generally is not a problem, but the issue is reported by `pnpm check`.
        // See https://github.com/dotansimha/graphql-code-generator/issues/4900
        beforeOneFileWrite(_path, content) {
            if (/^import type \* as Types from/m.test(content) && !/Types(\[|\.)/.test(content)) {
                return content.replace(/^import type \* as Types from .+$/m, '').trimStart()
            }
            return content
        },
    },
    generates: {
        // Legacy graphql-operations.ts file that is still used by some components.
        './src/lib/graphql-operations.ts': {
            documents: [
                'src/{lib,routes}/**/*.ts',
                '!src/lib/graphql-{operations,types,type-mocks}.ts',
                '!src/**/*.gql.ts',
            ],
            config: {
                onlyOperationTypes: true,
                enumValues: '$lib/graphql-types',
            },
            plugins: ['typescript', 'typescript-operations'],
        },
        'src/lib/graphql-types.ts': {
            plugins: ['typescript'],
        },
        'src/testing/graphql-type-mocks.ts': {
            documents: [
                'src/{lib,routes}/**/*.(ts|gql)',
                '!src/lib/graphql-{operations,types,type-mocks}.ts',
                '!src/**/*.gql.ts',
            ],
            config: {
                typesImport: '$lib/graphql-types',
                onlyOperationTypes: true,
            },
            plugins: [`typescript`, `typescript-operations`, `./dev/graphql-type-mocks.cjs`],
        },
        'src/': {
            documents: ['src/**/*.gql', '!src/**/*.gql.ts'],
            preset: 'near-operation-file',
            presetConfig: {
                baseTypesPath: 'lib/graphql-types',
                extension: '.gql.ts',
            },
            config: {
                useTypeImports: true,
                documentVariableSuffix: '', // The default is 'Document'
            },
            plugins: ['typescript-operations', 'typed-document-node'],
        },
    },
    schema: '../../cmd/frontend/graphqlbackend/*.graphql',
    errorsOnly: true,
    config: {
        // https://the-guild.dev/graphql/codegen/plugins/typescript/typescript-operations#config-api-reference
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
}

export function codegen(): Promise<void> {
    return generate(codgegenConfig, true)
}

/**
 * Vite plugin for generating TypeScript types for GraphQL.
 *
 * We don't use the vite-plugin-graphql-codegen package because it doesn't support
 * watch mode when documents are defined separately for every generated file.
 */
export default function graphqlCodegen(): Plugin {
    async function codegen(): Promise<void> {
        try {
            await generate(codgegenConfig, true)
        } catch {
            // generate already logs errors to the console
            // but we still need to catch it otherwise vite
            // will terminate during watch mode
        }
    }

    // Cheap custom function to check whether we should run codegen for the provided path
    // It will run the code generator more often than necessary (e.g. for .ts files that
    // don't contain GraphQL queries), but we avoid having to read the file contents
    // and codegen is fast enough.
    function shouldRunCodegen(path: string): boolean {
        // Do not run codegen for generated files
        if (/(graphql-(operations|types|type-mocks)|\.gql)\.ts$/.test(path)) {
            return false
        }
        return /\.(ts|gql)$/.test(path)
    }

    return {
        name: 'graphql-codegen',
        buildStart() {
            return codegen()
        },
        configureServer(server) {
            server.watcher.on('add', path => {
                if (shouldRunCodegen(path)) {
                    codegen()
                }
            })
            server.watcher.on('change', path => {
                if (shouldRunCodegen(path)) {
                    codegen()
                }
            })
        },
    }
}
