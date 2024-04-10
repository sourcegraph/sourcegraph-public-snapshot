import type { MouseEventHandler } from 'react'

import { type HistoryOrNavigate, compatNavigate } from '@sourcegraph/common'

/**
 * Returns a click handler for link element that will make sure clicks on in-app links are handled on the client
 * and don't cause a full page reload.
 */
export const createLinkClickHandler =
    (history: HistoryOrNavigate): MouseEventHandler<unknown> =>
    event => {
        // Do nothing if the link was requested to open in a new tab
        if (event.ctrlKey || event.metaKey) {
            return
        }

        // Check if click happened within an anchor inside the markdown
        const anchor = event.nativeEvent
            .composedPath()
            .slice(0, event.nativeEvent.composedPath().indexOf(event.currentTarget) + 1)
            .find(anyOf(isInstanceOf(HTMLAnchorElement), isInstanceOf(SVGAElement)))
        if (!anchor) {
            return
        }
        const href = typeof anchor.href === 'string' ? anchor.href : anchor.href.baseVal

        // Check if URL is outside the app
        if (isExternalLink(href)) {
            return
        }

        // Handle navigation programmatically
        event.preventDefault()
        const url = new URL(href)
        compatNavigate(history, url.pathname + url.search + url.hash)
    }

/**
 * Returns a function that returns `true` if the given value is an instance of the given class.
 * @param constructor A reference to a class, e.g. `HTMLElement`
 */
const isInstanceOf =
    <C extends new () => object>(constructor: C) =>
        (value: unknown): value is InstanceType<C> =>
            value instanceof constructor

/**
 * Combines multiple type guards into one type guard that checks if the value passes any of the provided type guards.
 */
function anyOf<T0, T1 extends T0, T2 extends Exclude<T0, T1>>(
    t1: (value: T0) => value is T1,
    t2: (value: Exclude<T0, T1>) => value is T2
): (value: T0) => value is T1 | T2
function anyOf(...typeGuards: any[]): any {
    return (value: unknown) => typeGuards.some((guard: (value: unknown) => boolean) => guard(value))
}

/**
 * Returns true if the given URL points outside the current site.
 */
const isExternalLink = (
    url: string,
    windowLocation__testingOnly: Pick<URL, 'origin' | 'href'> = window.location
): boolean =>
    !!tryCatch(() => new URL(url, windowLocation__testingOnly.href).origin !== windowLocation__testingOnly.origin)

/**
 * Run the passed function and return `undefined` if it throws an error.
 */
function tryCatch<T>(function_: () => T): T | undefined {
    try {
        return function_()
    } catch {
        return undefined
    }
}
