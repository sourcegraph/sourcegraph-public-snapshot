import { useCallback, useMemo } from 'react'

import { LocalStorageSubject } from './LocalStorageSubject'
import { useObservable } from './useObservable'

export const REDESIGN_TOGGLE_KEY = 'isRedesignEnabled'
export const REDESIGN_CLASS_NAME = 'theme-redesign'

export const getIsRedesignEnabled = (): boolean => localStorage.getItem(REDESIGN_TOGGLE_KEY) === 'true'

export const toggleIsRedesignEnabled = (isEnabled: boolean, element: HTMLElement = document.body) => {
    element.classList.toggle(REDESIGN_CLASS_NAME, isEnabled)
    localStorage.setItem(REDESIGN_TOGGLE_KEY, String(isEnabled))
    window.dispatchEvent(new StorageEvent('storage', { key: REDESIGN_TOGGLE_KEY, newValue: String(isEnabled) }))
}

/**
 * Hook to read and set the flag `isRedesignEnabled` that is persisted to localStorage
 * Used in the Web app and Storybook to toggle global CSS class - `REDESIGN_CLASS_NAME`.
 */
export const useRedesignToggle = (initialValue = false): [boolean, (value: boolean) => void] => {
    const subject = useMemo(() => new LocalStorageSubject<boolean>(REDESIGN_TOGGLE_KEY, initialValue), [initialValue])
    const value = useObservable(subject) as boolean // Since subject has an initial value, the observable should never be undefined
    const setValue = useCallback((value: boolean) => subject.next(value), [subject])
    return [value, setValue]
}
