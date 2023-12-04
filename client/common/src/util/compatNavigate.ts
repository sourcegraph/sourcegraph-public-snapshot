import type { History } from 'history'
import type { NavigateFunction, NavigateOptions, To } from 'react-router-dom'

export type HistoryOrNavigate = History | NavigateFunction

/**
 * Compatibility layer between react-router@6 `NavigateFunction` and `history.push`.
 * Exposes the `NavigateFunction` API that we can use during the migration. On migration
 * completion we can find-and-replace this helper with the `NavigateFunction` call
 *
 * During the migration;
 * ```ts
 * function helper(historyOrNavigate: HistoryOrNavigate) {
 *     const { url, state } = getNewLocationInfo()
 *
 *     compatNavigate(history, url, { state })
 * }
 *
 * ```
 *
 * On migration completion;
 * ```ts
 * function helper(navigate: NavigateFunction) {
 *     const { url, state } = getNewLocationInfo()
 *
 *     navigate(url, { state })
 * }
 * ```
 */
export function compatNavigate(historyOrNavigate: HistoryOrNavigate, to: To, options?: NavigateOptions): void {
    if (typeof historyOrNavigate === 'function') {
        // Use react-router to handle in-app navigation.
        historyOrNavigate(to, options)
    } else {
        // Use legacy `history.push` to change the location.
        historyOrNavigate.push(to, options?.state)
    }
}
