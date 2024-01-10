import { describe, expect, test } from 'vitest'

import { GRAPHQL_URI } from './constants'
import { buildGraphQLUrl, gql } from './graphql'

describe('buildGraphQLUrl', () => {
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

    const testCases: [string, string | undefined, string | undefined, string][] = [
        ['when "baseUrl" & "request" are empty', undefined, undefined, GRAPHQL_URI],
        ['when "request" is empty', undefined, 'https://example.com', `https://example.com${GRAPHQL_URI}`],
        [
            'when "baseUrl" is empty & "request" is mutation type',
            EXAMPLE_MUTATION_REQUEST,
            undefined,
            `${GRAPHQL_URI}?MyMutation`,
        ],
        [
            'when "baseUrl" is empty & "request" is query type',
            EXAMPLE_QUERY_REQUEST,
            undefined,
            `${GRAPHQL_URI}?MyQuery`,
        ],
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

    test.each(testCases)('correctly constructs %s', (_title, request, baseUrl, expectedResult) => {
        expect(buildGraphQLUrl({ request, baseUrl })).toEqual(expectedResult)
    })
})
