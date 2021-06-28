// @ts-check

const path = require('path')

const { generate } = require('@graphql-codegen/cli')

const ROOT_FOLDER = path.resolve(__dirname, '../../../')

const WEB_FOLDER = path.resolve(ROOT_FOLDER, './client/web')
const BROWSER_FOLDER = path.resolve(ROOT_FOLDER, './client/browser')
const SHARED_FOLDER = path.resolve(ROOT_FOLDER, './client/shared')
const SCHEMA_PATH = path.join(ROOT_FOLDER, './cmd/frontend/graphqlbackend/*.graphql')

const IGNORED_FILES = [`!${ROOT_FOLDER}/**/graphql-operations.ts`]

const SHARED_DOCUMENTS_GLOB = [
  `${SHARED_FOLDER}/src/**/*.{ts,tsx,graphql}`,
  `!${SHARED_FOLDER}/src/testing/**/*.*`,
  `!${SHARED_FOLDER}/src/graphql/schema.ts`,
  ...IGNORED_FILES,
]

const WEB_DOCUMENTS_GLOB = [
  `${WEB_FOLDER}/src/**/*.{ts,tsx,graphql}`,
  `!${WEB_FOLDER}/src/regression/**/*.*`,
  `!${WEB_FOLDER}/src/end-to-end/**/*.*`,
  ...IGNORED_FILES,
]

const BROWSER_DOCUMENTS_GLOB = [
  `${BROWSER_FOLDER}/src/**/*.{ts,tsx,graphql}`,
  `!${BROWSER_FOLDER}/src/end-to-end/**/*.*`,
  '!**/*.d.ts',
  ...IGNORED_FILES,
]

// Define ALL_DOCUMENTS_GLOB as the union of the previous glob arrays.
const ALL_DOCUMENTS_GLOB = [...new Set([...SHARED_DOCUMENTS_GLOB, ...WEB_DOCUMENTS_GLOB, ...BROWSER_DOCUMENTS_GLOB])]

const plugins = [`${SHARED_FOLDER}/dev/extractGraphQlOperationCodegenPlugin.js`, 'typescript', 'typescript-operations']

/**
 * Generates TypeScript files with types for all GraphQL operations.
 */
async function generateGraphQlOperations() {
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
        skipTypename: true,
        // Appends /*#__PURE__*/ before each export. This means Webpack can tree-shake out any generated code that we do not use.
        pureMagicComment: true,
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
          plugins,
        },

        [path.join(WEB_FOLDER, './src/graphql-operations.ts')]: {
          documents: WEB_DOCUMENTS_GLOB,
          config: {
            onlyOperationTypes: true,
            noExport: false,
            enumValues: '@sourcegraph/shared/src/graphql-operations',
            interfaceNameForOperations: 'WebGraphQlOperations',
          },
          plugins: [...plugins, 'typescript-react-apollo'],
        },

        [path.join(SHARED_FOLDER, './src/graphql-operations.ts')]: {
          documents: SHARED_DOCUMENTS_GLOB,
          config: {
            onlyOperationTypes: true,
            noExport: false,
            interfaceNameForOperations: 'SharedGraphQlOperations',
          },
          plugins,
        },
      },
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
