// @ts-check
const { isObjectType, visit, concatAST } = require('graphql')
const { isScalarType } = require('graphql')
const logger = require('signale')

/**
 * @param {import('@graphql-codegen/plugin-helpers').Types.DocumentFile[]} documents
 * @returns {import('graphql').ASTNode}
 */
function documentsToAST(documents) {
    const documentNodes = []
    for (const document of documents) {
        if (document.document) {
            documentNodes.push(document.document)
        }
    }
    return concatAST(documentNodes)
}

/**
 *
 * @param {import('graphql').GraphQLSchema} schema
 * @param {import('@graphql-codegen/plugin-helpers').Types.DocumentFile[]} documents
 * @param {{typesImport: string, mockInterfaceName?: string, operationResultSuffix?: string}} config
 * @returns {import('@graphql-codegen/plugin-helpers').Types.PluginOutput}
 */
const plugin = (schema, documents, config) => {
    const { mockInterfaceName = 'TypeMocks', typesImport, operationResultSuffix = '' } = config

    const interfaceFields = Object.values(schema.getTypeMap())
        .filter(value => !value.name.startsWith('__') && (isScalarType(value) || isObjectType(value)))
        .map(
            value =>
                `${value.name}?: MockFunction<${
                    isScalarType(value) ? `Scalars['${value.name}']['output']` : `DeepPartial<${value.name}Mock>`
                }>`
        )
    const objectTypes = Object.values(schema.getTypeMap()).filter(
        value => !value.name.startsWith('__') && isObjectType(value)
    )

    const operations = []
    visit(documentsToAST(documents), {
        OperationDefinition(node) {
            operations.push(node)
        },
    })

    if (interfaceFields.length === 0) {
        logger.warn('No types found to generate interface ' + mockInterfaceName)
    }

    return {
        prepend: [
            `import type { GraphQLResolveInfo } from 'graphql'`,
            `import type * as Types from '${typesImport}'`,
            'type DeepPartial<T> = T extends object ? {',
            '    [P in keyof T]?: DeepPartial<T[P]>',
            '} : T',
            'type MockFunction<Return> = (info: GraphQLResolveInfo) => Return',
        ],
        content: [
            ...objectTypes.map(type => `export type ${type.name}Mock = DeepPartial<Types.${type.name}>`),
            `export interface ${mockInterfaceName} {`,
            `    [key: string]: MockFunction<any>|undefined`,
            `    ${interfaceFields.join('\n    ')}`,
            '}',
            `export interface OperationMocks {`,
            `    [key: string]: ((variables: any) => DeepPartial<any>)|undefined`,
            `    ${operations
                .map(
                    operation =>
                        `${operation.name.value}?: (variables: ${operation.name.value}Variables) => DeepPartial<${operation.name.value}${operationResultSuffix}>`
                )
                .join('\n    ')}`,
            `}`,
            'export type ObjectMock = ' + objectTypes.map(type => `${type.name}Mock`).join(' | '),
        ].join('\n'),
    }
}
module.exports = { plugin }
