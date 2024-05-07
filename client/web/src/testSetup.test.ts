/* eslint-disable no-console */

import { vi } from 'vitest'

import '@sourcegraph/testing/src/jestDomMatchers'

vi.mock('zustand')

// We do not want to fire any logs when running tests
vi.mock('@sourcegraph/shared/src/telemetry/web/eventLogger', () => ({
    EVENT_LOGGER: {
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
