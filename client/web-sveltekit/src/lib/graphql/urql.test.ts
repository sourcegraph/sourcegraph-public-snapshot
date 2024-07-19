import { type AnyVariables, Client, type OperationResult, CombinedError, cacheExchange } from '@urql/core'
import { describe, test, expect, vi, beforeEach, afterEach } from 'vitest'
import { pipe, filter, map, merge } from 'wonka'

import { IncrementalRestoreStrategy, infinityQuery } from './urql'

describe('infinityQuery', () => {
    function getMockClient(responses: Partial<OperationResult<any, AnyVariables>>[]): Client {
        return new Client({
            url: '#testingonly',
            exchanges: [
                cacheExchange, // This is required because infiniteQuery expects that a cache exchange is present
                ({ forward }) =>
                    operations$ => {
                        const mockResults$ = pipe(
                            operations$,
                            filter(operation => {
                                switch (operation.kind) {
                                    case 'query':
                                    case 'mutation':
                                        return true
                                    default:
                                        return false
                                }
                            }),
                            map((operation): OperationResult<any, AnyVariables> => {
                                const response = responses.shift()
                                if (!response) {
                                    return {
                                        operation,
                                        error: new CombinedError({
                                            networkError: new Error('No more responses'),
                                        }),
                                        stale: false,
                                        hasNext: false,
                                    }
                                }
                                return {
                                    ...response,
                                    operation,
                                    data: response.data ?? undefined,
                                    error: response.error ?? undefined,
                                    stale: false,
                                    hasNext: false,
                                }
                            })
                        )

                        const forward$ = pipe(
                            operations$,
                            filter(operation => {
                                switch (operation.kind) {
                                    case 'query':
                                    case 'mutation':
                                        return false
                                    default:
                                        return true
                                }
                            }),
                            forward
                        )

                        return merge([mockResults$, forward$])
                    },
            ],
        })
    }

    function getQuery(client: Client) {
        return infinityQuery({
            client,
            query: 'query { list { nodes { id } } pageInfo { hasNextPage, endCursor } } }',
            variables: {
                first: 2,
                afterCursor: null as string | null,
            },
            map: result => {
                const list = result.data?.list
                return {
                    nextVariables: list?.pageInfo.hasNextPage ? { afterCursor: list.pageInfo.endCursor } : undefined,
                    data: list?.nodes,
                    error: result.error,
                }
            },
            merge: (prev, next) => [...(prev ?? []), ...(next ?? [])],
            createRestoreStrategy: api =>
                new IncrementalRestoreStrategy(
                    api,
                    n => n.length,
                    n => ({ first: n.length })
                ),
        })
    }

    let query: ReturnType<typeof getQuery>

    beforeEach(() => {
        vi.useFakeTimers()

        const client = getMockClient([
            {
                data: {
                    list: {
                        nodes: [{ id: 1 }, { id: 2 }],
                        pageInfo: {
                            hasNextPage: true,
                            endCursor: '2',
                        },
                    },
                },
            },
            {
                data: {
                    list: {
                        nodes: [{ id: 3 }, { id: 4 }],
                        pageInfo: {
                            hasNextPage: true,
                            endCursor: '4',
                        },
                    },
                },
            },
            {
                data: {
                    list: {
                        nodes: [{ id: 5 }, { id: 6 }],
                        pageInfo: {
                            hasNextPage: false,
                        },
                    },
                },
            },
        ])
        query = getQuery(client)
    })

    afterEach(() => {
        vi.useRealTimers()
    })

    test('fetch more', async () => {
        const subscribe = vi.fn()
        query.subscribe(subscribe)

        await vi.runAllTimersAsync()

        // 1. call: fetching -> true
        // 2. call: result
        expect(subscribe).toHaveBeenCalledTimes(2)
        expect(subscribe.mock.calls[0][0]).toMatchObject({
            fetching: true,
        })
        expect(subscribe.mock.calls[1][0]).toMatchObject({
            fetching: false,
            data: [{ id: 1 }, { id: 2 }],
            nextVariables: {
                afterCursor: '2',
            },
        })

        // Fetch more data
        query.fetchMore()
        await vi.runAllTimersAsync()

        // 3. call: fetching -> true
        // 4. call: result
        expect(subscribe).toHaveBeenCalledTimes(4)
        expect(subscribe.mock.calls[2][0]).toMatchObject({
            fetching: true,
        })
        expect(subscribe.mock.calls[3][0]).toMatchObject({
            fetching: false,
            data: [{ id: 1 }, { id: 2 }, { id: 3 }, { id: 4 }],
            nextVariables: {
                afterCursor: '4',
            },
        })
    })

    test('restoring state', async () => {
        const subscribe = vi.fn()
        query.subscribe(subscribe)
        const snapshot = query.capture()
        query.restore({ ...snapshot!, count: 6 })
        await vi.runAllTimersAsync()

        expect(subscribe.mock.calls[3][0]).toMatchObject({
            fetching: false,
            data: [{ id: 1 }, { id: 2 }, { id: 3 }, { id: 4 }, { id: 5 }, { id: 6 }],
        })
    })
})
