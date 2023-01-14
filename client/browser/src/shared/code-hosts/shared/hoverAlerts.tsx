import { Observable, of } from 'rxjs'
import { catchError, map, startWith, switchMap } from 'rxjs/operators'

import { combineLatestOrDefault } from '@sourcegraph/common'
import { MarkupKind } from '@sourcegraph/extension-api-classes'
import type { HoverAlert } from '@sourcegraph/shared/src/codeintel/legacy-extensions/api'
import { ButtonLink } from '@sourcegraph/wildcard'

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
 * Returns the alert to show when the user is on an unindexed repo and does not
 * have sourcegraph.com as the URL. The alert informs the user to setup add a
 * repo.
 */
export const createRepoNotFoundHoverAlert = (codeHost: Pick<CodeHost, 'hoverOverlayClassProps'>): HoverAlert => ({
    type: 'private-code',
    buttons: [
        <ButtonLink
            key="learn_more"
            href="/help/admin/repo/add"
            className={codeHost.hoverOverlayClassProps?.actionItemClassName ?? ''}
            target="_blank"
            rel="noopener norefferer"
        >
            Learn more
        </ButtonLink>,
    ],
    summary: {
        kind: MarkupKind.Markdown,
        value:
            '#### Repository not added\n\n' +
            'This repository is not indexed by your Sourcegraph instance. Add the repository to get Code Intelligence overlays.',
    },
})
