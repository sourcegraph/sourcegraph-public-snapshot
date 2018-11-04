import * as assert from 'assert'
import { Subscription } from 'rxjs'
import { TextDocumentPositionParams, TextDocumentRegistrationOptions } from '../../protocol'
import { FeatureProviderRegistry as AbstractFeatureProviderRegistry } from './registry'

/** Useful test fixtures. */
export const FIXTURE = {
    TextDocumentPositionParams: {
        position: { line: 1, character: 2 },
        textDocument: { uri: 'file:///f' },
    } as TextDocumentPositionParams,

    PartialEntry: {
        registrationOptions: {
            documentSelector: ['*'],
        } as TextDocumentRegistrationOptions,
    },
}

class FeatureProviderRegistry extends AbstractFeatureProviderRegistry<TextDocumentRegistrationOptions, {}> {}

describe('FeatureProviderRegistry', () => {
    it('is initially empty', () => {
        assert.deepStrictEqual(new FeatureProviderRegistry().providersSnapshot, [])
    })

    it('accepts initial providers', () => {
        const initialEntries = [
            {
                ...FIXTURE.PartialEntry,
                provider: () => ({}),
            },
        ]
        assert.deepStrictEqual(
            new FeatureProviderRegistry(initialEntries).providersSnapshot,
            initialEntries.map(({ provider }) => provider)
        )
    })

    it('registers and unregisters providers', () => {
        const subscriptions = new Subscription()
        const registry = new FeatureProviderRegistry()
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
