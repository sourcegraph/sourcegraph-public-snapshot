import { vi } from 'vitest'

// We do not want to fire any logs when running tests
vi.mock('../../src/tracking/eventLogger', () => ({
    eventLogger: {
        log: () => undefined,
        logViewEvent: () => undefined,
        logPageView: () => undefined,
    },
}))
