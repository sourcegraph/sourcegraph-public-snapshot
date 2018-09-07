import * as assert from 'assert'
import { TextDocumentLocationProviderRegistry } from '../../environment/providers/location'
import { ClientCapabilities, DefinitionRequest, ReferenceParams, TextDocumentPositionParams } from '../../protocol'
import { Client } from '../client'
import {
    TextDocumentDefinitionFeature,
    TextDocumentImplementationFeature,
    TextDocumentLocationFeature,
    TextDocumentReferencesFeature,
    TextDocumentTypeDefinitionFeature,
} from './location'

const create = <P extends TextDocumentPositionParams, F extends TextDocumentLocationFeature<P>>(
    FeatureClass: new (client: Client, registry: TextDocumentLocationProviderRegistry<P>) => F,
    RegistryClass: new () => TextDocumentLocationProviderRegistry<P>
): {
    client: Client
    registry: TextDocumentLocationProviderRegistry<P>
    feature: F
} => {
    const client = { options: {} } as Client
    const registry = new RegistryClass()
    const feature = new FeatureClass(client, registry)
    return { client, registry, feature }
}

describe('TextDocumentLocationFeature', () => {
    describe('registration', () => {
        it('supports dynamic registration and unregistration', () => {
            const { registry, feature } = create(
                class extends TextDocumentLocationFeature {
                    public readonly messages = DefinitionRequest.type
                    public fillClientCapabilities(): void {
                        /* noop */
                    }
                    public isSupported(): boolean {
                        return true
                    }
                },
                TextDocumentLocationProviderRegistry
            )
            feature.register(feature.messages, { id: 'a', registerOptions: { documentSelector: ['*'] } })
            assert.strictEqual(registry.providersSnapshot.length, 1)
            feature.unregister('a')
            assert.strictEqual(registry.providersSnapshot.length, 0)
        })
    })
})

describe('TextDocumentDefinitionFeature', () => {
    it('reports client capabilities', () => {
        const capabilities: ClientCapabilities = {}
        create(TextDocumentDefinitionFeature, TextDocumentLocationProviderRegistry).feature.fillClientCapabilities(
            capabilities
        )
        assert.deepStrictEqual(capabilities, {
            textDocument: { definition: { dynamicRegistration: true } },
        } as ClientCapabilities)
    })
})

describe('TextDocumentImplementationFeature', () => {
    it('reports client capabilities', () => {
        const capabilities: ClientCapabilities = {}
        create(TextDocumentImplementationFeature, TextDocumentLocationProviderRegistry).feature.fillClientCapabilities(
            capabilities
        )
        assert.deepStrictEqual(capabilities, {
            textDocument: { implementation: { dynamicRegistration: true } },
        } as ClientCapabilities)
    })
})

describe('TextDocumentTypeDefinitionFeature', () => {
    it('reports client capabilities', () => {
        const capabilities: ClientCapabilities = {}
        create(TextDocumentTypeDefinitionFeature, TextDocumentLocationProviderRegistry).feature.fillClientCapabilities(
            capabilities
        )
        assert.deepStrictEqual(capabilities, {
            textDocument: { typeDefinition: { dynamicRegistration: true } },
        } as ClientCapabilities)
    })
})

describe('TextDocumentReferencesFeature', () => {
    it('reports client capabilities', () => {
        const capabilities: ClientCapabilities = {}
        create(
            TextDocumentReferencesFeature,
            TextDocumentLocationProviderRegistry as new () => TextDocumentLocationProviderRegistry<ReferenceParams>
        ).feature.fillClientCapabilities(capabilities)
        assert.deepStrictEqual(capabilities, {
            textDocument: { references: { dynamicRegistration: true } },
        } as ClientCapabilities)
    })
})
