import { describe, expect, test } from '@jest/globals'

import { integrationTestContext } from '../../testing/testHelpers'

describe('Commands (integration)', () => {
    describe('commands.registerCommand', () => {
        test('registers and unregisters a single command', async () => {
            const { extensionAPI, mainThreadAPI } = await integrationTestContext()

            // Register the command and call it.
            const unsubscribe = extensionAPI.commands.registerCommand('c', () => 'a')
            // await extensionAPI.internal.sync()

            await expect(extensionAPI.commands.executeCommand('c')).resolves.toBe('a')
            await expect(mainThreadAPI.executeCommand('c', [])).resolves.toBe('a')

            // Unregister the command and ensure it's removed.
            unsubscribe.unsubscribe()
            await extensionAPI.internal.sync()
            await expect(extensionAPI.commands.executeCommand('c')).rejects.toMatchObject({
                message: 'command not found: "c"',
            })
            expect(() => mainThreadAPI.executeCommand('c', [])).toThrow()
        })

        test('supports multiple commands', async () => {
            const { mainThreadAPI, extensionAPI } = await integrationTestContext()

            // Register 2 commands with different results.
            extensionAPI.commands.registerCommand('c1', () => 'a1')
            extensionAPI.commands.registerCommand('c2', () => 'a2')
            await extensionAPI.internal.sync()

            await expect(extensionAPI.commands.executeCommand('c1')).resolves.toBe('a1')
            await expect(mainThreadAPI.executeCommand('c1', [])).resolves.toBe('a1')
            await expect(extensionAPI.commands.executeCommand('c2')).resolves.toBe('a2')
            await expect(mainThreadAPI.executeCommand('c2', [])).resolves.toBe('a2')
        })
    })
})
