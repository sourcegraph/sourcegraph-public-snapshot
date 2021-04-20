import { useCallback, useMemo } from 'react'

import { LocalStorageSubject } from './LocalStorageSubject'
import { useObservable } from './useObservable'

export const REDESIGN_TOGGLE_KEY = 'isRedesignEnabled'
export const REDESIGN_CLASS_NAME = 'theme-redesign'

export const getIsRedesignEnabled = (): boolean => localStorage.getItem(REDESIGN_TOGGLE_KEY) === 'true'

export const useRedesignToggle = (): [boolean | undefined, (value: boolean) => void] => {
    const subject = useMemo(() => new LocalStorageSubject<boolean>(REDESIGN_TOGGLE_KEY, false), [])
    const value = useObservable(subject)
    const setValue = useCallback((value: boolean) => subject.next(value), [subject])
    return [value, setValue]
}
