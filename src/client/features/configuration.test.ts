import * as assert from 'assert'
import { Subject } from 'rxjs'
import { ClientCapabilities } from '../../protocol'
import { Client } from '../client'
import { ConfigurationChangeNotificationFeature } from './configuration'

const create = (): {
    client: Client
    settings: Subject<any>
    feature: ConfigurationChangeNotificationFeature<any>
} => {
    const client = { options: { middleware: {} } } as Client
    const settings = new Subject<any>()
    const feature = new ConfigurationChangeNotificationFeature(client, settings)
    return { client, settings, feature }
}

describe('ConfigurationChangeNotificationFeature', () => {
    it('reports client capabilities', () => {
        const capabilities: ClientCapabilities = {}
        create().feature.fillClientCapabilities(capabilities)
        assert.deepStrictEqual(capabilities, {
            workspace: { didChangeConfiguration: { dynamicRegistration: true } },
        } as ClientCapabilities)
    })

    describe('upon initialization', () => {
        it('registers the provider', () => {
            const { feature } = create()
            feature.initialize({})
        })
    })

    describe('registration', () => {
        it('supports dynamic registration and unregistration', () => {
            const { feature } = create()
            feature.register(feature.messages, { id: 'a', registerOptions: undefined })
            feature.unregister('a')
        })

        it('supports dynamic registration and unregistration after static registration also occurred', () => {
            const { feature } = create()
            feature.initialize({})
            feature.register(feature.messages, { id: 'a', registerOptions: undefined })
            feature.unregister('a')
        })

        it('supports multiple dynamic registrations and unregistrations', () => {
            const { feature } = create()
            feature.register(feature.messages, { id: 'a', registerOptions: undefined })
            feature.register(feature.messages, { id: 'b', registerOptions: undefined })
            feature.unregister('b')
            feature.unregister('a')
        })

        it('prevents registration with conflicting IDs', () => {
            const { feature } = create()
            feature.register(feature.messages, { id: 'a', registerOptions: undefined })
            assert.throws(() => {
                feature.register(feature.messages, { id: 'a', registerOptions: undefined })
            })
        })

        it('throws an error if ID to unregister is not registered', () => {
            const { feature } = create()
            assert.throws(() => feature.unregister('a'))
        })
    })
})
