import { createMockClient } from '@apollo/client/testing'
import { describe, expect, test } from '@jest/globals'

import { SearchQuery, SearchRepoQuery, SearchResult, UserQuery } from './Query'
import { SourcegraphClient, SourcegraphService, BaseClient } from './SourcegraphClient'

describe('SourcegraphService', () => {
    test('current user query', async () => {
        // setup
        const data = {
            currentUser: {
                username: 'william',
            },
        }

        const query: UserQuery = new UserQuery()
        const base = new BaseClient(createMockClient(data, query.gql()))
        let sourcegraphService: SourcegraphService = new SourcegraphClient(base)
        // test
        const username = await sourcegraphService.Users.currentUsername()
        // verify
        expect(username).toBe('william')
    })
    test('search query with 2 results', async () => {
        // setup
        const expected: SearchResult[] = [
            { repository: 'repo 1', filename: 'filename-1', fileContent: 'logs of bogus content' },
            { repository: 'repo 1', filename: 'filename-2', fileContent: 'logs of bogus content' },
        ]
        const data = {
            search: {
                results: {
                    results: [
                        {
                            __typename: 'FileMatch',
                            repository: { name: expected[0].repository },
                            file: {
                                name: expected[0].filename,
                                content: expected[0].fileContent,
                            },
                        },
                        {
                            __typename: 'FileMatch',
                            repository: { name: expected[1].repository },
                            file: {
                                name: expected[1].filename,
                                content: expected[1].fileContent,
                            },
                        },
                    ],
                },
            },
        }

        const query: SearchQuery = new SearchQuery('mock query')
        const base = new BaseClient(createMockClient(data, query.gql(), query.vars()))
        let sourcegraphService: SourcegraphService = new SourcegraphClient(base)
        // test
        const results: SearchResult[] = await sourcegraphService.Search.searchQuery('mock query')
        // verify
        expect(results).toEqual(expected)
    })
    test('search query with 0 results', async () => {
        // setup
        const expected: SearchResult[] = []
        const data = {
            search: {
                results: {
                    results: [],
                },
            },
        }

        const query: SearchQuery = new SearchQuery('mock query')
        const base = new BaseClient(createMockClient(data, query.gql(), query.vars()))
        let sourcegraphService: SourcegraphService = new SourcegraphClient(base)
        // test
        const results: SearchResult[] = await sourcegraphService.Search.searchQuery('mock query')
        // verify
        expect(results).toEqual(expected)
    })
    test('search query with only repository names', async () => {
        // setup
        const expected: SearchResult[] = [
            { repository: 'repo-1', filename: '', fileContent: '' },
            { repository: 'repo-2', filename: '', fileContent: '' },
        ]
        const data = {
            search: {
                results: {
                    results: [
                        {
                            __typename: 'FileMatch',
                            repository: { name: 'repo-1' },
                        },
                        {
                            __typename: 'FileMatch',
                            repository: { name: 'repo-2' },
                        },
                    ],
                },
            },
        }

        const query: SearchRepoQuery = new SearchRepoQuery('mock query')
        const base = new BaseClient(createMockClient(data, query.gql(), query.vars()))
        let sourcegraphService: SourcegraphService = new SourcegraphClient(base)
        // test
        const results: SearchResult[] = await sourcegraphService.Search.doQuery(query)
        // verify
        expect(results).toEqual(expected)
    })
})
