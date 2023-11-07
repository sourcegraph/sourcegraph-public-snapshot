import { describe, expect, it } from 'vitest'

import { formatDurationLong } from './time'

describe('formatDurationLong', () => {
    it('should format durations as per the spec', () => {
        expect(formatDurationLong(0)).toEqual('0 milliseconds')
        expect(formatDurationLong(100)).toEqual('100 milliseconds')
        expect(formatDurationLong(1000)).toEqual('1 second')
        expect(formatDurationLong(10000)).toEqual('10 seconds')
        expect(formatDurationLong(100000)).toEqual('1 minute and 40 seconds')
        expect(formatDurationLong(1000000)).toEqual('16 minutes and 40 seconds')
        expect(formatDurationLong(10000000)).toEqual('2 hours and 46 minutes')
        expect(formatDurationLong(100000000)).toEqual('1 day and 3 hours')
    })
})
