import storage from '../../extension/storage'
import { FeatureFlags } from '../../extension/types'

/**
 * Gets the value of a feature flag.
 * @param key is the key the feature flag is stored under.
 */
export function get<K extends keyof FeatureFlags>(key: K): Promise<FeatureFlags[K]> {
    return new Promise(resolve => storage.getSync(({ featureFlags }) => resolve(featureFlags[key])))
}

/**
 * Set the value of a feature flag.
 * @param key
 * @param val
 * @returns a promise that resolves with the new value.
 */
export function set<K extends keyof FeatureFlags>(key: K, val: FeatureFlags[K]): Promise<FeatureFlags[K]> {
    return new Promise(resolve =>
        storage.getSync(({ featureFlags }) =>
            storage.setSync({ featureFlags: { ...featureFlags, [key]: val } }, () => get(key).then(resolve))
        )
    )
}

/**
 * Checks to see if the feature flag is set to a truthy value. Only useful for on/off feature flags.
 * @param key
 */
export function isEnabled<K extends keyof FeatureFlags>(key: K): Promise<boolean> {
    return get(key).then(val => !!val)
}

/** Toggle boolean feature flags. */
export function toggle<K extends keyof FeatureFlags>(key: K): Promise<FeatureFlags[K]> {
    return get(key).then(val => set(key, !val))
}
