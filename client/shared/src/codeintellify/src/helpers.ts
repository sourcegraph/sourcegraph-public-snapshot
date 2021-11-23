import { MaybeLoadingResult } from './loading'
import { Subscribable } from 'sourcegraph'
import { Observable, from } from 'rxjs'
import { isObject } from 'lodash'
import { map } from 'rxjs/operators'

/**
 * Checks if the given value is thenable.
 */
const isPromiseLike = (value: unknown): value is PromiseLike<unknown> =>
    isObject(value) && typeof (value as PromiseLike<unknown>).then === 'function'

/**
 * Converts a provider result, which can be a continuously-updating, maybe-loading Subscribable, or a Promise for a
 * single result, to the same type.
 */
export const toMaybeLoadingProviderResult = <T>(
    value: Subscribable<MaybeLoadingResult<T>> | PromiseLike<T>
): Observable<MaybeLoadingResult<T>> =>
    isPromiseLike(value) ? from(value).pipe(map(result => ({ isLoading: false, result }))) : from(value)

/**
 * Returns true if `val` is not `null` or `undefined`
 */
export const isDefined = <T>(val: T): val is NonNullable<T> => val !== undefined && val !== null

/**
 * Returns a function that returns `true` if the given `key` of the object is not `null` or `undefined`.
 *
 * I ❤️ TypeScript.
 */
export const propertyIsDefined = <T extends object, K extends keyof T>(key: K) => (
    val: T
): val is T & { [k in K]-?: NonNullable<T[k]> } => isDefined(val[key])

/**
 * Scrolls an element to the center if it is out of view.
 * Does nothing if the element is in view.
 *
 * @param container The scrollable container (that has `overflow: auto`)
 * @param content The content child that is being scrolled
 * @param target The element that should be scrolled into view
 */
export const scrollIntoCenterIfNeeded = (container: HTMLElement, content: HTMLElement, target: HTMLElement): void => {
    const containerRect = container.getBoundingClientRect()
    const rowRect = target.getBoundingClientRect()
    if (rowRect.top <= containerRect.top || rowRect.bottom >= containerRect.bottom) {
        const containerRect = container.getBoundingClientRect()
        const contentRect = content.getBoundingClientRect()
        const rowRect = target.getBoundingClientRect()
        const scrollTop = rowRect.top - contentRect.top - containerRect.height / 2 + rowRect.height / 2
        container.scrollTop = scrollTop
    }
}

/**
 * Returns a curried function that returns `true` if `e1` and `e2` overlap.
 */
export const elementOverlaps = (e1: HTMLElement) => (e2: HTMLElement): boolean => {
    const e1Rect = e1.getBoundingClientRect()
    const e2Rect = e2.getBoundingClientRect()
    return !(
        e1Rect.right < e2Rect.left ||
        e1Rect.left > e2Rect.right ||
        e1Rect.bottom < e2Rect.top ||
        e1Rect.top > e2Rect.bottom
    )
}
