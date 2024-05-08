// @vitest-environment jsdom

import hotkeys from 'hotkeys-js'
import { expect, describe, it, vi, beforeEach } from 'vitest'

import { evaluateKey, exportedForTesting, registerHotkey } from './Hotkey'

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

// Prevent the component from complaining that it's not in a svelte lifecycle.
vi.mock('svelte')

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

    describe('registerHotkey', () => {
        beforeEach(() => {
            hotkeys.getAllKeyCodes().map(hotkey => hotkeys.unbind(hotkey.shortcut))
        })

        it('should register hotkey', () => {
            registerHotkey({
                keys: {
                    key: 'hello',
                },
                handler: () => {},
            })

            const allKeyCodes = hotkeys.getAllKeyCodes()
            expect(allKeyCodes.length).toBe(1)
            expect(allKeyCodes[0].shortcut).toBe('hello')
        })

        it('should be able to re-bind a hotkey', () => {
            const { bind } = registerHotkey({
                keys: {
                    key: 'hello',
                },
                handler: () => {},
            })

            const allKeyCodesBefore = hotkeys.getAllKeyCodes()
            expect(allKeyCodesBefore.length).toBe(1)
            expect(allKeyCodesBefore[0].shortcut).toBe('hello')

            bind({
                keys: {
                    key: 'updated',
                },
                handler: () => {},
            })

            const allKeyCodesAfter = hotkeys.getAllKeyCodes()
            expect(allKeyCodesAfter.length).toBe(1)
            expect(allKeyCodesAfter[0].shortcut).toBe('updated')
        })

        it('should be able to un-register a hotkey', () => {
            const hotkey = registerHotkey({
                keys: {
                    key: 'hello',
                },
                handler: () => {},
            })

            const allKeyCodesBefore = hotkeys.getAllKeyCodes()
            expect(allKeyCodesBefore.length).toBe(1)
            expect(allKeyCodesBefore[0].shortcut).toBe('hello')

            hotkey.unregister()

            const allKeyCodesAfter = hotkeys.getAllKeyCodes()
            expect(allKeyCodesAfter.length).toBe(0)
        })

        it('should invoke the handler when hotkey is used', () => {
            let counter = 0
            registerHotkey({
                keys: {
                    key: 'hello',
                },
                handler: () => {
                    counter++
                },
            })

            expect(counter).toBe(0)
            hotkeys.trigger('hello')
            expect(counter).toBe(1)
        })

        it('should invoke the re-bound handler when hotkey is used', () => {
            let counter_1 = 0
            let counter_2 = 0
            const { bind } = registerHotkey({
                keys: {
                    key: 'hello',
                },
                handler: () => {
                    counter_1++
                },
            })

            bind({
                keys: {
                    key: 'hello',
                },
                handler: () => {
                    counter_2++
                },
            })

            expect(counter_1).toBe(0)
            expect(counter_2).toBe(0)
            hotkeys.trigger('hello')
            expect(counter_1).toBe(0)
            expect(counter_2).toBe(1)
        })
    })

    describe('isContentElement', () => {
        describe('with getAttribute', () => {
            it('should return false is no condition is met', () => {
                const element = document.createElement('div')
                expect(exportedForTesting.isContentElement(element)).toBe(false)
            })

            it('should return true if contenteditable is true', () => {
                const element = document.createElement('div')
                element.setAttribute('contenteditable', 'true')
                expect(exportedForTesting.isContentElement(element)).toBe(true)
            })

            it.each(['textarea', 'input', 'textbox'])('should return true if the role is %s', role => {
                const element = document.createElement('div')
                element.setAttribute('role', role)
                expect(exportedForTesting.isContentElement(element)).toBe(true)
            })

            it.each(['INPUT', 'TEXTAREA'])('should return true if the tagName is %s', tagName => {
                const element = document.createElement(tagName.toLowerCase())
                expect(exportedForTesting.isContentElement(element)).toBe(true)
            })
        })
    })
})
