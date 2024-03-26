import { expect, describe, it, vi } from 'vitest'
import { evaluateKey } from './Hotkey'

const mocks = vi.hoisted(() => ({
    isLinuxPlatform: vi.fn(),
    isWindowsPlatform: vi.fn(),
    isMacPlatform: vi.fn(),
}))
vi.mock('$lib/common', () => ({
    isLinuxPlatform: mocks.isLinuxPlatform,
    isWindowsPlatform: mocks.isWindowsPlatform,
    isMacPlatform: mocks.isMacPlatform,
}))

describe('Hotkey', () => {

    describe('evaluateKey', () => {
        it('should return default key', () => {
            const actual = evaluateKey({
                key: 'hello',
            })
            expect(actual).toBe('hello')
        })

        it('should return mac key', () => {
            mocks.isMacPlatform.mockImplementationOnce(() => true)

            const actual = evaluateKey({
                key: 'hello',
                mac: 'itsmac',
            })
            expect(actual).toBe('itsmac')
        })

        it('should return windows key', () => {
            mocks.isWindowsPlatform.mockImplementationOnce(() => true)

            const actual = evaluateKey({
                key: 'hello',
                windows: 'itswindows',
            })
            expect(actual).toBe('itswindows')
        })

        it('should return linux key', () => {
            mocks.isLinuxPlatform.mockImplementationOnce(() => true)

            const actual = evaluateKey({
                key: 'hello',
                linux: 'itslinux',
            })
            expect(actual).toBe('itslinux')
        })

        it('should not return key for wrong system', () => {
            mocks.isLinuxPlatform.mockImplementationOnce(() => false)
            mocks.isMacPlatform.mockImplementationOnce(() => true)

            const actual = evaluateKey({
                key: 'hello',
                linux: 'itslinux',
            })
            expect(actual).toBe('hello')
        })
    })
})
