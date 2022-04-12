import { Observable, of } from 'rxjs'
import { catchError, map, startWith, switchMap } from 'rxjs/operators'
import type { HoverAlert } from 'sourcegraph'

import { combineLatestOrDefault } from '@sourcegraph/common'
import { MarkupKind } from '@sourcegraph/extension-api-classes'

import { observeStorageKey, storage } from '../../../browser-extension/web-extension-api/storage'
import { SyncStorageItems } from '../../../browser-extension/web-extension-api/types'
import { isInPage } from '../../context'

import { CodeHost } from './codeHost'

/**
 * Returns an Observable of all hover alerts that have not yet
 * been dismissed by the user.
 */
export function getActiveHoverAlerts(allAlerts: Observable<HoverAlert>[]): Observable<HoverAlert[] | undefined> {
    if (isInPage) {
        return of(undefined)
    }
    return observeStorageKey('sync', 'dismissedHoverAlerts').pipe(
        switchMap(dismissedAlerts =>
            combineLatestOrDefault(allAlerts).pipe(
                map(alerts => (dismissedAlerts ? alerts.filter(({ type }) => !type || !dismissedAlerts[type]) : alerts))
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
export async function onHoverAlertDismissed(alertType: string): Promise<void> {
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

/**
 * Returns the alert to show when the user is on private code (that has not
 * been added to Cloud) and has sourcegraph.com as the URL.
 * The alert informs the user to setup a private Sourcegraph instance.
 */
export const createPrivateCodeHoverAlert = (codeHost: Pick<CodeHost, 'hoverOverlayClassProps'>): HoverAlert => ({
    type: 'private-code',
    summary: {
        kind: MarkupKind.Markdown,
        value:
            '#### Sourcegraph for private code\n\n' +
            'To get Sourcegraph hovers on your private repositories, you need to set up a private Sourcegraph instance and connect it to the browser extension.' +
            '\n\n' +
            `<a href="https://docs.sourcegraph.com/code_search/how-to/adding_repositories_to_cloud" class="${
                codeHost.hoverOverlayClassProps?.actionItemClassName ?? ''
            }" target="_blank" rel="noopener norefferer">Show more info</a>`,
    },
})
