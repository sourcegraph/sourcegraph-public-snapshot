// @ts-check

const { generate } = require('@graphql-codegen/cli');
const path = require('path');

const ROOT_FOLDER = path.resolve(__dirname, '../../')

const WEB_FOLDER = path.resolve(ROOT_FOLDER, './web')
const SHARED_FOLDER = path.resolve(ROOT_FOLDER, './shared')

const plugins = [
  `${SHARED_FOLDER}/dev/extractGraphQlOperationCodegenPlugin.js`,
  'typescript',
  'typescript-operations',
];

async function generateGraphQlOperations(watch = false) {
   await generate(
    {
      schema: path.join(ROOT_FOLDER, './cmd/frontend/graphqlbackend/schema.graphql'),
      watch,
      hooks: { afterOneFileWrite: 'prettier --write' },
      config: {
        onlyOperationTypes: true,
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
          documents: [
            `${WEB_FOLDER}/src/**/*.{ts,tsx}`,
            `!${WEB_FOLDER}/src/regression/**/*.*`,
            `!${WEB_FOLDER}/src/end-to-end/**/*.*`,
          ],
          config: {
            enumValues: '../../shared/src/graphql/schema',
            interfaceNameForOperations: 'WebGraphQlOperations',
          },
          plugins,
        },

        [path.join(ROOT_FOLDER, './shared/src/graphql-operations.ts')]: {
          documents: [
           `${SHARED_FOLDER}/src/**/*.{ts,tsx}`,
           `!${SHARED_FOLDER}/src/testing/**/*.*`,
          ],
          config: {
            enumValues: './graphql/schema',
            interfaceNameForOperations: 'SharedGraphQlOperations'
          },
          plugins,
        },
      },
    },
    true
  );
}

module.exports = { generateGraphQlOperations }
