import * as assert from 'assert'
import { Subscription } from 'rxjs'
import { CommandEntry, CommandRegistry, executeCommand } from './command'

describe('CommandRegistry', () => {
    it('is initially empty', () => {
        assert.deepStrictEqual(new CommandRegistry().commandsSnapshot, [])
    })

    it('registers and unregisters commands', () => {
        const subscriptions = new Subscription()
        const registry = new CommandRegistry()
        const entry1: CommandEntry = { command: 'command1', run: async () => void 0 }
        const entry2: CommandEntry = { command: 'command2', run: async () => void 0 }

        const unregister1 = subscriptions.add(registry.registerCommand(entry1))
        assert.deepStrictEqual(registry.commandsSnapshot, [entry1])

        const unregister2 = subscriptions.add(registry.registerCommand(entry2))
        assert.deepStrictEqual(registry.commandsSnapshot, [entry1, entry2])

        unregister1.unsubscribe()
        assert.deepStrictEqual(registry.commandsSnapshot, [entry2])

        unregister2.unsubscribe()
        assert.deepStrictEqual(registry.commandsSnapshot, [])
    })

    it('refuses to register 2 commands with the same ID', () => {
        const registry = new CommandRegistry()
        registry.registerCommand({ command: 'c', run: async () => void 0 })
        assert.throws(() => {
            registry.registerCommand({ command: 'c', run: async () => void 0 })
        })
    })

    it('runs the specified command', async () => {
        const registry = new CommandRegistry()
        registry.registerCommand({
            command: 'command1',
            run: async arg => {
                assert.strictEqual(arg, 123)
                return 456
            },
        })
        assert.strictEqual(await registry.executeCommand({ command: 'command1', arguments: [123] }), 456)
    })
})

describe('executeCommand', () => {
    it('runs the specified command with args', async () => {
        const commands: CommandEntry[] = [
            {
                command: 'command1',
                run: async arg => {
                    assert.strictEqual(arg, 123)
                    return 456
                },
            },
        ]
        assert.strictEqual(await executeCommand(commands, { command: 'command1', arguments: [123] }), 456)
    })

    it('runs the specified command with no args', async () => {
        const commands: CommandEntry[] = [{ command: 'command1', run: async arg => void 0 }]
        assert.strictEqual(await executeCommand(commands, { command: 'command1' }), undefined)
    })

    it('throws an exception if the command is not found', () => {
        assert.throws(() => executeCommand([], { command: 'c' }))
    })
})
