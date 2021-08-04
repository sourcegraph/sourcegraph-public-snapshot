import { useCallback, useContext } from 'react'

import { useObservable } from '@sourcegraph/shared/src/util/useObservable'

import { TemporarySettings } from './TemporarySettings'
import { TemporarySettingsContext } from './TemporarySettingsProvider'

/**
 * React Hook to get and set a single temporary setting.
 * The setting's value will be kept up to date if another part of the app changes it.
 *
 * @param key - name of the setting
 */
export const useTemporarySetting = <K extends keyof TemporarySettings>(
    key: K
): [
    TemporarySettings[K] | null,
    (newValue: TemporarySettings[K] | ((oldValue: TemporarySettings[K]) => TemporarySettings[K])) => void
] => {
    const temporarySettings = useContext(TemporarySettingsContext)

    const updatedValue = useObservable(temporarySettings.get(key))

    const setValueAndSave = useCallback(
        (newValue: TemporarySettings[K] | ((oldValue: TemporarySettings[K]) => TemporarySettings[K])): void => {
            let finalValue: TemporarySettings[K]
            if (typeof newValue === 'function') {
                finalValue = newValue(updatedValue)
            } else {
                finalValue = newValue
            }
            temporarySettings.set(key, finalValue)
        },
        [key, temporarySettings, updatedValue]
    )

    return [updatedValue, setValueAndSave]
}
