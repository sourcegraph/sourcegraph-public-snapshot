// We do not want to fire any logs when running tests
jest.mock('@sourcegraph/web/src/tracking/eventLogger', () => ({
    eventLogger: {
        log: () => undefined,
        logViewEvent: () => undefined,
    },
}))
