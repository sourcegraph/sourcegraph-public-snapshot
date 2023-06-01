import { isSiteVersionSupported } from './protocol'

describe('isSiteVersionSupported', () => {
    test('returns false for invalid input', () => {
        expect(isSiteVersionSupported(new Error('instance version not found'))).toBe(false)
        expect(isSiteVersionSupported(new Error(undefined))).toBe(false)
        expect(isSiteVersionSupported('')).toBe(false)
        expect(isSiteVersionSupported('foo')).toBe(false)
    })

    test('returns true for version 5.1.0 and above', () => {
        expect(isSiteVersionSupported('5.1.0')).toBe(true)
        expect(isSiteVersionSupported('5.2.0')).toBe(true)
        expect(isSiteVersionSupported('6.0.0')).toBe(true)
    })

    test('returns false for version below 5.1.0', () => {
        expect(isSiteVersionSupported('5.0.9')).toBe(false)
        expect(isSiteVersionSupported('4.9.9')).toBe(false)
        expect(isSiteVersionSupported('0.0.0')).toBe(false)
        expect(isSiteVersionSupported('3.20.19')).toBe(false)
    })

    test('returns true for all insider build version', () => {
        expect(isSiteVersionSupported('222587_2023-06-30_5.2-39cbcf1a50f0')).toBe(true)
        expect(isSiteVersionSupported('222587_2023-05-30_5.0-39cbcf1a50f0')).toBe(true)
        expect(isSiteVersionSupported('222587_2023-05-30_4.9-39cbcf1a50f0')).toBe(true)
    })
})
