import * as assert from 'assert'
import { MarkupKind } from 'vscode-languageserver-types'
import { TextDocumentHoverProviderRegistry } from '../../environment/providers/hover'
import { ClientCapabilities } from '../../protocol'
import { Client } from '../client'
import { TextDocumentHoverFeature } from './hover'

const create = (): {
    client: Client
    registry: TextDocumentHoverProviderRegistry
    feature: TextDocumentHoverFeature
} => {
    const client = { options: {} } as Client
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

    describe('upon initialization', () => {
        it('registers the provider if the server has hoverProvider', () => {
            const { registry, feature } = create()
            feature.initialize({ hoverProvider: true }, ['*'])
            assert.strictEqual(registry.providersSnapshot.length, 1)
        })

        it('does not register the provider if the server lacks hoverProvider', () => {
            const { registry, feature } = create()
            feature.initialize({ hoverProvider: false }, ['*'])
            assert.strictEqual(registry.providersSnapshot.length, 0)
        })
    })
})
