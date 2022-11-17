import { getContext } from 'svelte'
import { readable, type Readable } from 'svelte/store'

import type { GraphQLClient } from '$lib/http-client'
import type { SettingsCascade, AuthenticatedUser, PlatformContext, TemporarySettingsStorage } from '$lib/shared'
import { getWebGraphQLClient } from '$lib/web'

export interface SourcegraphContext {
    settings: Readable<SettingsCascade['final'] | null>
    user: Readable<AuthenticatedUser | null>
    platformContext: Readable<PlatformContext | null>
    isLightTheme: Readable<boolean>
    temporarySettingsStorage: Readable<TemporarySettingsStorage>
}

export const KEY = '__sourcegraph__'

export function getStores(): SourcegraphContext {
    const { settings, user, platformContext, isLightTheme, temporarySettingsStorage } =
        getContext<SourcegraphContext>(KEY)
    return { settings, user, platformContext, isLightTheme, temporarySettingsStorage }
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

export const platformContext = {
    subscribe(subscriber: (platformContext: PlatformContext | null) => void) {
        const { platformContext } = getStores()
        return platformContext.subscribe(subscriber)
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

// TODO: Standardize on getWebGraphQLCient or platformContext.requestGraphQL
export const graphqlClient = readable<GraphQLClient | null>(null, set => {
    // no-void conflicts with no-floating-promises
    // eslint-disable-next-line no-void
    void getWebGraphQLClient().then(client => set(client))
})
