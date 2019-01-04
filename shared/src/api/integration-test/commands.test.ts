import { integrationTestContext } from './testHelpers'

describe('Commands (integration)', () => {
    describe('commands.registerCommand', () => {
        test('registers and unregisters a single command', async () => {
            const { services, extensionHost } = await integrationTestContext()

            // Register the command and call it.
            const unsubscribe = extensionHost.commands.registerCommand('c', () => 'a')
            await expect(extensionHost.commands.executeCommand('c')).resolves.toBe('a')
            await expect(services.commands.executeCommand({ command: 'c' })).resolves.toBe('a')

            // Unregister the command and ensure it's removed.
            unsubscribe.unsubscribe()
            await extensionHost.internal.sync()
            await expect(extensionHost.commands.executeCommand('c')).rejects.toMatchObject({
                message: 'command not found: "c"',
            })
            expect(() => services.commands.executeCommand({ command: 'c' })).toThrow()
        })

        test('supports multiple commands', async () => {
            const { services, extensionHost } = await integrationTestContext()

            // Register 2 commands with different results.
            extensionHost.commands.registerCommand('c1', () => 'a1')
            extensionHost.commands.registerCommand('c2', () => 'a2')
            await extensionHost.internal.sync()

            await expect(extensionHost.commands.executeCommand('c1')).resolves.toBe('a1')
            await expect(services.commands.executeCommand({ command: 'c1' })).resolves.toBe('a1')
            await expect(extensionHost.commands.executeCommand('c2')).resolves.toBe('a2')
            await expect(services.commands.executeCommand({ command: 'c2' })).resolves.toBe('a2')
        })
    })
})
