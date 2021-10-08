import { useCallback, useContext, useEffect, useState } from 'react'

import { TemporarySettings } from './TemporarySettings'
import { TemporarySettingsContext } from './TemporarySettingsProvider'
import { SettingResponse } from './TemporarySettingsStorage'

/**
 * React Hook to get and set a single temporary setting.
 * The setting's value will be kept up to date if another part of the app changes
 * it.
 * Note: The setting might be loaded asynchronously, in which case the
 * first emitted value will be `undefined`. You have to take necessary steps to
 * ensure that your UI renders correctly during this "loading" phase.
 *
 * @param key - name of the setting
 * @param defaultValue - value to use when the setting hasn't been set yet
 */
export const useTemporarySetting = <K extends keyof TemporarySettings, D extends TemporarySettings[K]>(
    key: K,
    defaultValue: D
): [
    SettingResponse<K, D>,
    (newValue: TemporarySettings[K] | ((oldValue: TemporarySettings[K]) => TemporarySettings[K])) => void
] => {
    const temporarySettings = useContext(TemporarySettingsContext)
    const [response, setResponse] = useState<SettingResponse<K, D>>({
        loading: true,
    })

    useEffect(
        () => {
            const subscription = temporarySettings.get(key, defaultValue).subscribe({ next: setResponse })
            return () => subscription.unsubscribe()
        },
        // `defaultValue` should not be a dependency, otherwise the
        // observable would be recomputed if the caller used e.g. an object
        // literal as default value. `useTemporarySetting` works more like
        // `useState` in this regard.
        // eslint-disable-next-line react-hooks/exhaustive-deps
        [temporarySettings, key]
    )

    const setValueAndSave = useCallback(
        (newValue: TemporarySettings[K] | ((oldValue: TemporarySettings[K]) => TemporarySettings[K])): void => {
            let finalValue: TemporarySettings[K]
            if (typeof newValue === 'function') {
                finalValue = newValue('value' in response ? response.value : undefined)
            } else {
                finalValue = newValue
            }
            temporarySettings.set(key, finalValue)
        },
        [key, response, temporarySettings]
    )

    return [response, setValueAndSave]
}
