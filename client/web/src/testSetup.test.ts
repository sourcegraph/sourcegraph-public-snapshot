/* eslint-disable no-console */

// Use @testing-library/jest-dom in vitest. See
// https://github.com/testing-library/jest-dom/issues/439#issuecomment-1536524120.
import type { TestingLibraryMatchers } from '@testing-library/jest-dom/matchers'
import * as matchers from '@testing-library/jest-dom/matchers'
import { expect, vi } from 'vitest'

declare module 'vitest' {
    interface Assertion<T = any> extends jest.Matchers<void, T>, TestingLibraryMatchers<T, void> {}
}
expect.extend(matchers)

vi.mock('zustand')

// We do not want to fire any logs when running tests
vi.mock('../../src/tracking/eventLogger', () => ({
    eventLogger: {
        log: () => undefined,
        logViewEvent: () => undefined,
        logPageView: () => undefined,
    },
}))

// Suppresses DOMPurify huge console.warn messages. (It checks the value of each of these and logs
// that message if one doesn't exist.)
const oldConsoleWarn = console.warn
console.warn = (...args: any[]): void => {
    if (args[0] === 'fallback value for') {
        return
    }
    oldConsoleWarn(...args)
}
