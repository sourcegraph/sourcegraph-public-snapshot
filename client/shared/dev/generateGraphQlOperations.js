// @ts-check

const path = require('path')

const { generate } = require('@graphql-codegen/cli')

const ROOT_FOLDER = path.resolve(__dirname, '../../../')

const WEB_FOLDER = path.resolve(ROOT_FOLDER, './client/web')
const BROWSER_FOLDER = path.resolve(ROOT_FOLDER, './client/browser')
const SHARED_FOLDER = path.resolve(ROOT_FOLDER, './client/shared')
const SEARCH_FOLDER = path.resolve(ROOT_FOLDER, './client/search')
const VSCODE_FOLDER = path.resolve(ROOT_FOLDER, './client/vscode')
const JETBRAINS_FOLDER = path.resolve(ROOT_FOLDER, './client/jetbrains')
const SCHEMA_PATH = path.join(ROOT_FOLDER, './cmd/frontend/graphqlbackend/*.graphql')

const SHARED_DOCUMENTS_GLOB = [
  `${SHARED_FOLDER}/src/**/*.{ts,tsx}`,
  `!${SHARED_FOLDER}/src/testing/**/*.*`,
  `!${SHARED_FOLDER}/src/schema.ts`,
]

const WEB_DOCUMENTS_GLOB = [
  `${WEB_FOLDER}/src/**/*.{ts,tsx}`,
  `!${WEB_FOLDER}/src/regression/**/*.*`,
  `!${WEB_FOLDER}/src/end-to-end/**/*.*`,
]

const BROWSER_DOCUMENTS_GLOB = [
  `${BROWSER_FOLDER}/src/**/*.{ts,tsx}`,
  `!${BROWSER_FOLDER}/src/end-to-end/**/*.*`,
  '!**/*.d.ts',
]

const SEARCH_DOCUMENTS_GLOB = [`${SEARCH_FOLDER}/src/**/*.{ts,tsx}`]

const VSCODE_DOCUMENTS_GLOB = [`${VSCODE_FOLDER}/src/**/*.{ts,tsx}`]

const JETBRAINS_DOCUMENTS_GLOB = [`${JETBRAINS_FOLDER}/webview/src/**/*.{ts,tsx}`]

// Define ALL_DOCUMENTS_GLOB as the union of the previous glob arrays.
const ALL_DOCUMENTS_GLOB = [
  ...new Set([
    ...SHARED_DOCUMENTS_GLOB,
    ...WEB_DOCUMENTS_GLOB,
    ...BROWSER_DOCUMENTS_GLOB,
    ...SEARCH_DOCUMENTS_GLOB,
    ...VSCODE_DOCUMENTS_GLOB,
    ...JETBRAINS_DOCUMENTS_GLOB,
  ]),
]

const SHARED_PLUGINS = [
  `${SHARED_FOLDER}/dev/extractGraphQlOperationCodegenPlugin.js`,
  'typescript',
  'typescript-operations',
]

/**
 * Generates TypeScript files with types for all GraphQL operations.
 */
async function generateGraphQlOperations() {
  try {
    await generate(
      {
        schema: SCHEMA_PATH,
        hooks: {
          afterOneFileWrite: 'prettier --write',
        },
        errorsOnly: true,
        config: {
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
        generates: {
          [path.join(BROWSER_FOLDER, './src/graphql-operations.ts')]: {
            documents: BROWSER_DOCUMENTS_GLOB,
            config: {
              onlyOperationTypes: true,
              noExport: false,
              enumValues: '@sourcegraph/shared/src/graphql-operations',
              interfaceNameForOperations: 'BrowserGraphQlOperations',
            },
            plugins: SHARED_PLUGINS,
          },

          [path.join(WEB_FOLDER, './src/graphql-operations.ts')]: {
            documents: WEB_DOCUMENTS_GLOB,
            config: {
              onlyOperationTypes: true,
              noExport: false,
              enumValues: '@sourcegraph/shared/src/graphql-operations',
              interfaceNameForOperations: 'WebGraphQlOperations',
            },
            plugins: SHARED_PLUGINS,
          },

          [path.join(SHARED_FOLDER, './src/graphql-operations.ts')]: {
            documents: SHARED_DOCUMENTS_GLOB,
            config: {
              onlyOperationTypes: true,
              noExport: false,
              interfaceNameForOperations: 'SharedGraphQlOperations',
            },
            plugins: [...SHARED_PLUGINS, 'typescript-apollo-client-helpers'],
          },

          [path.join(SEARCH_FOLDER, './src/graphql-operations.ts')]: {
            documents: SEARCH_DOCUMENTS_GLOB,
            config: {
              onlyOperationTypes: true,
              noExport: false,
              enumValues: '@sourcegraph/shared/src/graphql-operations',
              interfaceNameForOperations: 'SearchGraphQlOperations',
            },
            plugins: SHARED_PLUGINS,
          },

          [path.join(VSCODE_FOLDER, './src/graphql-operations.ts')]: {
            documents: VSCODE_DOCUMENTS_GLOB,
            config: {
              onlyOperationTypes: true,
              noExport: false,
              enumValues: '@sourcegraph/shared/src/graphql-operations',
              interfaceNameForOperations: 'VSCodeGraphQlOperations',
            },
            plugins: SHARED_PLUGINS,
          },

          [path.join(JETBRAINS_FOLDER, './webview/src/graphql-operations.ts')]: {
            documents: JETBRAINS_DOCUMENTS_GLOB,
            config: {
              onlyOperationTypes: true,
              noExport: false,
              enumValues: '@sourcegraph/shared/src/graphql-operations',
              interfaceNameForOperations: 'JetBrainsGraphQlOperations',
            },
            plugins: SHARED_PLUGINS,
          },
        },
      },
      true
    )
  } catch (error) {
    console.log(error)
  }
}

module.exports = {
  generateGraphQlOperations,
  SHARED_DOCUMENTS_GLOB,
  WEB_DOCUMENTS_GLOB,
  ALL_DOCUMENTS_GLOB,
}
