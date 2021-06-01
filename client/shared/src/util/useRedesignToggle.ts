export const REDESIGN_TOGGLE_KEY = 'isRedesignEnabled'
export const REDESIGN_CLASS_NAME = 'theme-redesign'
export const NOT_REDESIGN_CLASS_NAME = 'theme-classic'

export const getIsRedesignEnabled = (): boolean => true

/**
 * Hook to read and set the flag `isRedesignEnabled` that is persisted to localStorage
 * Used in the Web app and Storybook to toggle global CSS class - `REDESIGN_CLASS_NAME`.
 */
export const useRedesignToggle = (initialValue = false): [boolean, (value: boolean) => void] => [true, () => {}]
