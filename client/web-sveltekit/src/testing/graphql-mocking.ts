import { faker } from '@faker-js/faker'
import {
    graphqlSync,
    isEnumType,
    isInterfaceType,
    isListType,
    isNonNullType,
    isObjectType,
    isScalarType,
    type ExecutionResult,
    type GraphQLFieldResolver,
    type GraphQLObjectType,
    type GraphQLResolveInfo,
    type GraphQLScalarType,
    type GraphQLSchema,
    type GraphQLType,
    type InlineFragmentNode,
    type GraphQLNullableType,
    GraphQLList,
    isUnionType,
} from 'graphql'

function isObject(value: unknown): value is { [key: string]: unknown } {
    return typeof value === 'object' && value !== null
}

function isObjectWithTypename(value: unknown): value is { __typename: string; [key: string]: unknown } {
    return isObject(value) && '__typename' in value
}

interface ObjectMock {
    [key: string]: unknown
}

interface Fixture extends ObjectMock {
    __typename: string
}

export interface Mocks {
    [typeName: string]: (info: GraphQLResolveInfo) => any
}

export interface TypePolicy {
    keyField?: string
}

export interface GraphQLMockServerOptions {
    /**
     * The GraphQL schema to mock.
     */
    schema: GraphQLSchema
    /**
     * A map of mock functions for each type in the schema.
     */
    mocks?: Mocks
    /**
     * A map of mock functions for each type in the schema.
     */
    fixtures?: Fixture[]
    /**
     * A seed for the random value generator.
     */
    seed?: number
    /**
     * The minimum length of lists.
     * @default 1
     */
    minListLength?: number
    /**
     * The maximum length of lists.
     * @default 1
     */
    maxListLength?: number
    /**
     * The probability of a nullable field returning null (between 0 and 1).
     * @default 0.3
     */
    nullProbability?: number
    typePolicies?: Record<string, TypePolicy>
}

interface OperationMock {
    operationName?: string
    mocks: Mocks
}

interface GraphQLMockServerContext {}

function isInlinedFragment(node: any): node is InlineFragmentNode {
    return !!node && node.kind === 'InlineFragment'
}

const defaultResolvers: Record<string, Record<string, GraphQLFieldResolver<any, GraphQLMockServerContext>>> = {
    Query: {
        node(_value, { id }, _context, info) {
            // Try to determine type name from inline fragment
            const typename =
                info.fieldNodes[0].selectionSet?.selections.find(isInlinedFragment)?.typeCondition?.name.value
            const type = typename && info.schema.getType(typename)
            if (type) {
                return {
                    __typename: typename,
                    id: id,
                }
            }
            throw new Error(
                'Mock error: Unable to determine typename for node query. Please use an inline fragment in the query.'
            )
        },
    },
}

export class GraphQLMockServer {
    private operationMocks: OperationMock[] = []
    private objectStore: Map<string, ObjectMock> = new Map()

    constructor(private readonly options: GraphQLMockServerOptions) {
        if (options.fixtures) {
            this.addFixtures(options.fixtures)
        }
    }

    public addFixtures(fixtures: Fixture[]): void {
        for (const fixture of fixtures) {
            const type = this.options.schema.getType(fixture.__typename)
            if (!type) {
                throw new Error(`Unknown type ${fixture.__typename}`)
            }
            if (!isObjectType(type)) {
                throw new Error(`Type ${fixture.__typename} is not an object type`)
            }
            const keyFieldName = this.options.typePolicies?.[type.name]?.keyField ?? 'id'
            const keyField = type.getFields()[keyFieldName]
            if (!keyField) {
                throw new Error(`Type ${fixture.__typename} does not have a key field`)
            }
            if (!(keyFieldName in fixture)) {
                throw new Error(`Fixture for type ${fixture.__typename} requires a value for key field ${keyFieldName}`)
            }
            this.objectStore.set(`${type.name}:${fixture[keyFieldName]}`, fixture)
        }
    }

    public addMocks(mocks: Mocks, operationName?: string): void {
        this.operationMocks.push({
            operationName,
            mocks,
        })
    }

    public reset(): void {
        this.operationMocks.length = 0
    }

    public query(query: string, variables?: Record<string, unknown>, operationName?: string): ExecutionResult {
        faker.seed(this.options.seed ?? 1)

        const result = graphqlSync({
            schema: this.options.schema,
            source: query,
            variableValues: variables,
            operationName,
            contextValue: {
                getMockValue: this.getMockValue.bind(this),
            },
            fieldResolver: this.fieldResolver,
        })
        return result
    }

    // By using a default field resolver to generate mock data we avoid having to
    // extend the provided schema, which can be expensive for large schemas.
    private fieldResolver: GraphQLFieldResolver<any, GraphQLMockServerContext> = (
        obj,
        variables,
        context,
        info
    ): unknown => {
        if (!obj) {
            // Must be a root query. We resolve the query here to make sure we
            // resolve operation-specific overrides correctly.
            obj = this.getMockValue(info.parentType, info)
        }
        if (isObjectType(info.parentType)) {
            // Restore any previously resolved object with the same ID
            //obj = this.resolveObject(info.parentType, obj, info)
        }

        // If the object already has a value for this field, return it.
        // This will be the case for fields that are explicitly mocked.
        if (obj && info.fieldName in obj) {
            return this.getMockValue(info.returnType, info, obj[info.fieldName])
        }

        if (defaultResolvers[info.parentType.name]?.[info.fieldName]) {
            return defaultResolvers[info.parentType.name][info.fieldName](obj, variables, context, info)
        }

        return this.getMockValue(info.returnType, info)
    }

