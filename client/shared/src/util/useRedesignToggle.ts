import { useLocalStorage } from '@sourcegraph/shared/src/util/useLocalStorage'

export const REDESIGN_TOGGLE_KEY = 'isRedesignEnabled'
export const REDESIGN_CLASS_NAME = 'theme-redesign'

export const getIsRedesignEnabled = () => {
    return localStorage.getItem(REDESIGN_TOGGLE_KEY) === 'true'
}

interface UseRedesignToggleReturn {
    isRedesignEnabled: boolean
    setIsRedesignEnabled: (isRedesignEnabled: boolean) => void
}

/**
 * Hook to read and set the flag `isRedesignEnabled` that is persisted to localStorage
 * Used in the Web app and Storybook to toggle global CSS class - `REDESIGN_CLASS_NAME`.
 */
export const useRedesignToggle = (initialValue = false): UseRedesignToggleReturn => {
    const [isRedesignEnabled, setIsRedesignEnabled] = useLocalStorage(REDESIGN_TOGGLE_KEY, initialValue)

    return {
        isRedesignEnabled,
        setIsRedesignEnabled,
    }
}
