import type hotkeys from 'hotkeys-js'
import { expect, describe, it, vi, afterEach } from 'vitest'

import { evaluateKey, registerHotkey } from './Hotkey'

const mocks = vi.hoisted(() => ({
    isLinuxPlatform: vi.fn(),
    isWindowsPlatform: vi.fn(),
    isMacPlatform: vi.fn(),
    isElementMock: vi.fn(),
    getAllKeyCodes: vi.fn(),
    unbind: vi.fn(),
}))
vi.mock('$lib/common', () => ({
    isLinuxPlatform: mocks.isLinuxPlatform,
    isWindowsPlatform: mocks.isWindowsPlatform,
    isMacPlatform: mocks.isMacPlatform,
}))

vi.mock('hotkeys-js', async importOriginal => {
    const originalModule = await importOriginal<typeof import('hotkeys-js')>()

    // Create a mock function for the default export
    const defaultExportMock = vi.fn() as unknown as typeof hotkeys

    // Attach additional properties to the mock function, mimicking the structure of hotkeys-js
    defaultExportMock.getAllKeyCodes = mocks.getAllKeyCodes
    defaultExportMock.unbind = mocks.unbind

    return {
        ...originalModule, // Spread the original module to keep other exports if necessary
        __esModule: true, // Indicate this is an ES module
        default: defaultExportMock, // Override the default export with the mock function
    }
})

// We have to mock svelte, or it will attempt to run through the component lifecycle
// "Error: Function called outside component initialization"
vi.mock('svelte')

describe('Hotkey', () => {
    afterEach(() => {
        vi.restoreAllMocks()
    })

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

    describe('registerHotkey', () => {
        it('should get all key codes from hotkeys to check if there are duplicates', () => {
            mocks.getAllKeyCodes.mockReturnValueOnce([])

            const { bind } = registerHotkey({
                keys: { key: 'hello' },
                handler: () => {},
            })

            expect(mocks.getAllKeyCodes).toHaveBeenCalledOnce()

            bind({
                keys: { key: 'goodbye' },
                handler: () => {},
            })

            expect(mocks.unbind).toHaveBeenCalledOnce()
        })
    })
})
