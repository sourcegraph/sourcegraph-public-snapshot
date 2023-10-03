import { getContext } from 'svelte'
import { readable, writable, type Readable, type Writable } from 'svelte/store'

import type { SettingsCascade, AuthenticatedUser, TemporarySettingsStorage } from '$lib/shared'

import type { FeatureFlag } from './featureflags'
import type { GraphQLClient } from './graphql'

export { isLightTheme } from './theme'

export interface SourcegraphContext {
    settings: Readable<SettingsCascade['final'] | null>
    user: Readable<AuthenticatedUser | null>
    temporarySettingsStorage: Readable<TemporarySettingsStorage>
    featureFlags: Readable<FeatureFlag[]>
    client: Readable<GraphQLClient>
}

export const KEY = '__sourcegraph__'

export function getStores(): SourcegraphContext {
    const { settings, user, temporarySettingsStorage, featureFlags, client } = getContext<SourcegraphContext>(KEY)
    return { settings, user, temporarySettingsStorage, featureFlags, client }
}

export const user = {
    subscribe(subscriber: (user: AuthenticatedUser | null) => void) {
        const { user } = getStores()
        return user.subscribe(subscriber)
    },
}

export const settings = {
    subscribe(subscriber: (settings: SettingsCascade['final'] | null) => void) {
        const { settings } = getStores()
        return settings.subscribe(subscriber)
    },
}

export const graphqlClient = {
    subscribe(subscriber: (client: GraphQLClient) => void) {
        const { client } = getStores()
        return client.subscribe(subscriber)
    },
}

/**
 * A store that updates every second to return the current time.
 */
export const currentDate: Readable<Date> = readable(new Date(), set => {
    set(new Date())
    const interval = setInterval(() => set(new Date()), 1000)
    return () => clearInterval(interval)
})

/**
 * This store syncs the provided value with localStorage. Values must be JSON (de)seralizable.
 */
export function createLocalWritable<T>(localStorageKey: string, defaultValue: T): Writable<T> {
    const { subscribe, set, update } = writable(defaultValue, set => {
        const existingValue = localStorage.getItem(localStorageKey)
        if (existingValue) {
            set(JSON.parse(existingValue))
        }
    })

    return {
        subscribe,
        set: value => {
            set(value)
            localStorage.setItem(localStorageKey, JSON.stringify(value))
        },
        update: fn => {
            update(value => {
                const newValue = fn(value)
                localStorage.setItem(localStorageKey, JSON.stringify(newValue))
                return newValue
            })
        },
    }
}

export const scrollAll = writable(false)
