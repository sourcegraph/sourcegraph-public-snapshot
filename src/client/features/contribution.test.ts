import * as assert from 'assert'
import { EMPTY_OBSERVABLE_ENVIRONMENT } from '../../environment/environment'
import { ContributionRegistry, ContributionsEntry } from '../../environment/providers/contribution'
import { ClientCapabilities } from '../../protocol'
import { Client } from '../client'
import { ContributionFeature } from './contribution'

const create = (): {
    client: Client
    registry: ContributionRegistry
    feature: ContributionFeature
} => {
    const client = {} as Client
    const registry = new ContributionRegistry(EMPTY_OBSERVABLE_ENVIRONMENT.environment)
    const feature = new ContributionFeature(registry)
    return { client, registry, feature }
}

describe('ContributionFeature', () => {
    it('reports client capabilities', () => {
        const capabilities: ClientCapabilities = {}
        create().feature.fillClientCapabilities(capabilities)
        assert.deepStrictEqual(capabilities, {
            window: { contribution: { dynamicRegistration: true } },
        } as ClientCapabilities)
    })

    describe('registration', () => {
        it('supports dynamic registration and unregistration', () => {
            const { registry, feature } = create()
            feature.register(feature.messages, { id: 'a', registerOptions: {} })
            assert.strictEqual(registry.entries.value.length, 1)
            feature.unregister('a')
            assert.strictEqual(registry.entries.value.length, 0)
        })

        it('supports multiple dynamic registrations and unregistrations', () => {
            const { registry, feature } = create()
            feature.register(feature.messages, { id: 'a', registerOptions: {} })
            assert.strictEqual(registry.entries.value.length, 1)
            feature.register(feature.messages, { id: 'b', registerOptions: {} })
            assert.strictEqual(registry.entries.value.length, 2)
            feature.unregister('b')
            assert.strictEqual(registry.entries.value.length, 1)
            feature.unregister('a')
            assert.strictEqual(registry.entries.value.length, 0)
        })

        it('prevents registration with conflicting IDs', () => {
            const { registry, feature } = create()
            feature.register(feature.messages, { id: 'a', registerOptions: {} })
            assert.strictEqual(registry.entries.value.length, 1)
            assert.throws(() => {
                feature.register(feature.messages, { id: 'a', registerOptions: {} })
            })
            assert.strictEqual(registry.entries.value.length, 1)
        })

        it('throws an error if ID to unregister is not registered', () => {
            const { registry, feature } = create()
            assert.throws(() => feature.unregister('a'))
            assert.strictEqual(registry.entries.value.length, 0)
        })

        describe('overwriteExisting', () => {
            it('rejects registration request with overwriteExisting when there is no existing registration', () => {
                const { registry, feature } = create()
                assert.throws(() =>
                    feature.register(feature.messages, { id: 'a', registerOptions: {}, overwriteExisting: true })
                )
                assert.strictEqual(registry.entries.value.length, 0)
            })

            it('overwrites existing registration', () => {
                const { registry, feature } = create()
                feature.register(feature.messages, {
                    id: 'a',
                    registerOptions: { actions: [{ id: '1', command: '1' }] },
                })
                feature.register(feature.messages, {
                    id: 'a',
                    registerOptions: { actions: [{ id: '2', command: '2' }] },
                    overwriteExisting: true,
                })
                assert.strictEqual(registry.entries.value.length, 1)
                assert.deepStrictEqual(registry.entries.value[0], {
                    contributions: { actions: [{ id: '2', command: '2' }] },
                } as ContributionsEntry)
            })
        })
    })
})
