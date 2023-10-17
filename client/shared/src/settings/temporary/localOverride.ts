import { Subject } from 'rxjs'

import type { TemporarySettings } from './TemporarySettings'

type TemporarySettingsKey = keyof TemporarySettings

interface FeatureFlagOverride<K extends TemporarySettingsKey> {
    value?: TemporarySettings[K]
}

function buildOverrideKey(key: string): string {
    return `temporarySettingOverride-${key}`
}

export function getTemporarySettingOverride<K extends TemporarySettingsKey>(name: K): FeatureFlagOverride<K> | null {
    const value = localStorage.getItem(buildOverrideKey(name))
    return typeof value === 'string' ? JSON.parse(value) : null
}

export function setTemporarySettingOverride<K extends TemporarySettingsKey>(
    name: K,
    value: FeatureFlagOverride<K>
): void {
    localStorage.setItem(buildOverrideKey(name), JSON.stringify(value))
    temporarySettingsOverrideUpdate.next()
}

export function removeTemporarySettingOverride<K extends TemporarySettingsKey>(flagName: K): void {
    localStorage.removeItem(buildOverrideKey(flagName))
    temporarySettingsOverrideUpdate.next()
}

/**
 * This is a necessary workaround for notifying the temporary storage system when
 * changes have been made.
 */
export const temporarySettingsOverrideUpdate = new Subject<void>()
