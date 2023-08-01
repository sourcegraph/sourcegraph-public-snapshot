import { faker } from '@faker-js/faker'
import { it, beforeEach, afterEach, expect, describe } from 'vitest'

import { useFakeTimers, useRealTimers } from '$mocks'

import { getRelativeTime } from './time'

const defaults = {
    Y: 2023,
    M: 5,
    D: 14,
    h: 12,
    m: 30,
    s: 30,
}

function d(options?: Partial<typeof defaults>): Date {
    const combined = { ...defaults, ...options }
    return new Date(combined.Y, combined.M, combined.D, combined.h, combined.m, combined.s)
}

describe('getRelativeTime', () => {
    beforeEach(() => {
        useFakeTimers(d())
    })

    afterEach(() => {
        useRealTimers()
    })

    it('uses the current time as reference by default', () => {
        expect(getRelativeTime(d({ h: 3 }))).toMatchInlineSnapshot('"9 hours ago"')
    })

    it('uses the provided reference date', () => {
        expect(getRelativeTime(d({ D: 5 }), d({ D: 10 }))).toMatchInlineSnapshot('"5 days ago"')
    })

    describe('specific times', () => {
        it.each([
            ['last second', d({ s: 29 })],
            ['seconds', d({ s: 5 })],
            ['last minute', d({ m: 29 })],
            ['minutes', d({ m: 5 })],
            ['last hour', d({ h: 11 })],
            ['hours', d({ h: 5 })],
            ['last day', d({ D: 13 })],
            ['days', d({ D: 2 })],
            ['last month', d({ M: 4 })],
            ['months', d({ M: 1 })],
            ['last year', d({ Y: 2022 })],
            ['years', d({ Y: 2015 })],
        ])('%s', (_, date) => {
            expect(getRelativeTime(date)).toMatchSnapshot()
        })
    })

    it('random times', () => {
        for (const date of faker.date.betweens({ from: d({ Y: 2021 }), to: d(), count: 10 })) {
            expect(getRelativeTime(date)).toMatchSnapshot()
        }
    })
})
