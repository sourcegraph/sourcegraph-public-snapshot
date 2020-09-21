// @ts-check

const { generate } = require('@graphql-codegen/cli')
const path = require('path')
const isInputNewer = require('./isInputNewer')

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
  const allGenerateOperations = {
    './browser/src/graphql-operations.ts': {
      documents: BROWSER_DOCUMENTS_GLOB,
      config: {
        onlyOperationTypes: true,
        noExport: false,
        enumValues: '../../shared/src/graphql-operations',
        interfaceNameForOperations: 'BrowserGraphQlOperations',
      },
      plugins,
    },
    './web/src/graphql-operations.ts': {
      documents: WEB_DOCUMENTS_GLOB,
      config: {
        onlyOperationTypes: true,
        noExport: false,
        enumValues: '../../shared/src/graphql-operations',
        interfaceNameForOperations: 'WebGraphQlOperations',
      },
      plugins,
    },
    './shared/src/graphql-operations.ts': {
      documents: SHARED_DOCUMENTS_GLOB,
      config: {
        onlyOperationTypes: true,
        noExport: false,
        interfaceNameForOperations: 'SharedGraphQlOperations',
      },
      plugins,
    },
  }
  const generateOperations = {}
  for (const outfile of Object.keys(allGenerateOperations)) {
    const inputs = allGenerateOperations[outfile].documents
    if (await isInputNewer(inputs, outfile)) {
      generateOperations[path.join(ROOT_FOLDER, outfile)] = allGenerateOperations[outfile]
    } else {
      console.log(`skipping generation of ${outfile}, because all inputs were older`)
    }
  }
  if (Object.keys(generateOperations).length === 0) {
    return
  }

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
      generates: generateOperations,
    },
    true
  )
}

module.exports = { generateGraphQlOperations, SHARED_DOCUMENTS_GLOB, WEB_DOCUMENTS_GLOB }
