import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'
import { HoverAlert } from '../../../../shared/src/hover/HoverOverlay'
import { observeStorageKey, storage } from '../../browser/storage'
import { nativeTooltipsAlert } from './native_tooltips'

const getAllHoverAlerts = (codeHostName: string): HoverAlert[] => [
    { type: 'nativeTooltips', content: nativeTooltipsAlert(codeHostName) },
]

/**
 * Returns an Osbervable of all hover alerts that have not yet
 * been dismissed by the user.
 */
export function getActiveHoverAlerts(codeHostName: string): Observable<HoverAlert[]> {
    const allAlerts = getAllHoverAlerts(codeHostName)
    return observeStorageKey('sync', 'dismissedHoverAlerts').pipe(
        map(dismissedAlerts => (dismissedAlerts ? allAlerts.filter(({ type }) => !dismissedAlerts[type]) : allAlerts))
    )
}
/**
 * Marks a hovewr alert as dismissed in sync storage.
 */
export async function onHoverAlertDismissed(alertType: string): Promise<void> {
    try {
        const partialStorageItems = {
            dismissedHoverAlerts: {},
            ...(await storage.sync.get('dismissedHoverAlerts')),
        }
        partialStorageItems.dismissedHoverAlerts[alertType] = true
        await storage.sync.set(partialStorageItems)
    } catch (err) {
        console.error('Error dismissing alert', err)
    }
}
