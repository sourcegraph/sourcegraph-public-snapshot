import { useCallback, useContext, useEffect, useMemo, useRef } from 'react'

import { useObservable } from '@sourcegraph/wildcard'

import { TemporarySettings } from './TemporarySettings'
import { TemporarySettingsContext } from './TemporarySettingsProvider'

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
export const useTemporarySetting = <K extends keyof TemporarySettings>(
    key: K,
    defaultValue?: TemporarySettings[K]
): [
    TemporarySettings[K],
    (newValue: TemporarySettings[K] | ((oldValue: TemporarySettings[K]) => TemporarySettings[K])) => void
] => {
    const temporarySettings = useContext(TemporarySettingsContext)

    const updatedValue = useObservable(
        useMemo(
            () => temporarySettings.get(key, defaultValue),
            // `defaultValue` should not be a dependency, otherwise the
            // observable would be recomputed if the caller used e.g. an object
            // literal as default value. `useTemporarySetting` works more like
            // `useState` in this regard.
            // eslint-disable-next-line react-hooks/exhaustive-deps
            [temporarySettings, key]
        )
    )

    // We want to avoid the setter function to be recreated whenever the
    // temporary settings value changes. To do this, we create a reference that
    // always points to the latest updated value.
    const updatedValueReference = useRef<TemporarySettings[K] | undefined>(updatedValue)
    useEffect(() => {
        updatedValueReference.current = updatedValue
    }, [updatedValue])

    const setValueAndSave = useCallback(
        (newValue: TemporarySettings[K] | ((oldValue: TemporarySettings[K]) => TemporarySettings[K])): void => {
            let finalValue: TemporarySettings[K]
            if (typeof newValue === 'function') {
                finalValue = newValue(updatedValueReference.current)
            } else {
                finalValue = newValue
            }
            temporarySettings.set(key, finalValue)
        },
        [key, temporarySettings]
    )

    return [updatedValue, setValueAndSave]
}
