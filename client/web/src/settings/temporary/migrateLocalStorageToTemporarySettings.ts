import { take } from 'rxjs/operators'

import { TemporarySettings } from './TemporarySettings'
import { TemporarySettingsStorage } from './TemporarySettingsStorage'

interface Migration {
    localStorageKey: string
    temporarySettingsKey: keyof TemporarySettings
    type: 'boolean' | 'number'
}

const migrations: Migration[] = [
    {
        localStorageKey: 'has-cancelled-onboarding-tour',
        temporarySettingsKey: 'search.onboarding.tourCancelled',
        type: 'boolean',
    },
    {
        localStorageKey: 'days-active-count',
        temporarySettingsKey: 'user.daysActiveCount',
        type: 'number',
    },
    {
        localStorageKey: 'has-dismissed-survey-toast',
        temporarySettingsKey: 'npsSurvey.hasTemporarilyDismissed',
        type: 'boolean',
    },
    {
        localStorageKey: 'has-permanently-dismissed-survey-toast',
        temporarySettingsKey: 'npsSurvey.hasPermanentlyDismissed',
        type: 'boolean',
    },
    {
        localStorageKey: 'finished-welcome-flow',
        temporarySettingsKey: 'signup.finishedWelcomeFlow',
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
                } else if (migration.type === 'number') {
                    storage.set(migration.temporarySettingsKey, parseInt(value, 10))
                }
                localStorage.removeItem(migration.localStorageKey)
            }
        }
    }
}
