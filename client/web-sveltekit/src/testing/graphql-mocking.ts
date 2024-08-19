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

export interface TypeMocks {
    [typeName: string]: ((info: GraphQLResolveInfo) => any) | undefined
}

interface OperationMocks {
    [operationName: string]: ((variables: { [key: string]: any }) => any) | undefined
}

/**
 * Type specific configuration.
 */
export interface TypePolicy {
    /**
     * The field which uniquely identifies the object.
     */
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
    mocks?: TypeMocks
    /**
     * A list of mock fixtures. These can be used to describe a partial
     * database state. Operation mocks only have to reference the key field
     * of the fixture to includ it.
     */
    fixtures?: Fixture[]
    /**
     * A seed for the random value generator. The seed will be reset for each query.
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
    /**
     * Type specific configuration.
     */
    typePolicies?: Record<string, TypePolicy>
}

interface GraphQLMockServerContext {
    warnOnMissingOperationMocks?: boolean
}

interface QueryOptions {
    /**
     *  Logs GraphQL errors to the console.
     */
    logGraphQLErrors?: boolean
    /**
     * Logs a warning if the query doesn't have any operation mocks.
     */
    warnOnMissingOperationMocks?: boolean
}

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

/**
 * A mock GraphQL server that can be used to mock a GraphQL API.
 *
 * There are three ways to provide mock data (in order of precedence):
 *
 * 1. Provide a map of mock functions for each type in the schema. The mock functions will be called
 *    when the corresponding type is encountered in the query. It's possible to provide multiple
 *    mock functions for the same type, in which case they will be called in order of precedence and
 *    the result is the union of all mock values.
 *    This can be used to define sensible default values for certain types.
 * 2. Provide a list of fixtures. These fixtures are partially mocked objects with a key (e.g. 'id') field.
 *    Wherever a query returns an object with the same ID, the fixture will be used as the base object and
 *    the type mock functions will be called to fill in the rest of the fields.
 * 3. Provide a list of operation-specific mock functions. These mock functions will be called when the
 *    corresponding operation is encountered in the query. Any field set in the operation mock will take
 *    precedence over fixture values and type mock values.
 */
export class GraphQLMockServer {
    private typeMocks: TypeMocks[] = []
    private operationMocks: OperationMocks = {}
    private objectStore: Map<string, ObjectMock> = new Map()
    // List of known key fields. This is used to find fixtures by ID for query overrides that don't specify
    // the type of the object.
    private keyFields = new Set(['id'])

    constructor(private readonly options: GraphQLMockServerOptions) {
        if (options.fixtures) {
            this.addFixtures(options.fixtures)
        }

        if (options.typePolicies) {
            for (const policy of Object.values(options.typePolicies)) {
                if (policy.keyField) {
                    this.keyFields.add(policy.keyField)
                }
            }
        }
    }

    /**
     * Add a list of fixtures to the mock server. This allows to provide mock data for multiple objects of
     * the same type and reuse the same mock data in multiple queries.
     */
    public addFixtures(fixtures: Fixture[]): void {
        // While it's OK to overwrite fixtures for the same ID in separate calls, we want to
        // ensure that we don't have objects with duplicate IDs in the same list.
        const seenIDs = new Set<string>()
        for (const fixture of fixtures) {
            const type = this.options.schema.getType(fixture.__typename)
            if (!type) {
                throw new Error(`Unknown type '${fixture.__typename}'.`)
            }
            if (!isObjectType(type)) {
                throw new Error(`Type '${fixture.__typename}' is not an object type.`)
            }
            const keyFieldName = this.options.typePolicies?.[type.name]?.keyField ?? 'id'
            const keyField = type.getFields()[keyFieldName]
            if (!keyField) {
                throw new Error(`Type ${fixture.__typename} does not have a key field.`)
            }
            if (!(keyFieldName in fixture)) {
                throw new Error(
                    `Fixture for type '${fixture.__typename}' requires a value for key field '${keyFieldName}'.`
                )
            }
            const id = String(fixture[keyFieldName])
            if (seenIDs.has(id)) {
                throw new Error(`Fixture for type '${fixture.__typename}' has duplicate ID '${id}'.`)
            }
            this.objectStore.set(this.getCacheKey(type, id), fixture)
        }
    }

    /**
     * Additional default mocks for all types in the schema.
     */
    public addTypeMocks(mocks: TypeMocks): void {
        this.typeMocks.push(mocks)
    }

    /**
     * Mocks for specific operations. These mocks take precedence over type mocks and fixtures.
     */
    public addOperationMocks(operationMocks: OperationMocks): void {
        Object.assign(this.operationMocks, operationMocks)
    }

    /**
     * Remove all mocks added via `addTypeMocks` and `addOperationMocks`.
     */
    public reset(): void {
        this.typeMocks.length = 0
        this.operationMocks = {}
    }

    /**
     * Execute a GraphQL query and return the result.
     *
     * If `error` is set to `true`, the query will
     */
    public query(
        query: string,
        variables?: Record<string, unknown>,
        operationName?: string,
        options?: QueryOptions
    ): ExecutionResult {
        faker.seed(this.options.seed ?? 1)

        const result = graphqlSync({
            schema: this.options.schema,
            source: query,
            variableValues: variables,
            operationName,
            contextValue: {
                warnOnMissingOperationMocks: options?.warnOnMissingOperationMocks,
            },
            fieldResolver: this.fieldResolver,
        })
        if (options?.logGraphQLErrors && result.errors) {
            console.error(result.errors)
        }
        return result
    }

