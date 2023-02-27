import { persistenceMapper, ROOT_QUERY_KEY, CacheObject } from './persistenceMapper'

describe('persistenceMapper', () => {
    const userKey = 'User:01'
    const settingsKey = 'Settings:01'

    const createStringifiedCache = (rootQuery: Record<string, unknown>, references?: Record<string, unknown>) =>
        JSON.stringify({
            [ROOT_QUERY_KEY]: {
                __typename: 'query',
                ...rootQuery,
            },
            ...references,
        })

    const parseCacheString = (cacheString: string) => JSON.parse(cacheString) as CacheObject

    it('does not persist anything if the cache is empty', async () => {
        const persistedString = await persistenceMapper(JSON.stringify({}))

        expect(Object.keys(parseCacheString(persistedString))).toEqual([])
    })

    it('persists only hardcoded queries', async () => {
        const persistedString = await persistenceMapper(
            createStringifiedCache({
                viewerSettings: { empty: null, data: true },
                shouldNotBePersisted: {},
            })
        )

        expect(Object.keys(parseCacheString(persistedString).ROOT_QUERY)).not.toContain('shouldNotBePersisted')
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

        expect(Object.keys(parseCacheString(persistedString))).toEqual([ROOT_QUERY_KEY, userKey, settingsKey])
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

        expect(Object.keys(parseCacheString(persistedString))).toEqual([ROOT_QUERY_KEY, userKey, settingsKey])
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

        expect(Object.keys(parseCacheString(persistedString))).toEqual([ROOT_QUERY_KEY, userKey])
    })
})
