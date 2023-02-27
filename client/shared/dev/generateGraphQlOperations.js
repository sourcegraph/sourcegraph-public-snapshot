// @ts-check

const path = require('path')

const { generate } = require('@graphql-codegen/cli')

const ROOT_FOLDER = path.resolve(__dirname, '../../../')

const WEB_FOLDER = path.resolve(ROOT_FOLDER, './client/web')
const SVELTEKIT_FOLDER = path.resolve(ROOT_FOLDER, './client/web-sveltekit')
const BROWSER_FOLDER = path.resolve(ROOT_FOLDER, './client/browser')
const SHARED_FOLDER = path.resolve(ROOT_FOLDER, './client/shared')
const VSCODE_FOLDER = path.resolve(ROOT_FOLDER, './client/vscode')
const JETBRAINS_FOLDER = path.resolve(ROOT_FOLDER, './client/jetbrains')
const SCHEMA_PATH = path.join(ROOT_FOLDER, './cmd/frontend/graphqlbackend/*.graphql')

const SHARED_DOCUMENTS_GLOB = [`${SHARED_FOLDER}/src/**/*.{ts,tsx}`]

const WEB_DOCUMENTS_GLOB = [
  `${WEB_FOLDER}/src/**/*.{ts,tsx}`,
  `${WEB_FOLDER}/src/regression/**/*.*`,
  `!${WEB_FOLDER}/src/end-to-end/**/*.*`,
]

const SVELTEKIT_DOCUMENTS_GLOB = [`${SVELTEKIT_FOLDER}/src/lib/**/*.ts`]

const BROWSER_DOCUMENTS_GLOB = [
  `${BROWSER_FOLDER}/src/**/*.{ts,tsx}`,
  `!${BROWSER_FOLDER}/src/end-to-end/**/*.*`,
  '!**/*.d.ts',
]

const VSCODE_DOCUMENTS_GLOB = [`${VSCODE_FOLDER}/src/**/*.{ts,tsx}`]

const JETBRAINS_DOCUMENTS_GLOB = [`${JETBRAINS_FOLDER}/webview/src/**/*.{ts,tsx}`]

const GLOBS = {
  BrowserGraphQlOperations: BROWSER_DOCUMENTS_GLOB,
  JetBrainsGraphQlOperations: JETBRAINS_DOCUMENTS_GLOB,
  SharedGraphQlOperations: SHARED_DOCUMENTS_GLOB,
  VSCodeGraphQlOperations: VSCODE_DOCUMENTS_GLOB,
  WebGraphQlOperations: WEB_DOCUMENTS_GLOB,
  SvelteKitGraphQlOperations: SVELTEKIT_DOCUMENTS_GLOB,
}

const EXTRA_PLUGINS = {
  SharedGraphQlOperations: ['typescript-apollo-client-helpers'],
}

// Define ALL_DOCUMENTS_GLOB as the union of the previous glob arrays.
const ALL_DOCUMENTS_GLOB = [
  ...new Set([
    ...SHARED_DOCUMENTS_GLOB,
    ...WEB_DOCUMENTS_GLOB,
    ...SVELTEKIT_DOCUMENTS_GLOB,
    ...BROWSER_DOCUMENTS_GLOB,
    ...VSCODE_DOCUMENTS_GLOB,
    ...JETBRAINS_DOCUMENTS_GLOB,
  ]),
]

const SHARED_PLUGINS = [
  `${SHARED_FOLDER}/dev/extractGraphQlOperationCodegenPlugin.js`,
  'typescript',
  'typescript-operations',
]

const PRETTIER = path.join(path.dirname(require.resolve('prettier')), 'bin-prettier.js')

/**
 * Generates TypeScript files with types for all GraphQL operations.
 */
async function generateGraphQlOperations() {
  try {
    await _generateGraphQlOperations([
      {
        interfaceNameForOperations: 'BrowserGraphQlOperations',
        outputPath: path.join(BROWSER_FOLDER, './src/graphql-operations.ts'),
      },
      {
        interfaceNameForOperations: 'WebGraphQlOperations',
        outputPath: path.join(WEB_FOLDER, './src/graphql-operations.ts'),
      },
      {
        interfaceNameForOperations: 'SvelteKitGraphQlOperations',
        outputPath: path.join(SVELTEKIT_FOLDER, './src/lib/graphql-operations.ts'),
      },
      {
        interfaceNameForOperations: 'SharedGraphQlOperations',
        outputPath: path.join(SHARED_FOLDER, './src/graphql-operations.ts'),
      },
      {
        interfaceNameForOperations: 'VSCodeGraphQlOperations',
        outputPath: path.join(VSCODE_FOLDER, './src/graphql-operations.ts'),
      },
      {
        interfaceNameForOperations: 'JetBrainsGraphQlOperations',
        outputPath: path.join(JETBRAINS_FOLDER, './webview/src/graphql-operations.ts'),
      },
    ])
  } catch (error) {
    console.log(error)
  }
}

async function _generateGraphQlOperations(operations) {
  const generates = operations.reduce((generates, operation) => {
    generates[operation.outputPath] = {
      documents: GLOBS[operation.interfaceNameForOperations],
      config: {
        onlyOperationTypes: true,
        noExport: false,
        enumValues:
          operation.interfaceNameForOperations === 'SharedGraphQlOperations'
            ? undefined
            : '@sourcegraph/shared/src/graphql-operations',
        interfaceNameForOperations: operation.interfaceNameForOperations,
      },
      plugins: [...SHARED_PLUGINS, ...(EXTRA_PLUGINS[operation.interfaceNameForOperations] || [])],
    }
    return generates
  }, {})

  await generate(
    {
      schema: SCHEMA_PATH,
      hooks: {
        afterOneFileWrite: `${PRETTIER} --write`,
      },
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
      generates,
    },
    true
  )
}

module.exports = {
  generateGraphQlOperations,
  SHARED_DOCUMENTS_GLOB,
  WEB_DOCUMENTS_GLOB,
  ALL_DOCUMENTS_GLOB,
}

// Bazel entry point to generate a single graphql operations file; the legacy build
// continues to import `generateGraphQlOperations` and generate all operations files.
async function main(args) {
  if (args.length !== 2) {
    throw new Error('Usage: <schemaName> <outputPath>')
  }

  const [interfaceNameForOperations, outputPath] = args

  await _generateGraphQlOperations([{ interfaceNameForOperations, outputPath }])
}

if (require.main === module) {
  main(process.argv.slice(2)).catch(error => {
    console.error(error)
    process.exit(1)
  })
}