    // By using a default field resolver to generate mock data we avoid having to
    // extend the provided schema, which can be expensive for large schemas.
    private fieldResolver: GraphQLFieldResolver<any, GraphQLMockServerContext> = (
        obj,
        args,
        context,
        info
    ): unknown => {
        // Operation mocks might use field aliases, in which case we also need to
        // use the alias to look up the appropriat mock value
        const fieldName = info.fieldNodes[0].alias?.value ?? info.fieldName

        // Must be a root query. We resolve the query here to make sure we
        // resolve operation-specific overrides correctly.
        if (info.parentType.name === 'Query' || info.parentType.name === 'Mutation') {
            if (info.operation.name) {
                const operationMock = this.operationMocks[info.operation.name.value]
                if (operationMock) {
                    obj = operationMock(info.variableValues ?? {})
                } else if (context.warnOnMissingOperationMocks) {
                    console.warn(`No mock found for operation '${info.operation.name.value}'.`)
                }
            }
            obj = this.getMockValue(info.parentType, info, obj)
        }

        // If the object already has a value for this field, return it.
        // This will be the case for fields that are explicitly mocked.
        if (obj && fieldName in obj) {
            // Partial mock data is available and should take precedence over any other mock data
            return this.getMockValue(info.returnType, info, obj[fieldName])
        }

        if (defaultResolvers[info.parentType.name]?.[fieldName]) {
            return defaultResolvers[info.parentType.name][fieldName](obj, args, context, info)
        }

        return this.getMockValue(info.returnType, info)
    }

    private getMockValue(type: GraphQLType, info: GraphQLResolveInfo, override?: unknown): unknown {
        if (isNonNullType(type)) {
            type = type.ofType
        } else {
            // Return null in ~30% of cases by default
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
            const implementations = info.schema.getImplementations(type).objects
            const objType = this.getObjectType(override, implementations) ?? faker.helpers.arrayElement(implementations)
            if (!isObjectType(objType)) {
                throw new Error(`Unable to determine object type for interface ${type.name}`)
            }
            const mockValue = this.getObjectMockValue(objType, info, isObject(override) ? override : undefined)
            // Needs to be explicitly set for interface types
            mockValue.__typename = objType.name
            return mockValue
        }

        if (isUnionType(type)) {
            const types = type.getTypes()
            const objType = this.getObjectType(override, types) ?? faker.helpers.arrayElement(types)
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
        let mockValue: ObjectMock = this.resolveObject(type, override)
        const interfaces = type.getInterfaces()
        for (const interfaceType of interfaces) {
            for (const backend of this.backends(interfaceType)) {
                const mockFunction = backend[interfaceType.name]
                if (mockFunction) {
                    Object.assign(mockValue, mockFunction(info))
                }
            }
        }
        for (const backend of this.backends(type)) {
            const mockFunction = backend[type.name]
            if (mockFunction) {
                Object.assign(mockValue, mockFunction(info))
            }
        }
        return { ...mockValue, ...override }
    }

    private getScalarMockValue(type: GraphQLScalarType, info: GraphQLResolveInfo): unknown {
        let mockValue: unknown = null
        for (const mocks of this.backends(type)) {
            const mockFunction = mocks[type.name]
            if (mockFunction) {
                mockValue = mockFunction(info)
            }
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

    private getCacheKey(type: GraphQLObjectType, keyValue: string): string {
        return `${type.name}:${keyValue}`
    }

    private resolveObject(type: GraphQLObjectType, override: ObjectMock = {}): ObjectMock {
        const keyFieldName = this.options.typePolicies?.[type.name]?.keyField ?? 'id'
        let keyFieldType = type.getFields()[keyFieldName]?.type
        if (isNonNullType(keyFieldType)) {
            keyFieldType = keyFieldType.ofType
        }
        if (!keyFieldType || !isScalarType(keyFieldType)) {
            return { ...override }
        }
        const key = override[keyFieldName]
        const cacheKey = this.getCacheKey(type, String(key))

        const obj = this.objectStore.get(cacheKey)
        return obj ? { ...obj, ...override } : { ...override }
    }

    /**
     * Yields mock backends in order of precedence (low to high).
     * Default mocks have the lowest precedence, then unamed dynamic mocks, then operation-specific mocks.
     */
    private *backends(type: Exclude<GraphQLNullableType, GraphQLList<any>>): Generator<TypeMocks> {
        if (this.options.mocks && this.options.mocks[type.name]) {
            yield this.options.mocks
        }

        for (const mocks of this.typeMocks) {
            if (mocks[type.name]) {
                yield mocks
            }
        }
    }

    /**
     * Helper function to determine the type of a partially mocked object. Either the type is
     * explicitly set via the __typename field, or we determine the type type from a list of possible
     * types by checking if a fixture exists for the object.
     */
    private getObjectType(value: unknown, possibleTypes: readonly GraphQLObjectType[]): GraphQLObjectType | null {
        if (isObjectWithTypename(value)) {
            const type = this.options.schema.getType(value.__typename)
            if (!type) {
                throw new Error(`Unknown type '${value.__typename}'.`)
            }
            if (!isObjectType(type)) {
                throw new Error(`Type '${value.__typename}' is not an object type.`)
            }
            return type
        } else if (isObject(value)) {
            for (const type of possibleTypes) {
                const keyFieldName = this.options.typePolicies?.[type.name]?.keyField ?? 'id'
                const keyField = type.getFields()[keyFieldName]
                if (!keyField) {
                    continue
                }
                if (keyFieldName in value) {
                    const id = String(value[keyFieldName])
                    const cacheKey = this.getCacheKey(type, id)
                    if (this.objectStore.has(cacheKey)) {
                        return type
                    }
                }
            }
        }

        return null
    }
}
