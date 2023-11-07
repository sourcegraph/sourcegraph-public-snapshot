import { describe, expect, test } from 'vitest'

import { formatDays, formatRate } from './format'

describe('format', () => {
    describe('formatRate', () => {
        const testCases: {
            rate: string | number
            result: string
            name: string
        }[] = [
            { rate: 0, result: 'None', name: 'is a number' },
            { rate: '0/seconds', result: 'None', name: 'is zero changesets per seconds' },
            { rate: '3/hour', result: '3 changesets per hour', name: 'is 3 changesets per hour' },
            { rate: '10/minute', result: '10 changesets per minute', name: 'is 10 changesets per minute' },
        ]

        test.each(testCases)('rate $name', ({ rate, result }) => {
            expect(formatRate(rate)).toEqual(result)
        })
    })

    describe('formatDays', () => {
        const testCases: {
            name: string
            days: string[] | undefined
            result: string
        }[] = [
            { days: undefined, result: 'every other day', name: 'undefined' },
            { days: [], result: 'every other day', name: 'empty array' },
            { days: ['monday'], result: 'Monday', name: 'single day' },
            { days: ['tuesday', 'friday'], result: 'Tuesday, Friday', name: 'multiple days' },
        ]

        test.each(testCases)('argument is $name', ({ days, result }) => {
            expect(formatDays(days)).toEqual(result)
        })
    })
})
