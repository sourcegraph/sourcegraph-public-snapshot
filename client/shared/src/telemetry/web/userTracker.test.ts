import { describe, expect, it } from 'vitest'

import { UserTracker } from './userTracker'

describe('UserTracker', () => {
    it('initializes values without existing cookies', () => {
        let setCalls = 0
        const tracker = new UserTracker({
            get(): string | undefined {
                return undefined
            },
            set(): string | undefined {
                setCalls += 1
                return undefined
            },
        })
        expect(tracker.anonymousUserID).toBeTruthy()
        expect(tracker.cohortID).toBeTruthy()
        expect(tracker.deviceID).toBeTruthy()
        expect(tracker.deviceSessionID).toBeTruthy()
        expect(setCalls).toEqual(4)
    })
})
