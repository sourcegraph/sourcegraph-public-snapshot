import { Observable } from 'rxjs'

export interface ObserveQuerySelectorInit {
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
