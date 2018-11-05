import * as assert from 'assert'
import { integrationTestContext } from './helpers.test'

describe('Commands (integration)', () => {
    describe('commands.registerCommand', () => {
        it('registers and unregisters a single command', async () => {
            const { clientController, extensionHost, ready } = await integrationTestContext()
            await ready

            // Register the command and call it.
            const unsubscribe = extensionHost.commands.registerCommand('c', () => 'a')
            assert.strictEqual(await extensionHost.commands.executeCommand('c'), 'a')
            assert.strictEqual(await clientController.registries.commands.executeCommand({ command: 'c' }), 'a')

            // Unregister the command and ensure it's removed.
            unsubscribe.unsubscribe()
            await extensionHost.internal.sync()
            assert.rejects(() => extensionHost.commands.executeCommand('c')) // tslint:disable-line no-floating-promises
            assert.throws(() => clientController.registries.commands.executeCommand({ command: 'c' }))
        })

        it('supports multiple commands', async () => {
            const { clientController, extensionHost, ready } = await integrationTestContext()
            await ready

            // Register 2 commands with different results.
            extensionHost.commands.registerCommand('c1', () => 'a1')
            extensionHost.commands.registerCommand('c2', () => 'a2')
            await extensionHost.internal.sync()

            assert.strictEqual(await extensionHost.commands.executeCommand('c1'), 'a1')
            assert.strictEqual(await clientController.registries.commands.executeCommand({ command: 'c1' }), 'a1')
            assert.strictEqual(await extensionHost.commands.executeCommand('c2'), 'a2')
            assert.strictEqual(await clientController.registries.commands.executeCommand({ command: 'c2' }), 'a2')
        })
    })
})
