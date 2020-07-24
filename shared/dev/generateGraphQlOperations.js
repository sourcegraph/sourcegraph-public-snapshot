// @ts-check

const { generate } = require('@graphql-codegen/cli');
const path = require('path');

const ROOT_FOLDER = path.resolve(__dirname, '../../')

const WEB_FOLDER = path.resolve(ROOT_FOLDER, './web')
const SHARED_FOLDER = path.resolve(ROOT_FOLDER, './shared')
const SCHEMA_PATH = path.join(ROOT_FOLDER, './cmd/frontend/graphqlbackend/schema.graphql');

const SHARED_DOCUMENTS_GLOB = [
  `${SHARED_FOLDER}/src/**/*.{ts,tsx}`,
  `!${SHARED_FOLDER}/src/testing/**/*.*`,
]

const WEB_DOCUMENTS_GLOB = [
  `${WEB_FOLDER}/src/**/*.{ts,tsx}`,
  `!${WEB_FOLDER}/src/regression/**/*.*`,
  `!${WEB_FOLDER}/src/end-to-end/**/*.*`,
]

const plugins = [
  `${SHARED_FOLDER}/dev/extractGraphQlOperationCodegenPlugin.js`,
  'typescript',
  'typescript-operations',
];

async function generateGraphQlOperations() {
   await generate(
    {
      schema: SCHEMA_PATH,
      hooks: { afterOneFileWrite: 'prettier --write' },
      config: {
        preResolveTypes: true,
        operationResultSuffix: 'Result',
        omitOperationSuffix: true,
        skipTypename: true,
        namingConvention: 'keep',
        declarationKind: 'interface',
        avoidOptionals: true,
        scalars: {
          DateTime: 'string',
          JSON: 'object',
          JSONValue: 'unknown',
          GitObjectID: 'string',
          JSONCString: 'string',
        },
      },
      generates: {
        [path.join(ROOT_FOLDER, './web/src/graphql-operations.ts')]: {
          documents: WEB_DOCUMENTS_GLOB,
          config: {
            onlyOperationTypes: true,
            noExport: true,
            enumValues: '../../shared/src/graphql/schema',
            interfaceNameForOperations: 'WebGraphQlOperations',
          },
          plugins,
        },

        [path.join(ROOT_FOLDER, './shared/src/graphql-operations.ts')]: {
          documents: SHARED_DOCUMENTS_GLOB,
          config: {
            onlyOperationTypes: true,
            noExport: true,
            enumValues: './graphql/schema',
            interfaceNameForOperations: 'SharedGraphQlOperations'
          },
          plugins,
        },

        [path.join(ROOT_FOLDER, './shared/src/graphql/schema.ts')]: {
          config: {
            namingConvention: 'pascalCase',
            skipTypename: false,
            nonOptionalTypename: true,
            avoidOptionals: false,
            typesPrefix: 'I',
            enumPrefix: false
          },
          plugins: ['typescript']
        },
      },
    },
    true
  );
}
// union no I
// capital cases for types
// no optionals?
module.exports = { generateGraphQlOperations, SHARED_DOCUMENTS_GLOB, WEB_DOCUMENTS_GLOB }
