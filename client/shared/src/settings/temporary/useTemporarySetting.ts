import { useCallback, useContext, useEffect, useMemo, useState } from 'react'

import { type ObservableStatus, useObservableWithStatus } from '@sourcegraph/wildcard'

import type { TemporarySettings } from './TemporarySettings'
import { TemporarySettingsContext } from './TemporarySettingsProvider'

export type UseTemporarySettingsReturnType<K extends keyof TemporarySettings> = [
    TemporarySettings[K],
    (newValue: TemporarySettings[K] | ((previousValue: TemporarySettings[K]) => TemporarySettings[K])) => void,
    TemporarySettingsLoadStatus
]

type TemporarySettingsLoadStatus = 'initial' | 'loaded' | 'error'

const mapObservableStatusToStatus = (observableStatus: ObservableStatus): TemporarySettingsLoadStatus => {
    switch (observableStatus) {
        case 'initial': {
            return 'initial'
        }
        case 'error': {
            return 'error'
        }
        default: {
            return 'loaded'
        }
    }
}

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
): UseTemporarySettingsReturnType<K> => {
    const temporarySettings = useContext(TemporarySettingsContext)

    const [observableValue, observableStatus] = useObservableWithStatus(
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

    // Using useState to handle all setValueAndSave(previousValue => newValue) calls
    // since when using directly temporary settings doesn't keep previousValue up to date
    const [value, setValue] = useState(observableValue)

    // Using separate "status" state to be in sync with "value"
    const [status, setStatus] = useState<TemporarySettingsLoadStatus>('initial')

    useEffect(() => {
        setStatus(mapObservableStatusToStatus(observableStatus))
        setValue(observableValue)
    }, [key, observableStatus, observableValue])

    const setValueAndSave: typeof setValue = useCallback(
        newValue =>
            setValue(previousValue => {
                const finalValue = typeof newValue === 'function' ? newValue(previousValue) : newValue
                temporarySettings.set(key, finalValue)
                return finalValue
            }),
        [key, temporarySettings]
    )

    return [value, setValueAndSave, status]
}
