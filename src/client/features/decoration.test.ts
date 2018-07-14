import * as assert from 'assert'
import { TextDocumentDecorationProviderRegistry } from '../../environment/providers/decoration'
import { NotificationHandler } from '../../jsonrpc2/handlers'
import { NotificationType } from '../../jsonrpc2/messages'
import {
    ClientCapabilities,
    TextDocumentPublishDecorationsNotification,
    TextDocumentPublishDecorationsParams,
    TextDocumentRegistrationOptions,
} from '../../protocol'
import { Client } from '../client'
import { TextDocumentDynamicDecorationFeature, TextDocumentStaticDecorationFeature } from './decoration'

describe('TextDocumentStaticDecorationFeature', () => {
    const create = (): {
        client: Client
        registry: TextDocumentDecorationProviderRegistry
        feature: TextDocumentStaticDecorationFeature
    } => {
        const client = { clientOptions: { middleware: {} } } as Client
        const registry = new TextDocumentDecorationProviderRegistry()
        const feature = new TextDocumentStaticDecorationFeature(client, registry)
        return { client, registry, feature }
    }

    it('reports client capabilities', () => {
        const capabilities: ClientCapabilities = {}
        create().feature.fillClientCapabilities(capabilities)
        assert.deepStrictEqual(capabilities, {
            decoration: { static: true },
        } as ClientCapabilities)
    })

    describe('upon initialization', () => {
        it('registers the provider if the server has static decorationProvider', () => {
            const { registry, feature } = create()
            feature.initialize({ decorationProvider: { static: true } }, ['*'])
            assert.strictEqual(registry.providersSnapshot.length, 1)
        })

        it('does not register the provider if the server lacks static decorationProvider', () => {
            const { registry, feature } = create()
            feature.initialize({ decorationProvider: { static: false } }, ['*'])
            assert.strictEqual(registry.providersSnapshot.length, 0)
        })
    })
})

describe('TextDocumentDynamicDecorationFeature', () => {
    const create = (): {
        client: Client
        registry: TextDocumentDecorationProviderRegistry
        feature: TextDocumentDynamicDecorationFeature
    } => {
        const client = { clientOptions: { middleware: {} } } as Client
        const registry = new TextDocumentDecorationProviderRegistry()
        const feature = new TextDocumentDynamicDecorationFeature(client, registry)
        return { client, registry, feature }
    }

    it('reports client capabilities', () => {
        const capabilities: ClientCapabilities = {}
        create().feature.fillClientCapabilities(capabilities)
        assert.deepStrictEqual(capabilities, {
            decoration: { dynamic: true },
        } as ClientCapabilities)
    })

    describe('upon initialization', () => {
        it('registers the provider and listens for notifications if the server has dynamic decorationProvider', done => {
            const { client, registry, feature } = create()

            function mockOnNotification(method: string, handler: NotificationHandler<any>): void
            function mockOnNotification(
                type: NotificationType<TextDocumentPublishDecorationsParams, TextDocumentRegistrationOptions>,
                params: NotificationHandler<TextDocumentPublishDecorationsParams>
            ): void
            function mockOnNotification(
                type: string | NotificationType<TextDocumentPublishDecorationsParams, TextDocumentRegistrationOptions>,
                params: NotificationHandler<any>
            ): void {
                assert.strictEqual(
                    typeof type === 'string' ? type : type.method,
                    TextDocumentPublishDecorationsNotification.type.method
                )
                done()
            }
            client.onNotification = mockOnNotification

            feature.initialize({ decorationProvider: { dynamic: true } }, ['*'])
            assert.strictEqual(registry.providersSnapshot.length, 1)
        })

        it('does not register the provider if the server lacks dynamic decorationProvider', () => {
            const { registry, feature } = create()
            feature.initialize({ decorationProvider: { dynamic: false } }, ['*'])
            assert.strictEqual(registry.providersSnapshot.length, 0)
        })
    })
})
