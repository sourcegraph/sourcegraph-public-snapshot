import { describe, expect, it } from 'vitest'

import { humanizeSize } from './size'

describe('humanizeSize', () => {
    it('returns byte size for values under 1KB', () => {
        expect(humanizeSize(500)).toBe('500B')
    })

    it('returns KB size for values between 1KB and 1MB', () => {
        expect(humanizeSize(12345)).toBe('12.35KB')
    })

    it('returns MB size for values between 1MB and 1GB', () => {
        expect(humanizeSize(123456789)).toBe('123.46MB')
    })

    it('returns GB size for values above 1GB', () => {
        expect(humanizeSize(123456789023)).toBe('123.46GB')
    })

    it('returns TB size for values above 1TB', () => {
        expect(humanizeSize(1234567890123)).toBe('1.23TB')
    })

    it('rounds decimal values to 2 places', () => {
        expect(humanizeSize(1234567)).toBe('1.23MB')
    })
})
