import { useMemo } from 'react'

import { Observable } from 'rxjs'
import { catchError, debounceTime, map } from 'rxjs/operators'

import { observeResize } from '@sourcegraph/common'
import { useObservable } from '@sourcegraph/wildcard'

interface ObserveQuerySelectorInit {
    /**
     * Any valid HTML/CSS selector
     */
    selector: string
    /**
     * Timeout in milliseconds
     */
    timeout: number
    /**
     * Target element to observe for changes.
     * Default is document
     */
    target?: HTMLElement
}

class ElementNotFoundError extends Error {
    public readonly name = 'ElementNotFoundError'
    constructor({ selector, timeout: timeoutMs }: ObserveQuerySelectorInit) {
        super(`Could not find element with selector ${selector} within ${timeoutMs}ms.`)
    }
}

/**
 * Returns an observable that emits when an element that matches `selector` is found.
 * Errors out if the selector doesn't yield an element by `timeoutMs`
 */
export const observeQuerySelector = ({ selector, timeout, target }: ObserveQuerySelectorInit): Observable<Element> =>
    new Observable(function subscribe(observer) {
        const targetElement = target ?? document
        const intervalId = setInterval(() => {
            const element = targetElement.querySelector(selector)
            if (element) {
                observer.next(element)
                observer.complete()
            }
        }, Math.min(100, timeout))

        const timeoutId = setTimeout(() => {
            clearInterval(intervalId)
            // If the element still hasn't appeared, call error handler.
            observer.error(ElementNotFoundError)
        }, timeout)

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
                        // contentRect used as fallback for versions of safari that does not support borderBoxSize
                        const width = borderBoxSize?.inlineSize ?? entry?.contentRect.width

                        if (!width) {
                            return false
                        }

                        return width >= breakpoint
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
): ResizeObserverSize | undefined => (!Array.isArray(resizeObserverSize) ? resizeObserverSize : resizeObserverSize[0])

export function createElement<K extends keyof HTMLElementTagNameMap>(
    tagName: K,
    properties: Partial<HTMLElementTagNameMap[K]> | null = null,
    ...children: (Node | string)[]
): HTMLElementTagNameMap[K] {
    const element = Object.assign(document.createElement(tagName), properties)
    for (const child of children) {
        element.append(typeof child === 'string' ? document.createTextNode(child) : child)
    }
    return element
}
