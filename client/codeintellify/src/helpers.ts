import { isObject } from 'lodash'
import { Observable, from } from 'rxjs'
import { map } from 'rxjs/operators'
import { Subscribable } from 'sourcegraph'

import { isDefined } from '@sourcegraph/common'

import { MaybeLoadingResult } from './loading'

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
 * Returns a function that returns `true` if the given `key` of the object is not `null` or `undefined`.
 *
 * I ❤️ TypeScript.
 */
export const propertyIsDefined = <T extends object, K extends keyof T>(key: K) => (
    value: T
): value is T & { [k in K]-?: NonNullable<T[k]> } => isDefined(value[key])

/**
 * Scrolls an element to the center if it is out of view.
 * Does nothing if the element is in view.
 *
 * @param container The scrollable container (that has `overflow: auto`)
 * @param content The content child that is being scrolled
 * @param target The element that should be scrolled into view
 */
export const scrollIntoCenterIfNeeded = (container: HTMLElement, content: HTMLElement, target: HTMLElement): void => {
    const containerRectangle = container.getBoundingClientRect()
    const rowRectangle = target.getBoundingClientRect()
    if (rowRectangle.top <= containerRectangle.top || rowRectangle.bottom >= containerRectangle.bottom) {
        const containerRectangle = container.getBoundingClientRect()
        const contentRectangle = content.getBoundingClientRect()
        const rowRectangle_ = target.getBoundingClientRect()
        const scrollTop =
            rowRectangle_.top - contentRectangle.top - containerRectangle.height / 2 + rowRectangle_.height / 2
        container.scrollTop = scrollTop
    }
}

/**
 * Returns a curried function that returns `true` if `e1` and `e2` overlap.
 */
export const elementOverlaps = (element1: HTMLElement) => (element2: HTMLElement): boolean => {
    const rectangle1 = element1.getBoundingClientRect()
    const rectangle2 = element2.getBoundingClientRect()
    return !(
        rectangle1.right < rectangle2.left ||
        rectangle1.left > rectangle2.right ||
        rectangle1.bottom < rectangle2.top ||
        rectangle1.top > rectangle2.bottom
    )
}
