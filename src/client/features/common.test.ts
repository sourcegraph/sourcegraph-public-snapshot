import * as assert from 'assert'
import { Unsubscribable } from 'rxjs'
import { TextDocumentRegistrationOptions } from '../../protocol'
import { Client } from '../client'
import { Feature as AbstractTextDocumentFeature } from './common'

const create = <F extends AbstractTextDocumentFeature<TextDocumentRegistrationOptions>>(
    FeatureClass: new (client: Client) => F
): {
    client: Client
    feature: F
} => {
    const client = {} as Client
    const feature = new FeatureClass(client)
    return { client, feature }
}

class TextDocumentFeature extends AbstractTextDocumentFeature<TextDocumentRegistrationOptions> {
    public readonly messages = { method: 'm' }
    protected registerProvider(): Unsubscribable {
        return { unsubscribe: () => void 0 }
    }
    protected validateRegistrationOptions(data: any): TextDocumentRegistrationOptions {
        return data
    }
    public fillClientCapabilities(): void {
        /* noop */
    }
}

const FIXTURE_REGISTER_OPTIONS: TextDocumentRegistrationOptions = { documentSelector: ['*'], extensionID: 'test' }

describe('TextDocumentFeature', () => {
    describe('dynamic registration', () => {
        it('supports dynamic registration and unregistration', () => {
            const { feature } = create(TextDocumentFeature)
            feature.register(feature.messages, { id: 'a', registerOptions: FIXTURE_REGISTER_OPTIONS })
            feature.unregister('a')
        })

        it('supports multiple dynamic registrations and unregistrations', () => {
            const { feature } = create(TextDocumentFeature)
            feature.register(feature.messages, { id: 'a', registerOptions: FIXTURE_REGISTER_OPTIONS })
            feature.register(feature.messages, { id: 'b', registerOptions: FIXTURE_REGISTER_OPTIONS })
            feature.unregister('b')
            feature.unregister('a')
        })

        it('prevents registration with conflicting IDs', () => {
            const { feature } = create(TextDocumentFeature)
            feature.register(feature.messages, { id: 'a', registerOptions: FIXTURE_REGISTER_OPTIONS })
            assert.throws(() => {
                feature.register(feature.messages, { id: 'a', registerOptions: FIXTURE_REGISTER_OPTIONS })
            })
        })

        it('throws an error if ID to unregister is not registered', () => {
            const { feature } = create(TextDocumentFeature)
            assert.throws(() => feature.unregister('a'))
        })
    })
})
