import React from 'react'

import { act, cleanup, render } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { CompatRouter } from 'react-router-dom-v5-compat'

import { BrowserExtensionTracker } from './BrowserExtensionTracker'

const BROWSER_EXTENSION_LAST_DETECTION_KEY = 'integrations.browser.lastDetectionTimestamp'
const BROWSER_EXTENSION_MARKER_ELEMENT = 'sourcegraph-app-background'
describe('BrowserExtensionTracker', () => {
    const DATE_NOW = '1646922320064'

    afterAll(cleanup)

    beforeEach(() => {
        jest.useFakeTimers()
        jest.setSystemTime(Number(DATE_NOW))
    })

    afterEach(() => {
        jest.runOnlyPendingTimers()
        jest.useRealTimers()
        localStorage.clear()
    })

    const cases: [string, string | null][] = [
        ['/github.com/sourcegraph/sourcegraph?utm_source=chrome-extension&utm_campaign=view-on-sourcegraph', DATE_NOW],
        ['/github.com/sourcegraph/sourcegraph?utm_source=firefox-extension&utm_campaign=view-on-sourcegraph', DATE_NOW],
        ['/github.com/sourcegraph/sourcegraph?utm_source=safari-extension&utm_campaign=view-on-sourcegraph', DATE_NOW],
        ['/?something=different', null],
    ]
    test.each(cases)('Detects query parameters for %p', (url, expectedResult) => {
        expect(localStorage.getItem(BROWSER_EXTENSION_LAST_DETECTION_KEY)).toBeNull()

        render(
            <MemoryRouter initialEntries={[url]}>
                <CompatRouter>
                    <BrowserExtensionTracker />
                </CompatRouter>
            </MemoryRouter>
        )

        expect(localStorage.getItem(BROWSER_EXTENSION_LAST_DETECTION_KEY)).toEqual(expectedResult)
    })

    test('Detects extension marker DOM element', async () => {
        jest.runOnlyPendingTimers()
        jest.useRealTimers()
        expect(localStorage.getItem(BROWSER_EXTENSION_LAST_DETECTION_KEY)).toBeNull()
        const wrapper: React.FunctionComponent<React.PropsWithChildren<unknown>> = ({ children }) => (
            <div>
                {children}
                <div id={BROWSER_EXTENSION_MARKER_ELEMENT} />
            </div>
        )
        render(
            <MemoryRouter>
                <CompatRouter>
                    <BrowserExtensionTracker />
                </CompatRouter>
            </MemoryRouter>,
            { wrapper }
        )
        await act(() => new Promise(resolve => setTimeout(resolve, 150)))
        expect(localStorage.getItem(BROWSER_EXTENSION_LAST_DETECTION_KEY)).toBeTruthy()
        jest.useFakeTimers()
    })
})
