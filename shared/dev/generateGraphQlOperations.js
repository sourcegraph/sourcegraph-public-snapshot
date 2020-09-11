// @ts-check

const { generate } = require('@graphql-codegen/cli')
const path = require('path')

const ROOT_FOLDER = path.resolve(__dirname, '../../')

const WEB_FOLDER = path.resolve(ROOT_FOLDER, './web')
const BROWSER_FOLDER = path.resolve(ROOT_FOLDER, './browser')
const SHARED_FOLDER = path.resolve(ROOT_FOLDER, './shared')
const SCHEMA_PATH = path.join(ROOT_FOLDER, './cmd/frontend/graphqlbackend/schema.graphql')

const SHARED_DOCUMENTS_GLOB = [`${SHARED_FOLDER}/src/**/*.{ts,tsx}`, `!${SHARED_FOLDER}/src/testing/**/*.*`]

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

const plugins = [`${SHARED_FOLDER}/dev/extractGraphQlOperationCodegenPlugin.js`, 'typescript', 'typescript-operations']

/**
 * Generates TypeScript files with types for all GraphQL operations.
 *
 * @param {{ watch?: boolean }} [options]
 */
async function generateGraphQlOperations({ watch } = {}) {
  await generate(
    {
      watch,
      schema: SCHEMA_PATH,
      hooks: {
        afterOneFileWrite: 'prettier --write',
      },
      config: {
        preResolveTypes: true,
        operationResultSuffix: 'Result',
        omitOperationSuffix: true,
        skipTypename: true,
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
        },
      },
      generates: {
        [path.join(ROOT_FOLDER, './browser/src/graphql-operations.ts')]: {
          documents: BROWSER_DOCUMENTS_GLOB,
          config: {
            onlyOperationTypes: true,
            noExport: false,
            enumValues: '../../shared/src/graphql/schema',
            interfaceNameForOperations: 'BrowserGraphQlOperations',
          },
          plugins,
        },

        [path.join(ROOT_FOLDER, './web/src/graphql-operations.ts')]: {
          documents: WEB_DOCUMENTS_GLOB,
          config: {
            onlyOperationTypes: true,
            noExport: false,
            enumValues: '../../shared/src/graphql/schema',
            interfaceNameForOperations: 'WebGraphQlOperations',
          },
          plugins,
        },

        [path.join(ROOT_FOLDER, './shared/src/graphql-operations.ts')]: {
          documents: SHARED_DOCUMENTS_GLOB,
          config: {
            onlyOperationTypes: true,
            noExport: false,
            enumValues: './graphql/schema',
            interfaceNameForOperations: 'SharedGraphQlOperations',
          },
          plugins,
        },
      },
    },
    true
  )
}

module.exports = { generateGraphQlOperations, SHARED_DOCUMENTS_GLOB, WEB_DOCUMENTS_GLOB }
