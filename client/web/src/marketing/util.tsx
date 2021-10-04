import { DAYS_ACTIVE_STORAGE_KEY, LAST_DAY_ACTIVE_STORAGE_KEY } from './constants'

export const IS_CHROME = !!window.chrome

export const getDaysActiveCount = (): number => parseInt(localStorage.getItem(DAYS_ACTIVE_STORAGE_KEY) || '', 10) || 0

/**
 * Function called upon initial app load that checks if the user is visiting
 * on a new day, and if so, updates persistent local storage with that information
 */
export function updateUserSessionStores(): void {
    if (new Date().toDateString() !== localStorage.getItem(LAST_DAY_ACTIVE_STORAGE_KEY)) {
        const daysActiveCount = getDaysActiveCount() + 1
        localStorage.setItem(DAYS_ACTIVE_STORAGE_KEY, daysActiveCount.toString())
        localStorage.setItem(LAST_DAY_ACTIVE_STORAGE_KEY, new Date().toDateString())
    }
}
