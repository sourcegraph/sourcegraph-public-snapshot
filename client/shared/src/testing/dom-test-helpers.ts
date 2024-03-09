import type { Driver } from './driver'

/**
 * Checks that given element has page focus.
 *
 * @param driver - testing page driver
 * @param selector - selector for looking an element that should have focus
 */
export async function hasFocus(driver: Driver, selector: string): Promise<boolean> {
    return driver.page.evaluate(selector => {
        const element = document.querySelector<HTMLElement>(selector)
        const focusedElement = document.activeElement

        return element === focusedElement
    }, selector)
}

/**
 * Stub {@link https://developer.mozilla.org/en-US/docs/Web/API/Element/scrollIntoView | Element.prototype.scrollIntoView}
 * for the duration of the test.
 *
 * @summary
 * `jsdom` does not implement `Element.prototype.scrollIntoView`. This may break some tests which render elements
 * that scroll into view on certain events. Stubbing it seems to be the commonly accepted workaround.
 *
 * Use this function to setup tests for such components; `stupScrollIntoview` can be passed to both `beforeAll` and `beforeEach` hooks
 * and *will require no explicit call* to `afterAll`/`afterEach`, as it returns a cleanup function to be run on teardown.
 *
 * @see https://github.com/jsdom/jsdom/issues/1695 and https://github.com/jsdom/jsdom/pull/3639
 *
 * @example
 * describe('ScrollableComponentWithChildren', () => {
 *  beforeAll(stubScrollIntoView)
 * })
 *
 * @returns Cleanup function to restore the original `Element.prototype.scrollIntoView` method.
 */
export function stubScrollIntoView(): () => void {
    const orig = Element.prototype.scrollIntoView
    Element.prototype.scrollIntoView = () => {}
    return () => {
        Element.prototype.scrollIntoView = orig
    }
}
