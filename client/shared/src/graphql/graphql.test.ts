import { GRAPHQL_URI } from './constants'
import { buildRequestURL, gql } from './graphql'

describe('buildRequestURL', () => {
    const EXAMPLE_QUERY_REQUEST = gql`
        query MyQuery {
            repository {
                name
            }
        }
    `

    const EXAMPLE_MUTATION_REQUEST = gql`
        mutation MyMutation {
            deleteUser(user: "12345") {
                alwaysNil
            }
        }
    `
    const testCases: [string, string, string | undefined, string][] = [
        [
            'when "baseUrl" is empty & "mutation" type request',
            EXAMPLE_MUTATION_REQUEST,
            undefined,
            `${GRAPHQL_URI}?MyMutation`,
        ],
        ['when "baseUrl" is empty & "query" type request', EXAMPLE_QUERY_REQUEST, undefined, `${GRAPHQL_URI}?MyQuery`],
        [
            'when "baseUrl" is a domain',
            EXAMPLE_QUERY_REQUEST,
            'https://example.com',
            `https://example.com${GRAPHQL_URI}?MyQuery`,
        ],
        [
            'when "baseUrl" contains trailing /',
            EXAMPLE_QUERY_REQUEST,
            'https://example.com/',
            `https://example.com${GRAPHQL_URI}?MyQuery`,
        ],
        [
            'when "baseUrl" contains path',
            EXAMPLE_QUERY_REQUEST,
            'https://example.com/sourcegraph',
            `https://example.com/sourcegraph${GRAPHQL_URI}?MyQuery`,
        ],
    ]

    test.each(testCases)('correctly constructs %s', (title, query, baseURL, expectedResult) => {
        expect(buildRequestURL(query, baseURL)).toEqual(expectedResult)
    })
})
