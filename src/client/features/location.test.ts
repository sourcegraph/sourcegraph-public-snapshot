import * as assert from 'assert'
import { TextDocumentLocationProviderRegistry } from '../../environment/providers/location'
import { ClientCapabilities, ServerCapabilities } from '../../protocol'
import { Client } from '../client'
import {
    ProvideTextDocumentLocationMiddleware,
    TextDocumentDefinitionFeature,
    TextDocumentImplementationFeature,
    TextDocumentLocationFeature,
    TextDocumentTypeDefinitionFeature,
} from './location'

const create = <F extends TextDocumentLocationFeature>(
    FeatureClass: new (client: Client, registry: TextDocumentLocationProviderRegistry) => F
): {
    client: Client
    registry: TextDocumentLocationProviderRegistry
    feature: F
} => {
    const client = { clientOptions: { middleware: {} } } as Client
    const registry = new TextDocumentLocationProviderRegistry()
    const feature = new FeatureClass(client, registry)
    return { client, registry, feature }
}

describe('TextDocumentLocationFeature', () => {
    describe('upon initialization', () => {
        it('registers the provider if the server has support', () => {
            const { registry, feature } = create(
                class extends TextDocumentLocationFeature {
                    public fillClientCapabilities(): void {
                        /* noop */
                    }
                    public isSupported(): boolean {
                        return true
                    }
                    public getMiddleware(): ProvideTextDocumentLocationMiddleware | undefined {
                        return undefined
                    }
                }
            )
            feature.initialize({ definitionProvider: true }, ['*'])
            assert.strictEqual(registry.providersSnapshot.length, 1)
        })

        it('does not register the provider if the server lacks support', () => {
            const { registry, feature } = create(
                class extends TextDocumentLocationFeature {
                    public fillClientCapabilities(): void {
                        /* noop */
                    }
                    public isSupported(): boolean {
                        return false
                    }
                    public getMiddleware(): ProvideTextDocumentLocationMiddleware | undefined {
                        return undefined
                    }
                }
            )
            feature.initialize({ definitionProvider: false }, ['*'])
            assert.strictEqual(registry.providersSnapshot.length, 0)
        })
    })
})

describe('TextDocumentDefinitionFeature', () => {
    it('reports client capabilities', () => {
        const capabilities: ClientCapabilities = {}
        create(TextDocumentDefinitionFeature).feature.fillClientCapabilities(capabilities)
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
            }
        )
        assert.strictEqual(feature.isSupported({}), false)
        assert.strictEqual(feature.isSupported({ definitionProvider: true }), true)
    })
})

describe('TextDocumentImplementationFeature', () => {
    it('reports client capabilities', () => {
        const capabilities: ClientCapabilities = {}
        create(TextDocumentImplementationFeature).feature.fillClientCapabilities(capabilities)
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
            }
        )
        assert.strictEqual(feature.isSupported({}), false)
        assert.strictEqual(feature.isSupported({ implementationProvider: true }), true)
    })
})

describe('TextDocumentTypeDefinitionFeature', () => {
    it('reports client capabilities', () => {
        const capabilities: ClientCapabilities = {}
        create(TextDocumentTypeDefinitionFeature).feature.fillClientCapabilities(capabilities)
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
            }
        )
        assert.strictEqual(feature.isSupported({}), false)
        assert.strictEqual(feature.isSupported({ typeDefinitionProvider: true }), true)
    })
})
