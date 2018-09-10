import storage from '../../browser/storage'
import { FeatureFlags } from '../../browser/types'
import { isInPage } from '../context'

interface FeatureFlagsStorage {
    /**
     * Checks to see if the feature flag is set enabled.
     */
    isEnabled<K extends keyof FeatureFlags>(key: K): Promise<boolean>
    /**
     * Enable a feature flag.
     */
    enable<K extends keyof FeatureFlags>(key: K): Promise<boolean>
    /**
     * Disable a feature flag.
     */
    disable<K extends keyof FeatureFlags>(key: K): Promise<boolean>
    /**
     * Set a feature flag.
     */
    set<K extends keyof FeatureFlags>(key: K, enabled: boolean): Promise<boolean>
    /** Toggle a feature flag. */
    toggle<K extends keyof FeatureFlags>(key: K): Promise<boolean>
}

interface FeatureFlagUtilities {
    get<K extends keyof FeatureFlags>(key: K): Promise<boolean>
    set<K extends keyof FeatureFlags>(key: K, enabled: boolean): Promise<boolean>
}

const createFeatureFlagStorage = ({ get, set }: FeatureFlagUtilities): FeatureFlagsStorage => ({
    set,
    enable: key => set(key, true),
    disable: key => set(key, false),
    isEnabled<K extends keyof FeatureFlags>(key: K): Promise<boolean> {
        return get(key).then(val => !!val)
    },
    toggle<K extends keyof FeatureFlags>(key: K): Promise<FeatureFlags[K]> {
        return get(key).then(val => set(key, !val))
    },
})

function bextGet<K extends keyof FeatureFlags>(key: K): Promise<FeatureFlags[K]> {
    return new Promise(resolve => storage.getSync(({ featureFlags }) => resolve(featureFlags[key])))
}

function bextSet<K extends keyof FeatureFlags>(key: K, val: FeatureFlags[K]): Promise<FeatureFlags[K]> {
    return new Promise(resolve =>
        storage.getSync(({ featureFlags }) =>
            storage.setSync({ featureFlags: { ...featureFlags, [key]: val } }, () => bextGet(key).then(resolve))
        )
    )
}

const browserExtensionFeatureFlags = createFeatureFlagStorage({
    get: bextGet,
    set: bextSet,
})

const inPageFeatureFlags = createFeatureFlagStorage({
    get: key => new Promise(resolve => resolve(!!localStorage.getItem(key))),
    set: (key, val) =>
        new Promise(resolve => {
            localStorage.setItem(key, val.toString())
            resolve(val)
        }),
})

export const featureFlags: FeatureFlagsStorage = isInPage ? inPageFeatureFlags : browserExtensionFeatureFlags