    private resolveObject(
        type: GraphQLObjectType,
        value: ObjectMock | undefined,
        info: GraphQLResolveInfo
    ): ObjectMock {
        if (!value) {
            value = {}
        }

        const keyField = this.options.typePolicies?.[type.name]?.keyField ?? 'id'
        let idFieldType = type.getFields()[keyField]?.type
        if (isNonNullType(idFieldType)) {
            idFieldType = idFieldType.ofType
        }
        if (!idFieldType || !isScalarType(idFieldType)) {
            return value
        }
        const id = value[keyField] ?? this.getScalarMockValue(idFieldType, info)
        const cacheKey = `${type.name}:${id}`

        const obj = this.objectStore.get(cacheKey)
        return obj ? { ...obj, ...value } : { ...value }
    }

    /**
     * Yields mock backends in order of precedence (low to high).
     * Default mocks have the lowest precedence, then unamed dynamic mocks, then operation-specific mocks.
     */
    private *backends(
        type: Exclude<GraphQLNullableType, GraphQLList<any>>,
        info: GraphQLResolveInfo
    ): Generator<Mocks> {
        if (this.options.mocks && this.options.mocks[type.name]) {
            yield this.options.mocks
        }

        for (const { operationName, mocks } of this.operationMocks) {
            if (!operationName && mocks[type.name]) {
                yield mocks
            }
        }

        for (const { operationName, mocks } of this.operationMocks) {
            if (operationName && operationName === info.operation.name?.value && mocks[type.name]) {
                yield mocks
            }
        }
    }

    private getMockValue(type: GraphQLType, info: GraphQLResolveInfo, override?: unknown): unknown {
        if (isNonNullType(type)) {
            type = type.ofType
        } else {
            // Return null in ~30% of cases
            if (
                override === null ||
                (override === undefined && faker.number.float() < (this.options.nullProbability ?? 0.3))
            ) {
                return null
            }
        }
        if (isListType(type)) {
            type = type.ofType

            const list = (
                Array.isArray(override)
                    ? override
                    : Array.from({
                          length: faker.number.int({
                              min: this.options.minListLength ?? 1,
                              max: this.options.maxListLength ?? 1,
                          }),
                      })
            ).map((value: unknown) => this.getMockValue(type, info, value))
            return list
        }

        if (isInterfaceType(type)) {
            const objType = isObjectWithTypename(override)
                ? info.schema.getType(override.__typename)
                : faker.helpers.arrayElement(info.schema.getImplementations(type).objects)
            if (!isObjectType(objType)) {
                throw new Error(`Unable to determine object type for interface ${type.name}`)
            }
            const mockValue = this.getObjectMockValue(objType, info, isObject(override) ? override : undefined)
            // Needs to be explicitly set for interface types
            mockValue.__typename = objType.name
            return mockValue
        }

        if (isUnionType(type)) {
            const objType = isObjectWithTypename(override)
                ? info.schema.getType(override.__typename)
                : faker.helpers.arrayElement(type.getTypes())
            if (!isObjectType(objType)) {
                throw new Error(`Unable to determine object type for union ${type.name}`)
            }
            const mockValue = this.getObjectMockValue(objType, info, isObject(override) ? override : undefined)
            // Needs to be explicitly set for union types
            mockValue.__typename = objType.name
            return mockValue
        }

        if (isObjectType(type)) {
            return this.getObjectMockValue(type, info, isObject(override) ? override : undefined)
        }

        if (override !== undefined) {
            return override
        }

        if (isEnumType(type)) {
            return faker.helpers.arrayElement(type.getValues()).value
        }

        if (isScalarType(type)) {
            return this.getScalarMockValue(type, info)
        }

        throw new Error(`Unsupported return type ${type}`)
    }

    private getObjectMockValue(type: GraphQLObjectType, info: GraphQLResolveInfo, override?: ObjectMock): ObjectMock {
        let mockValue: ObjectMock = this.resolveObject(type, override, info)
        const interfaces = type.getInterfaces()
        for (const interfaceType of interfaces) {
            for (const backend of this.backends(interfaceType, info)) {
                Object.assign(mockValue, backend[interfaceType.name](info))
            }
        }
        for (const backend of this.backends(type, info)) {
            Object.assign(mockValue, backend[type.name](info))
        }
        return { ...mockValue, ...override }
    }

    private getScalarMockValue(type: GraphQLScalarType, info: GraphQLResolveInfo): unknown {
        let mockValue: unknown = null
        for (const mocks of this.backends(type, info)) {
            mockValue = mocks[type.name](info)
        }
        if (mockValue !== null) {
            return mockValue
        }

        switch (type.name) {
            case 'String':
                return faker.string.alpha(10)
            case 'Int':
                // Ints are 32-bit signed integers
                return faker.number.int({ max: 2 ** 31 - 1 })
            case 'Float':
                return faker.number.float()
            case 'Boolean':
                return faker.number.float() < 0.5
            case 'ID':
                return faker.string.uuid()
            default:
                throw new Error(`Unknown scalar type ${type.name}`)
        }
    }
}
