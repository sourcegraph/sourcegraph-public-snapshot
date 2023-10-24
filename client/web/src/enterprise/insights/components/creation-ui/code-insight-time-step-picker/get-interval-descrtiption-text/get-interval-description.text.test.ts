import { describe, expect, it } from '@jest/globals'

import { formatDuration } from './get-interval-description-text'

describe('formatDuration should work properly ', () => {
    it('with minutes', () => {
        expect(formatDuration({ minutes: 60 })).toStrictEqual('1 hour')
    })

    it('with hours', () => {
        expect(formatDuration({ hours: 24 })).toStrictEqual('1 day')
    })

    it('with hours and days', () => {
        expect(formatDuration({ hours: 36 })).toStrictEqual('1 day and 12 hours')
    })

    it('with days', () => {
        expect(formatDuration({ days: 7 })).toStrictEqual('1 week')
    })

    it('with days and week', () => {
        expect(formatDuration({ days: 10 })).toStrictEqual('1 week and 3 days')
    })

    it('with weeks', () => {
        expect(formatDuration({ weeks: 5 })).toStrictEqual('1 month')
    })

    it('with weeks and months', () => {
        expect(formatDuration({ weeks: 6 })).toStrictEqual('1 month and 1 week')
    })

    it('with mounts', () => {
        expect(formatDuration({ months: 36 })).toStrictEqual('3 years')
    })

    it('with mounts and years', () => {
        expect(formatDuration({ months: 40 })).toStrictEqual('3 years and 4 months')
    })
})
