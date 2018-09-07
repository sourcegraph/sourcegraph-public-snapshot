import * as assert from 'assert'
import { TextDocumentLocationProviderRegistry } from '../../environment/providers/location'
import {
    ClientCapabilities,
    DefinitionRequest,
    ReferenceParams,
    ServerCapabilities,
    TextDocumentPositionParams,
} from '../../protocol'
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
    describe('upon initialization', () => {
        it('registers the provider if the server has support', () => {
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
            feature.initialize({ definitionProvider: true }, ['*'])
            assert.strictEqual(registry.providersSnapshot.length, 1)
        })

        it('does not register the provider if the server lacks support', () => {
            const { registry, feature } = create(
                class extends TextDocumentLocationFeature {
                    public readonly messages = DefinitionRequest.type
                    public fillClientCapabilities(): void {
                        /* noop */
                    }
                    public isSupported(): boolean {
                        return false
                    }
                },
                TextDocumentLocationProviderRegistry
            )
            feature.initialize({ definitionProvider: false }, ['*'])
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

    it('reports server support', () => {
        const { feature } = create(
            // Create anonymous subclass to make isSupported public.
            class extends TextDocumentDefinitionFeature {
                public isSupported(capabilities: ServerCapabilities): boolean {
                    return super.isSupported(capabilities)
                }
            },
            TextDocumentLocationProviderRegistry
        )
        assert.strictEqual(feature.isSupported({}), false)
        assert.strictEqual(feature.isSupported({ definitionProvider: true }), true)
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

    it('reports server support', () => {
        const { feature } = create(
            // Create anonymous subclass to make isSupported public.
            class extends TextDocumentImplementationFeature {
                public isSupported(capabilities: ServerCapabilities): boolean {
                    return super.isSupported(capabilities)
                }
            },
            TextDocumentLocationProviderRegistry
        )
        assert.strictEqual(feature.isSupported({}), false)
        assert.strictEqual(feature.isSupported({ implementationProvider: true }), true)
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

    it('reports server support', () => {
        const { feature } = create(
            // Create anonymous subclass to make isSupported public.
            class extends TextDocumentTypeDefinitionFeature {
                public isSupported(capabilities: ServerCapabilities): boolean {
                    return super.isSupported(capabilities)
                }
            },
            TextDocumentLocationProviderRegistry
        )
        assert.strictEqual(feature.isSupported({}), false)
        assert.strictEqual(feature.isSupported({ typeDefinitionProvider: true }), true)
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

    it('reports server support', () => {
        const { feature } = create(
            // Create anonymous subclass to make isSupported public.
            class extends TextDocumentReferencesFeature {
                public isSupported(capabilities: ServerCapabilities): boolean {
                    return super.isSupported(capabilities)
                }
            },
            TextDocumentLocationProviderRegistry as new () => TextDocumentLocationProviderRegistry<ReferenceParams>
        )
        assert.strictEqual(feature.isSupported({}), false)
        assert.strictEqual(feature.isSupported({ referencesProvider: true }), true)
    })
})
