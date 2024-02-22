import { buildSchema } from 'graphql'
import { describe, test, expect, beforeEach } from 'vitest'

import { GraphQLMockServer } from './graphql-mocking'

const schema = buildSchema(`
    enum UserStatus {
        ACTIVE
        INACTIVE
    }

    scalar Date

    type Query {
        string: String!
        int: Int!
        float: Float!
        boolean: Boolean!
        id: ID!

        node(id: ID!): Node

        viewer: Viewer!
        currentUser: User!
        users: [User!]!
        action: Action!
    }

    interface Node {
        id: ID!
    }

    interface Viewer implements Node {
        id: ID!
        name: String!
    }

    type User implements Viewer & Node {
        id: ID!
        name: String!
        age: Int
        status: UserStatus!
        date: Date!
        friends: [User!]!
    }

    type Organization implements Viewer & Node {
        id: ID!
        name: String!
    }

    type Action {
        key: ID!
        name: String!
    }
`)

describe('default operation', () => {
    test.each([
        ['String', 'string', { string: expect.any(String) }],
        ['Int', 'int', { int: expect.any(Number) }],
        ['Float', 'float', { float: expect.any(Number) }],
        ['Boolean', 'boolean', { boolean: expect.any(Boolean) }],
        ['ID', 'id', { id: expect.any(String) }],
        ['Objects', 'currentUser {id}', { currentUser: { id: expect.any(String) } }],
        ['Lists', 'users {id}', { users: [{ id: expect.any(String) }] }],
        ['Enum', 'currentUser {status}', { currentUser: { status: expect.any(String) } }],
    ])('random data for %s', (_, query, expected) => {
        const server = new GraphQLMockServer({ schema, mocks: {} })
        expect(server.query(`query {${query}}`)).toMatchObject({ data: expected })
    })

    test('random list length', () => {
        const server = new GraphQLMockServer({ schema, mocks: {}, minListLength: 3, maxListLength: 5 })
        const result = server.query(`query {users {id}}`).data?.users.length
        expect(result).toBeLessThanOrEqual(5)
        expect(result).toBeGreaterThanOrEqual(3)
    })

    test('random interface implementation', () => {
        const server = new GraphQLMockServer({ schema })
        expect(server.query(`query {viewer {__typename name}}`)).toMatchObject({
            data: { viewer: { __typename: expect.any(String), name: expect.any(String) } },
        })
    })

    test('null values', () => {
        {
            const server = new GraphQLMockServer({ schema, nullProbability: 1 })
            expect(server.query(`query {currentUser {age}}`)).toMatchObject({
                data: { currentUser: { age: null } },
            })
        }

        {
            const server = new GraphQLMockServer({ schema, nullProbability: 0 })
            expect(server.query(`query {currentUser {age}}`)).toMatchObject({
                data: { currentUser: { age: expect.any(Number) } },
            })
        }
    })
})

