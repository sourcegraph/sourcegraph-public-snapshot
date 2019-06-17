import { Observable } from 'rxjs'
import { catchError, map } from 'rxjs/operators'
import { HoverAlert } from '../../../../shared/src/hover/HoverOverlay'
import { observeStorageKey, storage } from '../../browser/storage'
import { StorageItems } from '../../browser/types'
import { nativeTooltipsAlert } from './native_tooltips'

export type ExtensionHoverAlertType = 'nativeTooltips'

const getAllHoverAlerts = (codeHostName?: string): HoverAlert<ExtensionHoverAlertType>[] => [
    { type: 'nativeTooltips', content: nativeTooltipsAlert(codeHostName) },
]

/**
 * Returns an Osbervable of all hover alerts that have not yet
 * been dismissed by the user.
 */
export function getActiveHoverAlerts(codeHostName?: string): Observable<HoverAlert[] | undefined> {
    const allAlerts = getAllHoverAlerts(codeHostName)
    return observeStorageKey('sync', 'dismissedHoverAlerts').pipe(
        map(dismissedAlerts => (dismissedAlerts ? allAlerts.filter(({ type }) => !dismissedAlerts[type]) : allAlerts)),
        catchError(err => {
            console.error('Error getting hover alerts', err)
            return [undefined]
        })
    )
}
/**
 * Marks a hovewr alert as dismissed in sync storage.
 */
export async function onHoverAlertDismissed(alertType: ExtensionHoverAlertType): Promise<void> {
    try {
        const partialStorageItems: Pick<StorageItems, 'dismissedHoverAlerts'> = {
            dismissedHoverAlerts: {},
            ...(await storage.sync.get('dismissedHoverAlerts')),
        }
        partialStorageItems.dismissedHoverAlerts[alertType] = true
        await storage.sync.set(partialStorageItems)
    } catch (err) {
        console.error('Error dismissing alert', err)
    }
}
