import * as assert from 'assert'
import { CommandRegistry } from '../../environment/providers/command'
import { ClientCapabilities } from '../../protocol'
import { Client } from '../client'
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

    describe('upon initialization', () => {
        it('registers the provider if the server has executeCommandProvider', () => {
            const { registry, feature } = create()
            feature.initialize({ executeCommandProvider: { commands: ['c1', 'c2'] } })
            assert.strictEqual(registry.commandsSnapshot.length, 2)
        })

        it('does not register the provider if the server lacks executeCommandProvider', () => {
            const { registry, feature } = create()
            feature.initialize({ executeCommandProvider: undefined })
            assert.strictEqual(registry.commandsSnapshot.length, 0)
        })
    })
})