describe('custom mocks', () => {
    // If boolean values are not properly handled, this query will always return 'true'
    test('boolean', () => {
        const server = new GraphQLMockServer({
            schema,
            mocks: {
                Query: () => ({
                    boolean: false,
                }),
                Boolean: () => true,
            },
        })
        expect(server.query(`query {boolean}`)).toMatchObject({
            data: { boolean: false },
        })
    })

    test('custom scalar', () => {
        const server = new GraphQLMockServer({
            schema,
            mocks: {
                Date: () => 'custom date',
            },
        })
        expect(server.query(`query {currentUser {date}}`)).toMatchObject({
            data: { currentUser: { date: 'custom date' } },
        })
    })

    test('partial custom object', () => {
        const server = new GraphQLMockServer({
            schema,
            mocks: {
                User: () => ({
                    name: 'custom name',
                }),
            },
        })

        expect(server.query(`query {currentUser {id name}}`)).toMatchObject({
            data: { currentUser: { id: expect.any(String), name: 'custom name' } },
        })
    })

    test('custom mock overrides nullability', () => {
        const server = new GraphQLMockServer({
            schema,
            mocks: {
                User: () => ({
                    age: 42,
                }),
            },
            nullProbability: 1,
        })

        expect(server.query(`query {currentUser {age}}`)).toMatchObject({
            data: { currentUser: { age: 42 } },
        })
    })

    test('partial custom lists', () => {
        const server = new GraphQLMockServer({
            schema,
            mocks: {
                User: () => ({
                    friends: [{ name: 'friend1' }, { name: 'friend2' }],
                }),
            },
        })

        expect(server.query(`query {currentUser {id friends {id, name}}}`)).toMatchObject({
            data: {
                currentUser: {
                    id: expect.any(String),
                    friends: [
                        { id: expect.any(String), name: 'friend1' },
                        { id: expect.any(String), name: 'friend2' },
                    ],
                },
            },
        })
    })

    test('partial deep custom lists', () => {
        const server = new GraphQLMockServer({
            schema,
            mocks: {
                User: () => ({
                    friends: [
                        { name: 'friend1', friends: [{ name: 'friend2' }] },
                        { name: 'friend2', friends: [{ name: 'friend1' }] },
                    ],
                }),
            },
        })
        expect(server.query(`query {currentUser {id name friends {id name friends {name}}}}`)).toMatchObject({
            data: {
                currentUser: {
                    id: expect.any(String),
                    name: expect.any(String),
                    friends: [
                        {
                            id: expect.any(String),
                            name: 'friend1',
                            friends: [{ name: 'friend2' }],
                        },
                        {
                            id: expect.any(String),
                            name: 'friend2',
                            friends: [{ name: 'friend1' }],
                        },
                    ],
                },
            },
        })
    })

    test('partial root fields', () => {
        const server = new GraphQLMockServer({
            schema,
            mocks: {
                Query: () => ({
                    users: [{ name: 'user1' }, { name: 'user2' }],
                }),
            },
        })
        expect(server.query(`query {users {id name}}`)).toMatchObject({
            data: {
                users: [
                    {
                        id: expect.any(String),
                        name: 'user1',
                    },
                    {
                        id: expect.any(String),
                        name: 'user2',
                    },
                ],
            },
        })
    })

    test('merge multiple mocks - last one wins - operation mocks have precedence', () => {
        const server = new GraphQLMockServer({
            schema,
            mocks: {
                User: () => ({
                    name: 'user1',
                    age: 42,
                }),
            },
        })
        server.addOperationMocks({
            customOperation: () => ({ currentUser: { name: 'custom', friends: [{ name: 'friend1' }] } }),
        })
        server.addTypeMocks({ User: () => ({ name: 'user2' }) })

        expect(server.query(`query {currentUser {name age friends {name age}}}`)).toMatchObject({
            data: { currentUser: { name: 'user2', age: 42, friends: [{ name: 'user2', age: 42 }] } },
        })

        expect(server.query(`query customOperation {currentUser {name age friends {name age}}}`)).toMatchObject({
            data: { currentUser: { name: 'custom', age: 42, friends: [{ name: 'friend1', age: 42 }] } },
        })
    })

    test('override partial non-id object', () => {
        const server = new GraphQLMockServer({
            schema,
            mocks: {
                Action: () => ({
                    name: 'default',
                }),
            },
        })
        server.addOperationMocks({
            customOperation: () => ({
                action: { name: 'custom' },
            }),
        })

        expect(server.query(`query customOperation {action {name}}`)).toMatchObject({
            data: { action: { name: 'custom' } },
        })
    })
})

