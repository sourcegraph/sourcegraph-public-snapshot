import { Observable, of } from 'rxjs'
import { catchError, map, startWith, switchMap } from 'rxjs/operators'
import { HoverAlert } from '../../../../../shared/src/hover/HoverOverlay'
import { combineLatestOrDefault } from '../../../../../shared/src/util/rxjs/combineLatestOrDefault'
import { observeStorageKey, storage } from '../../../browser-extension/web-extension-api/storage'
import { SyncStorageItems } from '../../../browser-extension/web-extension-api/types'
import { isInPage } from '../../context'

export type ExtensionHoverAlertType = 'nativeTooltips'

/**
 * Returns an Observable of all hover alerts that have not yet
 * been dismissed by the user.
 */
export function getActiveHoverAlerts(
    allAlerts: Observable<HoverAlert<ExtensionHoverAlertType>>[]
): Observable<HoverAlert<ExtensionHoverAlertType>[] | undefined> {
    if (isInPage) {
        return of(undefined)
    }
    return observeStorageKey('sync', 'dismissedHoverAlerts').pipe(
        switchMap(dismissedAlerts =>
            combineLatestOrDefault(allAlerts).pipe(
                map(alerts => (dismissedAlerts ? alerts.filter(({ type }) => !dismissedAlerts[type]) : alerts))
            )
        ),
        catchError(error => {
            console.error('Error getting hover alerts', error)
            return [undefined]
        }),
        startWith([])
    )
}
/**
 * Marks a hover alert as dismissed in sync storage.
 */
export async function onHoverAlertDismissed(alertType: ExtensionHoverAlertType): Promise<void> {
    try {
        const partialStorageItems: Pick<SyncStorageItems, 'dismissedHoverAlerts'> = {
            dismissedHoverAlerts: {},
            ...(await storage.sync.get('dismissedHoverAlerts')),
        }
        partialStorageItems.dismissedHoverAlerts[alertType] = true
        await storage.sync.set(partialStorageItems)
    } catch (error) {
        console.error('Error dismissing alert', error)
    }
}
