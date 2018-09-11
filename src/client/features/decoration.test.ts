import * as assert from 'assert'
import { ClientCapabilities, TextDocumentPublishDecorationsNotification } from '../../protocol'
import { NotificationHandler } from '../../protocol/jsonrpc2/handlers'
import { Client } from '../client'
import { TextDocumentDecorationProviderRegistry } from '../providers/decoration'
import { TextDocumentDecorationFeature } from './decoration'

describe('TextDocumentDecorationFeature', () => {
    const create = (): {
        client: Client
        registry: TextDocumentDecorationProviderRegistry
        feature: TextDocumentDecorationFeature
    } => {
        const client = { options: {} } as Client
        const registry = new TextDocumentDecorationProviderRegistry()
        const feature = new TextDocumentDecorationFeature(client, registry)
        return { client, registry, feature }
    }

    it('reports client capabilities', () => {
        const capabilities: ClientCapabilities = {}
        create().feature.fillClientCapabilities(capabilities)
        assert.deepStrictEqual(capabilities, { decoration: {} } as ClientCapabilities)
    })

    describe('upon initialization', () => {
        it('registers the provider and listens for notifications', done => {
            const { client, registry, feature } = create()

            function mockOnNotification(method: string, _handler: NotificationHandler<any>): void {
                assert.strictEqual(method, TextDocumentPublishDecorationsNotification.type)
                done()
            }
            client.onNotification = mockOnNotification

            feature.initialize()
            assert.strictEqual(registry.providersSnapshot.length, 1)
        })
    })
})
