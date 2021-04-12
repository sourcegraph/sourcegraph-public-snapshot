import { head, isArray } from 'lodash'
import { useMemo } from 'react'
import { Observable } from 'rxjs'
import { catchError, debounceTime, map } from 'rxjs/operators'

import { useObservable } from '@sourcegraph/shared/src/util/useObservable'

/**
 * An Observable wrapper around ResizeObserver
 */
export const observeResize = (target: HTMLElement): Observable<ResizeObserverEntry | undefined> =>
    new Observable(observer => {
        const resizeObserver = new ResizeObserver(entries => {
            observer.next(head(entries))
        })
        resizeObserver.observe(target)
        return () => resizeObserver.disconnect()
    })

/** Media breakpoints */
const breakpoints = {
    sm: 576,
    md: 768,
    lg: 992,
    xl: 1220,
} as const

export function useBreakpoint(size: keyof typeof breakpoints, debounceMs = 50): boolean {
    const breakpoint = breakpoints[size]

    return !!useObservable(
        useMemo(
            () =>
                observeResize(document.body).pipe(
                    debounceTime(debounceMs),
                    map(entry => {
                        const borderBoxSize = normalizeResizeObserverSize(entry?.borderBoxSize)

                        if (!borderBoxSize) {
                            return false
                        }

                        return borderBoxSize.inlineSize >= breakpoint
                    }),
                    // TODO: debug log.
                    // On error, be conservative and report that the screen is smaller than the breakpoint
                    catchError(() => [false])
                ),
            [breakpoint, debounceMs]
        )
    )
}

/**
 * Firefox `ResizeObserverSize`s are single objects, whereas Chrome's are wrapped in an array.
 * See: https://developer.mozilla.org/en-US/docs/Web/API/ResizeObserver#examples
 */
const normalizeResizeObserverSize = (
    resizeObserverSize: undefined | readonly ResizeObserverSize[] | ResizeObserverSize
): ResizeObserverSize | undefined => (!isArray(resizeObserverSize) ? resizeObserverSize : resizeObserverSize[0])
