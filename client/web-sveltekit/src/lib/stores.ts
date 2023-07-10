import { getContext } from 'svelte'
import { readable, writable, type Readable, type Writable } from 'svelte/store'

import type { GraphQLClient } from '$lib/http-client'
import type { SettingsCascade, AuthenticatedUser, TemporarySettingsStorage } from '$lib/shared'
import { getWebGraphQLClient } from '$lib/web'

export interface SourcegraphContext {
    settings: Readable<SettingsCascade['final'] | null>
    user: Readable<AuthenticatedUser | null>
    isLightTheme: Readable<boolean>
    temporarySettingsStorage: Readable<TemporarySettingsStorage>
}

export const KEY = '__sourcegraph__'

export function getStores(): SourcegraphContext {
    const { settings, user, isLightTheme, temporarySettingsStorage } = getContext<SourcegraphContext>(KEY)
    return { settings, user, isLightTheme, temporarySettingsStorage }
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

export const isLightTheme = {
    subscribe(subscriber: (isLightTheme: boolean) => void) {
        const { isLightTheme } = getStores()
        return isLightTheme.subscribe(subscriber)
    },
}

/**
 * A store that updates every second to return the current time.
 */
export const currentDate: Readable<Date> = readable(new Date(), set => {
    const interval = setInterval(() => set(new Date()), 1000)
    return () => clearInterval(interval)
})

export const graphqlClient = readable<GraphQLClient | null>(null, set => {
    // no-void conflicts with no-floating-promises
    // eslint-disable-next-line no-void
    void getWebGraphQLClient().then(client => set(client))
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
