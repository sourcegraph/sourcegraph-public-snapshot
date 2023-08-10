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
