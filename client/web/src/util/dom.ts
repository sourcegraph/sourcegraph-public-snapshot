import { isArray } from 'lodash'
import { useMemo } from 'react'
import { Observable } from 'rxjs'
import { catchError, debounceTime, map } from 'rxjs/operators'

import { observeResize } from '@sourcegraph/shared/src/util/dom'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'

interface ObserveQuerySelectorInit {
    selector: string
    timeoutMs: number
    target?: HTMLElement
}

class ElementNotFoundError extends Error {
    public readonly name = 'ElementNotFoundError'
    constructor({ selector, timeoutMs }: ObserveQuerySelectorInit) {
        super(`Could not find element with selector ${selector} within ${timeoutMs}ms.`)
    }
}

/**
 * Returns an observable that emits when an element that matches `selector` is found.
 * Errors out if the selector doesn't yield an element by `timeoutMs`
 */
export const observeQuerySelector = ({ selector, timeoutMs, target }: ObserveQuerySelectorInit): Observable<Element> =>
    new Observable(function subscribe(observer) {
        const targetElement = target ?? document
        const intervalId = setInterval(
            () => {
                const element = targetElement.querySelector(selector)
                if (element) {
                    observer.next(element)
                    observer.complete()
                }
            },
            timeoutMs > 100 ? 100 : timeoutMs
        )
        const timeoutId = setTimeout(() => {
            clearInterval(intervalId)
            // If the element still hasn't appeared, call error handler.
            observer.error(ElementNotFoundError)
        }, timeoutMs)

        return function unsubscribe() {
            clearTimeout(timeoutId)
            clearInterval(intervalId)
        }
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
