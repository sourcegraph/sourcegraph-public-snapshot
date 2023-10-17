import type { InMemoryCache, NormalizedCacheObject } from '@apollo/client'
import { type PersistentStorage, CachePersistor } from 'apollo3-cache-persist'

import { cache } from '@sourcegraph/shared/src/backend/apolloCache'

import { type CacheObject, persistenceMapper, ROOT_QUERY_KEY } from './persistenceMapper'

/**
 * 🚨 SECURITY: Use two unique keys for authenticated and anonymous users
 * to avoid keeping private information in localStorage after logout.
 */
const getApolloPersistCacheKey = (isAuthenticated: boolean): string =>
    `apollo-cache-persist-${isAuthenticated ? 'authenticated' : 'anonymous'}`

/**
 * Persistence storage is based on `localStorage`, which uses the
 * data preloaded on the server and transferred to the client via `window.context`.
 */
class LocalStorageWrapper implements PersistentStorage<CacheObject | null> {
    constructor(private preloadedQueries: Record<string, unknown>) {}

    public removeItem(key: string): void {
        window.localStorage.removeItem(key)
    }

    public setItem(key: string, value: CacheObject | null): void {
        window.localStorage.setItem(key, JSON.stringify(value))
    }

    public getItem(key: string): CacheObject | null {
        const maybeData = window.localStorage.getItem(key) ?? '{}'
        const persistedData = JSON.parse(maybeData) as CacheObject

        const hydratedData: CacheObject = {
            ...persistedData,
            [ROOT_QUERY_KEY]: {
                __typename: 'Query',
                ...persistedData[ROOT_QUERY_KEY],
                ...this.preloadedQueries,
            },
        }

        return hydratedData
    }
}

interface GetPersistentCacheOptions {
    isAuthenticatedUser: boolean
    preloadedQueries: Record<string, unknown>
}

export async function getPersistentCache(options: GetPersistentCacheOptions): Promise<InMemoryCache> {
    const { isAuthenticatedUser, preloadedQueries } = options

    const persistor = new CachePersistor<NormalizedCacheObject>({
        cache,
        persistenceMapper,
        // Use max 4 MB for persistent cache. Leave 1 MB for other means out of 5 MB available.
        // If exceeded, persistence will pause and app will start up cold on next launch.
        maxSize: 1024 * 1024 * 4,
        key: getApolloPersistCacheKey(isAuthenticatedUser),
        storage: new LocalStorageWrapper(preloadedQueries) as any, // `as any` is required because third-party types are incorrect.
        serialize: false as true, // `as true` is required because third-party types are incorrect.
    })

    // 🚨 SECURITY: Drop persisted cache item in case `isAuthenticatedUser` value changed.
    localStorage.removeItem(getApolloPersistCacheKey(!isAuthenticatedUser))
    await persistor.restore()

    return cache
}
