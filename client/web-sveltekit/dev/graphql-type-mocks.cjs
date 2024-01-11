// @ts-check
const { isObjectType, isEnumType, isInputType } = require('graphql')
const { isScalarType } = require('graphql')
const logger = require('signale')

/**
 *
 * @param {import('graphql').GraphQLSchema} schema
 * @param {import('@graphql-codegen/plugin-helpers').Types.DocumentFile[]} _documents
 * @param {{typesImport: string, mockInterfaceName?: string}} config
 */
const plugin = (schema, _documents, config) => {
    const { mockInterfaceName = 'TypeMocks', typesImport } = config
    schema.getTypeMap()

    const interfaceFields = Object.values(schema.getTypeMap())
        .filter(value => !value.name.startsWith('__') && !isEnumType(value) && !isInputType(value))
        .map(
            value =>
                `${value.name}?: (info: GraphQLResolveInfo) => ${
                    isScalarType(value) ? `Types.Scalars['${value.name}']['output']` : `${value.name}Mock`
                }\n`
        )
    const objectTypes = Object.values(schema.getTypeMap()).filter(
        value => !value.name.startsWith('__') && isObjectType(value) && !isInputType(value)
    )

    if (interfaceFields.length === 0) {
        logger.warn('No types found to generate interface ' + mockInterfaceName)
    }

    return [
        `import type { GraphQLResolveInfo } from 'graphql'`,
        `import type * as Types from '${typesImport}'`,
        'type DeepPartial<T> = T extends object ? {',
        '    [P in keyof T]?: DeepPartial<T[P]>',
        '} : T',
        ...objectTypes.map(type => `export type ${type.name}Mock = DeepPartial<Types.${type.name}>`),
        `export interface ${mockInterfaceName} extends Record<string, ((info: GraphQLResolveInfo) => any)|undefined> {`,
        `    ${interfaceFields.join('    ')}`,
        '}',
        'export type ObjectMock = ' + objectTypes.map(type => `${type.name}Mock`).join(' | '),
    ].join('\n')
}
module.exports = { plugin }
