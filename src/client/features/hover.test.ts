import * as assert from 'assert'
import { MarkupKind } from 'sourcegraph'
import { ClientCapabilities } from '../../protocol'
import { Client } from '../client'
import { TextDocumentHoverProviderRegistry } from '../providers/hover'
import { TextDocumentHoverFeature } from './hover'

const create = (): {
    client: Client
    registry: TextDocumentHoverProviderRegistry
    feature: TextDocumentHoverFeature
} => {
    const client = { options: {}, id: 'test' } as Client
    const registry = new TextDocumentHoverProviderRegistry()
    const feature = new TextDocumentHoverFeature(client, registry)
    return { client, registry, feature }
}

describe('TextDocumentHoverFeature', () => {
    it('reports client capabilities', () => {
        const capabilities: ClientCapabilities = {}
        create().feature.fillClientCapabilities(capabilities)
        assert.deepStrictEqual(capabilities, {
            textDocument: {
                hover: { dynamicRegistration: true, contentFormat: [MarkupKind.Markdown, MarkupKind.PlainText] },
            },
        } as ClientCapabilities)
    })

    describe('registration', () => {
        it('supports dynamic registration and unregistration', () => {
            const { registry, feature } = create()
            feature.register(feature.messages, {
                id: 'a',
                registerOptions: { documentSelector: ['*'] },
            })
            assert.strictEqual(registry.providersSnapshot.length, 1)
            feature.unregister('a')
            assert.strictEqual(registry.providersSnapshot.length, 0)
        })
    })
})
