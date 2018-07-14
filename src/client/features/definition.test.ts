import * as assert from 'assert'
import { TextDocumentLocationProviderRegistry } from '../../environment/providers/location'
import { ClientCapabilities } from '../../protocol'
import { Client } from '../client'
import { TextDocumentDefinitionFeature } from './definition'

const create = (): {
    client: Client
    registry: TextDocumentLocationProviderRegistry
    feature: TextDocumentDefinitionFeature
} => {
    const client = { clientOptions: { middleware: {} } } as Client
    const registry = new TextDocumentLocationProviderRegistry()
    const feature = new TextDocumentDefinitionFeature(client, registry)
    return { client, registry, feature }
}

describe('TextDocumentDefinitionFeature', () => {
    it('reports client capabilities', () => {
        const capabilities: ClientCapabilities = {}
        create().feature.fillClientCapabilities(capabilities)
        assert.deepStrictEqual(capabilities, {
            textDocument: { definition: { dynamicRegistration: true } },
        } as ClientCapabilities)
    })

    describe('upon initialization', () => {
        it('registers the provider if the server has definitionProvider', () => {
            const { registry, feature } = create()
            feature.initialize({ definitionProvider: true }, ['*'])
            assert.strictEqual(registry.providersSnapshot.length, 1)
        })

        it('does not register the provider if the server lacks definitionProvider', () => {
            const { registry, feature } = create()
            feature.initialize({ definitionProvider: false }, ['*'])
            assert.strictEqual(registry.providersSnapshot.length, 0)
        })
    })
})
