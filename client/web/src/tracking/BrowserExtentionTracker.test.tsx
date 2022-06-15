import React from 'react'

import { act, cleanup, render } from '@testing-library/react'
import { renderHook, cleanup as hookCleanup } from '@testing-library/react-hooks'
import { MemoryRouter } from 'react-router-dom'

import { BrowserExtensionTracker, useIsBrowserExtensionActiveUser } from './BrowserExtensionTracker'

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
        [
            'https://sourcegraph.com/github.com/sourcegraph/sourcegraph?utm_source=chrome-extension&utm_campaign=view-on-sourcegraph',
            DATE_NOW,
        ],
        [
            'https://sourcegraph.com/github.com/sourcegraph/sourcegraph?utm_source=firefox-extension&utm_campaign=view-on-sourcegraph',
            DATE_NOW,
        ],
        [
            'https://sourcegraph.com/github.com/sourcegraph/sourcegraph?utm_source=safari-extension&utm_campaign=view-on-sourcegraph',
            DATE_NOW,
        ],
        ['https://sourcegraph.com/?something=different', null],
    ]
    test.each(cases)('Detects query parameters for %p', (url, expectedResult) => {
        expect(localStorage.getItem(BROWSER_EXTENSION_LAST_DETECTION_KEY)).toBeNull()

        render(
            <MemoryRouter initialEntries={[url]}>
                <BrowserExtensionTracker />
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
                <BrowserExtensionTracker />
            </MemoryRouter>,
            { wrapper }
        )
        await act(() => new Promise(resolve => setTimeout(resolve, 150)))
        expect(localStorage.getItem(BROWSER_EXTENSION_LAST_DETECTION_KEY)).toBeTruthy()
        jest.useFakeTimers()
    })
})

describe('useIsBrowserExtensionActiveUser', () => {
    afterAll(hookCleanup)

    afterEach(() => {
        localStorage.clear()
    })

    test('Returns falsy', async () => {
        const { result, waitForNextUpdate } = renderHook(() => useIsBrowserExtensionActiveUser())
        expect(result.current).toBeUndefined()
        await waitForNextUpdate()
        expect(result.current).toBeFalsy()
    })

    test('Returns truthy if "localStorage" item exist', () => {
        localStorage.setItem(BROWSER_EXTENSION_LAST_DETECTION_KEY, `${Date.now()}`)
        const { result } = renderHook(() => useIsBrowserExtensionActiveUser())
        expect(result.current).toBeTruthy()
    })

    test('Returns truthy if extension marker DOM element exist', async () => {
        jest.spyOn(document, 'querySelector').mockImplementation(selector =>
            selector === `#${BROWSER_EXTENSION_MARKER_ELEMENT}` ? (document.createElement('div') as Element) : null
        )

        const { result, waitForNextUpdate } = renderHook(() => useIsBrowserExtensionActiveUser())
        expect(result.current).toBeUndefined()

        await waitForNextUpdate()

        expect(result.current).toBeTruthy()

        jest.resetAllMocks()
    })
})
