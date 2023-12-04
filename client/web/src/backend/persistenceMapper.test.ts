import { describe, expect, it } from 'vitest'

import { persistenceMapper, ROOT_QUERY_KEY } from './persistenceMapper'

describe('persistenceMapper', () => {
    const userKey = 'User:01'
    const settingsKey = 'Settings:01'

    const createStringifiedCache = (rootQuery: Record<string, unknown>, references?: Record<string, unknown>) =>
        ({
            [ROOT_QUERY_KEY]: {
                __typename: 'Query',
                ...rootQuery,
            },
            ...references,
        } as const)

    it('does not persist anything if the cache is empty', async () => {
        const persistedString = await persistenceMapper({})

        expect(Object.keys(persistedString)).toEqual([])
    })

    it('persists only hardcoded queries', async () => {
        const persistedString = await persistenceMapper(
            createStringifiedCache({
                viewerSettings: { empty: null, data: true },
                shouldNotBePersisted: {},
            })
        )

        expect(Object.keys(persistedString.ROOT_QUERY!)).not.toContain('shouldNotBePersisted')
    })

    it('persists cache references', async () => {
        const persistedString = await persistenceMapper(
            createStringifiedCache(
                {
                    viewerSettings: { data: { __ref: userKey } },
                    shouldNotBePersisted: {},
                },
                {
                    [userKey]: { settings: { __ref: settingsKey } },
                    [settingsKey]: { data: true },
                }
            )
        )

        expect(Object.keys(persistedString)).toEqual([ROOT_QUERY_KEY, userKey, settingsKey])
    })

    it('persists array of cache references', async () => {
        const persistedString = await persistenceMapper(
            createStringifiedCache(
                {
                    viewerSettings: { data: [{ __ref: userKey }, { __ref: settingsKey }] },
                    shouldNotBePersisted: {},
                },
                {
                    [userKey]: { settings: { __ref: settingsKey } },
                    [settingsKey]: { data: true },
                }
            )
        )

        expect(Object.keys(persistedString)).toEqual([ROOT_QUERY_KEY, userKey, settingsKey])
    })

    it('persists deeply nested cache references', async () => {
        const persistedString = await persistenceMapper(
            createStringifiedCache(
                {
                    viewerSettings: { data: { settings: { sourcegraph: { user: { __ref: userKey } } } } },
                    shouldNotBePersisted: {},
                },
                {
                    [userKey]: { data: true },
                }
            )
        )

        expect(Object.keys(persistedString)).toEqual([ROOT_QUERY_KEY, userKey])
    })
})
