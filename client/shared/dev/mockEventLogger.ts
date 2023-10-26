import { jest } from '@jest/globals'

// We do not want to fire any logs when running tests
jest.mock('../../src/tracking/eventLogger', () => ({
    eventLogger: {
        log: () => undefined,
        logViewEvent: () => undefined,
        logPageView: () => undefined,
    },
}))
