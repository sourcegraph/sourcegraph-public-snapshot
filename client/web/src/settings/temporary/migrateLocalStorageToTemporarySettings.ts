import { take } from 'rxjs/operators'

import { TemporarySettings } from './TemporarySettings'
import { TemporarySettingsStorage } from './TemporarySettingsStorage'

interface Migration {
    localStorageKey: string
    temporarySettingsKey: keyof TemporarySettings
    type: 'boolean'
}

const migrations: Migration[] = [
    {
        localStorageKey: 'has-cancelled-onboarding-tour',
        temporarySettingsKey: 'search.onboarding.tourCancelled',
        type: 'boolean',
    },
]

export async function migrateLocalStorageToTemporarySettings(storage: TemporarySettingsStorage): Promise<void> {
    for (const migration of migrations) {
        // Use the first value of the setting to check if it exists.
        // Only migrate if the setting is not already set.
        const temporarySetting = await storage.get(migration.temporarySettingsKey).pipe(take(1)).toPromise()
        if (typeof temporarySetting === 'undefined') {
            const value = localStorage.getItem(migration.localStorageKey)
            if (value) {
                if (migration.type === 'boolean') {
                    storage.set(migration.temporarySettingsKey, value === 'true')
                }
                localStorage.removeItem(migration.localStorageKey)
            }
        }
    }
}
