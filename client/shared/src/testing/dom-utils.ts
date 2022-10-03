// Inspired by React's packages/dom-event-testing-library/domEnvironment.js
import { Driver } from './driver'

const originalPlatform = window.navigator.platform
const platformGetter = jest.spyOn(window.navigator, 'platform', 'get')

/**
 * Change environment host platform.
 */
export const platform = {
    reset() {
        platformGetter.mockReturnValue(originalPlatform)
    },
    set(name: 'mac' | 'windows') {
        switch (name) {
            case 'mac': {
                platformGetter.mockReturnValue('MacIntel')
                break
            }
            case 'windows': {
                platformGetter.mockReturnValue('Win32')
                break
            }
            default: {
                break
            }
        }
    },
}

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
