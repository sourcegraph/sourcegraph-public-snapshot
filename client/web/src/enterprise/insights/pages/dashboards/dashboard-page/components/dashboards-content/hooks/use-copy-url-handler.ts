import copy from 'copy-to-clipboard'
import { useCallback } from 'react'
import { merge, Observable, of } from 'rxjs'
import { delay, startWith, switchMapTo, tap } from 'rxjs/operators'

import { useEventObservable } from '@sourcegraph/shared/src/util/useObservable'
import { TooltipController } from '@sourcegraph/wildcard'

type useCopiedHandlerReturn = [() => void, boolean | undefined]

/**
 * Provide logic for copy dashboard URL logic.
 * Returns handler to copy a dashboard URL and as the second parameter copied tooltip state
 */
export function useCopyURLHandler(): useCopiedHandlerReturn {
    const copyDashboardURL = useCallback((): void => {
        copy(window.location.href)
    }, [])

    return useEventObservable(
        useCallback(
            (clicks: Observable<void>) =>
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
