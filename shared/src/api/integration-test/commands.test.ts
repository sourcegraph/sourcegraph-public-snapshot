import assert from 'assert'
import { integrationTestContext } from './testHelpers'

describe('Commands (integration)', () => {
    describe('commands.registerCommand', () => {
        test('registers and unregisters a single command', async () => {
            const { services, extensionHost } = await integrationTestContext()

            // Register the command and call it.
            const unsubscribe = extensionHost.commands.registerCommand('c', () => 'a')
            expect(await extensionHost.commands.executeCommand('c')).toBe('a')
            expect(await services.commands.executeCommand({ command: 'c' })).toBe('a')

            // Unregister the command and ensure it's removed.
            unsubscribe.unsubscribe()
            await extensionHost.internal.sync()
            assert.rejects(() => extensionHost.commands.executeCommand('c')) // tslint:disable-line no-floating-promises
            expect(() => services.commands.executeCommand({ command: 'c' })).toThrow()
        })

        test('supports multiple commands', async () => {
            const { services, extensionHost } = await integrationTestContext()

            // Register 2 commands with different results.
            extensionHost.commands.registerCommand('c1', () => 'a1')
            extensionHost.commands.registerCommand('c2', () => 'a2')
            await extensionHost.internal.sync()

            expect(await extensionHost.commands.executeCommand('c1')).toBe('a1')
            expect(await services.commands.executeCommand({ command: 'c1' })).toBe('a1')
            expect(await extensionHost.commands.executeCommand('c2')).toBe('a2')
            expect(await services.commands.executeCommand({ command: 'c2' })).toBe('a2')
        })
    })
})
