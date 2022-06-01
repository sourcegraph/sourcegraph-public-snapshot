import { useCallback } from 'react'

import copy from 'copy-to-clipboard'
import { merge, Observable, of } from 'rxjs'
import { delay, startWith, switchMapTo, tap } from 'rxjs/operators'

import { TooltipController, useEventObservable } from '@sourcegraph/wildcard'

type URLValue = string | undefined
type useCopiedHandlerReturn = [(value?: URLValue) => void, boolean | undefined]

/**
 * Provide logic for copy dashboard URL logic.
 * Returns handler to copy a dashboard URL and as the second parameter copied tooltip state
 */
export function useCopyURLHandler(): useCopiedHandlerReturn {
    const copyDashboardURL = useCallback((linkURL?: URLValue): void => {
        copy(linkURL ?? window.location.href)
    }, [])

    return useEventObservable<URLValue, boolean>(
        useCallback(
            (clicks: Observable<URLValue>) =>
                clicks.pipe(
                    tap(copyDashboardURL),
                    switchMapTo(merge(of(true), of(false).pipe(delay(2000)))),
                    tap(() => TooltipController.forceUpdate()),
                    startWith(false)
                ),
            [copyDashboardURL]
        )
    )
}