describe('fixtures', () => {
    const server = new GraphQLMockServer({
        schema,
        fixtures: [
            {
                __typename: 'User',
                id: '1',
                name: 'user1',
            },
            {
                __typename: 'Action',
                key: '5',
                name: 'action',
            },
            {
                __typename: 'Organization',
                id: '2',
                name: 'organization',
            },
        ],
        mocks: {
            Query: () => ({
                currentUser: { id: '1' },
                action: { key: '5' },
            }),
        },
        typePolicies: {
            Action: {
                keyField: 'key',
            },
        },
    })

    beforeEach(() => {
        server.reset()
    })

    test('load fixtures with given ID', () => {
        expect(server.query(`query {currentUser {id name}}`)).toMatchObject({
            data: { currentUser: { id: '1', name: 'user1' } },
        })
    })

    test('load fixtures with different key field', () => {
        expect(server.query(`query {action  {key name}}`)).toMatchObject({
            data: { action: { key: '5', name: 'action' } },
        })
    })

    test('override specific fields per operation', () => {
        server.addTypeMocks({
            User: () => ({
                id: '1',
                name: 'default',
            }),
        })
        server.addOperationMocks({
            customQuery: () => ({
                currentUser: { name: 'custom' },
            }),
        })

        expect(server.query(`query {currentUser {id name}}`)).toMatchObject({
            data: { currentUser: { id: '1', name: 'default' } },
        })
        expect(server.query(`query customQuery {currentUser {id name}}`)).toMatchObject({
            data: { currentUser: { id: '1', name: 'custom' } },
        })
    })

    test('load correct fixture without typename', () => {
        server.addOperationMocks({
            customQuery: () => ({
                viewer: { id: '1' },
            }),
        })

        expect(server.query(`query customQuery {viewer {id name __typename}}`)).toMatchObject({
            data: { viewer: { id: '1', name: 'user1', __typename: 'User' } },
        })

        server.addOperationMocks({
            customQuery: () => ({
                viewer: { id: '2' },
            }),
        })

        expect(server.query(`query customQuery {viewer {id name __typename}}`)).toMatchObject({
            data: { viewer: { id: '2', name: 'organization', __typename: 'Organization' } },
        })
    })
})

describe('interface mocks', () => {
    test('enforce interface implementation', () => {
        let server = new GraphQLMockServer({
            schema,
            mocks: {
                Query: () => ({
                    viewer: { __typename: 'Organization' },
                }),
            },
        })
        expect(server.query(`query {viewer {__typename name}}`)).toMatchObject({
            data: { viewer: { __typename: 'Organization', name: expect.any(String) } },
        })

        server = new GraphQLMockServer({
            schema,
            mocks: {
                Query: () => ({
                    viewer: { __typename: 'User' },
                }),
            },
        })
        expect(server.query(`query {viewer {__typename name}}`)).toMatchObject({
            data: { viewer: { __typename: 'User', name: expect.any(String) } },
        })
    })

    test('custom interface mock', () => {
        const server = new GraphQLMockServer({
            schema,
            mocks: {
                Viewer: () => ({
                    name: 'custom name',
                }),
            },
        })

        expect(server.query(`query {viewer {__typename name}}`)).toMatchObject({
            data: { viewer: { __typename: expect.any(String), name: 'custom name' } },
        })
    })
})

describe('operation overrides', () => {
    test('override specific operations', () => {
        const server = new GraphQLMockServer({
            schema,
            mocks: {
                Query: () => ({
                    currentUser: { name: 'defaultUser' },
                }),
            },
        })
        server.addOperationMocks({
            customOperation: () => ({
                currentUser: { name: 'customUser' },
            }),
        })

        expect(server.query(`query {currentUser {name}}`)).toMatchObject({
            data: { currentUser: { name: 'defaultUser' } },
        })
        expect(server.query(`query customOperation {currentUser {name}}`)).toMatchObject({
            data: { currentUser: { name: 'customUser' } },
        })
    })

    test('works with field aliases', () => {
        const server = new GraphQLMockServer({
            schema,
        })
        server.addOperationMocks({
            customOperation: () => ({
                foo: { name: 'customUser' },
            }),
        })

        expect(server.query(`query customOperation {foo: currentUser {name}}`)).toMatchObject({
            data: { foo: { name: 'customUser' } },
        })
    })
})

describe('resolvers', () => {
    test('default node resolver', () => {
        const server = new GraphQLMockServer({
            schema,
        })
        expect(server.query(`query {node(id: "1") { ... on User {id, name}}}`)).toMatchObject({
            data: { node: { id: '1', name: expect.any(String) } },
        })
    })
})
