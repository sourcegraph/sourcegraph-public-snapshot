import * as assert from 'assert'
import { TextDocumentDecorationProviderRegistry } from '../../environment/providers/decoration'
import { NotificationHandler } from '../../jsonrpc2/handlers'
import { NotificationType } from '../../jsonrpc2/messages'
import {
    ClientCapabilities,
    TextDocumentPublishDecorationsNotification,
    TextDocumentPublishDecorationsParams,
} from '../../protocol'
import { Client } from '../client'
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
        it('registers the provider and listens for notifications if the server has a decorationProvider', done => {
            const { client, registry, feature } = create()

            function mockOnNotification(method: string, handler: NotificationHandler<any>): void
            function mockOnNotification(
                type: NotificationType<TextDocumentPublishDecorationsParams, void>,
                params: NotificationHandler<TextDocumentPublishDecorationsParams>
            ): void
            function mockOnNotification(
                type: string | NotificationType<TextDocumentPublishDecorationsParams, void>,
                params: NotificationHandler<any>
            ): void {
                assert.strictEqual(
                    typeof type === 'string' ? type : type.method,
                    TextDocumentPublishDecorationsNotification.type.method
                )
                done()
            }
            client.onNotification = mockOnNotification

            feature.initialize({ decorationProvider: {} })
            assert.strictEqual(registry.providersSnapshot.length, 1)
        })

        it('does not register the provider if the server lacks a decorationProvider', () => {
            const { registry, feature } = create()
            feature.initialize({ decorationProvider: undefined })
            assert.strictEqual(registry.providersSnapshot.length, 0)
        })
    })
})
