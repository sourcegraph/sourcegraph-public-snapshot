import * as assert from 'assert'
import { Subscription } from 'rxjs'
import { Position } from 'vscode-languageserver-types'
import { TextDocumentPositionParams, TextDocumentRegistrationOptions } from '../../protocol'
import { TextDocumentFeatureProviderRegistry as AbstractTextDocumentFeatureProviderRegistry } from './textDocument'

/** Useful test fixtures. */
export const FIXTURE = {
    TextDocumentPositionParams: {
        position: Position.create(1, 2),
        textDocument: { uri: 'file:///f' },
    } as TextDocumentPositionParams,

    PartialEntry: {
        registrationOptions: {
            documentSelector: ['*'],
        } as TextDocumentRegistrationOptions,
    },
}

class TextDocumentFeatureProviderRegistry extends AbstractTextDocumentFeatureProviderRegistry<
    TextDocumentRegistrationOptions,
    {}
> {}

describe('TextDocumentFeatureProviderRegistry', () => {
    it('is initially empty', () => {
        assert.deepStrictEqual(new TextDocumentFeatureProviderRegistry().providersSnapshot, [])
    })

    it('accepts initial providers', () => {
        const initialEntries = [
            {
                ...FIXTURE.PartialEntry,
                provider: () => ({}),
            },
        ]
        assert.deepStrictEqual(
            new TextDocumentFeatureProviderRegistry(initialEntries).providersSnapshot,
            initialEntries.map(({ provider }) => provider)
        )
    })

    it('registers and unregisters providers', () => {
        const subscriptions = new Subscription()
        const registry = new TextDocumentFeatureProviderRegistry()
        const provider1 = () => ({})
        const provider2 = () => ({})

        const unregister1 = subscriptions.add(
            registry.registerProvider(FIXTURE.PartialEntry.registrationOptions, provider1)
        )
        assert.deepStrictEqual(registry.providersSnapshot, [provider1])

        const unregister2 = subscriptions.add(
            registry.registerProvider(FIXTURE.PartialEntry.registrationOptions, provider2)
        )
        assert.deepStrictEqual(registry.providersSnapshot, [provider1, provider2])

        unregister1.unsubscribe()
        assert.deepStrictEqual(registry.providersSnapshot, [provider2])

        unregister2.unsubscribe()
        assert.deepStrictEqual(registry.providersSnapshot, [])
    })
})
