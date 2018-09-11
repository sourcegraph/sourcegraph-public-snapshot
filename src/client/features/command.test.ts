import * as assert from 'assert'
import { ClientCapabilities } from '../../protocol'
import { Client } from '../client'
import { CommandRegistry } from '../providers/command'
import { ExecuteCommandFeature } from './command'

const create = (): {
    client: Client
    registry: CommandRegistry
    feature: ExecuteCommandFeature
} => {
    const client = {} as Client
    const registry = new CommandRegistry()
    const feature = new ExecuteCommandFeature(client, registry)
    return { client, registry, feature }
}

describe('ExecuteCommandFeature', () => {
    it('reports client capabilities', () => {
        const capabilities: ClientCapabilities = {}
        create().feature.fillClientCapabilities(capabilities)
        assert.deepStrictEqual(capabilities, {
            workspace: { executeCommand: { dynamicRegistration: true } },
        } as ClientCapabilities)
    })

    describe('registration', () => {
        it('supports dynamic registration and unregistration', () => {
            const { registry, feature } = create()
            feature.register(feature.messages, { id: 'a', registerOptions: { commands: ['c'] } })
            assert.strictEqual(registry.commandsSnapshot.length, 1)
            feature.unregister('a')
            assert.strictEqual(registry.commandsSnapshot.length, 0)
        })

        it('supports multiple dynamic registrations and unregistrations', () => {
            const { registry, feature } = create()
            feature.register(feature.messages, { id: 'a', registerOptions: { commands: ['c1'] } })
            assert.strictEqual(registry.commandsSnapshot.length, 1)
            feature.register(feature.messages, { id: 'b', registerOptions: { commands: ['c2'] } })
            assert.strictEqual(registry.commandsSnapshot.length, 2)
            feature.unregister('b')
            assert.strictEqual(registry.commandsSnapshot.length, 1)
            feature.unregister('a')
            assert.strictEqual(registry.commandsSnapshot.length, 0)
        })

        it('prevents registration with conflicting IDs', () => {
            const { registry, feature } = create()
            feature.register(feature.messages, { id: 'a', registerOptions: { commands: ['c1'] } })
            assert.strictEqual(registry.commandsSnapshot.length, 1)
            assert.throws(() => {
                feature.register(feature.messages, { id: 'a', registerOptions: { commands: ['c2'] } })
            })
            assert.strictEqual(registry.commandsSnapshot.length, 1)
        })

        it('throws an error if ID to unregister is not registered', () => {
            const { registry, feature } = create()
            assert.throws(() => feature.unregister('a'))
            assert.strictEqual(registry.commandsSnapshot.length, 0)
        })

        it('atomically unregisters successfully registered commands if register call fails', () => {
            const { registry, feature } = create()
            // Duplicate command "c1" will yield an error.
            assert.throws(() =>
                feature.register(feature.messages, { id: 'a', registerOptions: { commands: ['c1', 'c2', 'c1'] } })
            )
            assert.strictEqual(registry.commandsSnapshot.length, 0)
        })
    })
})
