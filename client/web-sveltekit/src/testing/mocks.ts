import { faker } from '@faker-js/faker'
import signale from 'signale'
import { writable, type Readable, type Writable } from 'svelte/store'
import { vi } from 'vitest'

import { KEY, type SourcegraphContext } from '$lib/stores'
import type { FeatureFlagName } from '$lib/web'

let fakerRefDate: Date

/**
 * Use fake timers and optionally set the current date and reference date for data generation.
 */
export function useFakeTimers(refDate?: Date) {
    if (!refDate) {
        refDate = faker.defaultRefDate()
    } else {
        fakerRefDate = faker.defaultRefDate()
        faker.setDefaultRefDate(refDate)
    }
    vi.useFakeTimers()
    vi.setSystemTime(refDate)
    faker.setDefaultRefDate(refDate)
}

/**
 * Use real timers. The reference date for date generation will be
 * restored to a fixed default value.
 */
export function useRealTimers() {
    faker.setDefaultRefDate(fakerRefDate)
    vi.useFakeTimers()
    vi.useRealTimers()
}

// Stores all mocked context values
export let mockedContexts = new Map<any, any>()

type SourcegraphContextKey = keyof SourcegraphContext
type MockedSourcegraphContextValue<T> = T extends Readable<infer U> ? Writable<U> : T

// Sets up stubs for mocking the Sourcegraph context. The sourcegraph context makes
// certain values available app-wide by using Svelte context API.
const unmocked: unique symbol = Symbol('unmocked')
const mockedSourcgraphContext: {
    [key in SourcegraphContextKey]: MockedSourcegraphContextValue<SourcegraphContext[key]> | typeof unmocked
} = {
    user: writable(null),
    client: unmocked,
    settings: writable({}),
    featureFlags: writable([]),
    isLightTheme: writable(true),
    temporarySettingsStorage: unmocked,
}

// Creates a proxy object for the mocked Sourcegraph context object.
// If a value hasn't been mocked a warning is printed.
mockedContexts.set(
    KEY,
    Object.defineProperties(
        {},
        Object.fromEntries(
            Object.keys(mockedSourcgraphContext).map(key => [
                key,
                {
                    get: () => {
                        if (mockedSourcgraphContext[key as SourcegraphContextKey] === unmocked) {
                            signale.warn(`Sourcegraph context ${key} is unmocked`)
                        }
                        return mockedSourcgraphContext[key as SourcegraphContextKey]
                    },
                },
            ])
        )
    )
)

/**
 * Sets the app's feature flags to the provided value. If the function is called multiple times without
 * calling `unmockFeatureFlags` in between then subsequent calls will update the underlying feature flag
 * store, updating all subscribers.
 */
export function mockFeatureFlags(evaluatedFeatureFlags: Partial<Record<FeatureFlagName, boolean>>) {
    const flags = Object.entries(evaluatedFeatureFlags).map(([name, value]) => ({ name, value }))

    if (mockedSourcgraphContext.featureFlags === unmocked) {
        mockedSourcgraphContext.featureFlags = writable(flags)
    } else {
        mockedSourcgraphContext.featureFlags.set(flags)
    }
}

/**
 * Unmock all feature flags.
 */
export function unmockFeatureFlags() {
    mockedSourcgraphContext.featureFlags = unmocked
}
