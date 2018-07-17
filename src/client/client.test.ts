import * as assert from 'assert'
import { MessageType as RPCMessageType } from '../jsonrpc2/messages'
import { ClientCapabilities, InitializeParams, RegistrationParams, UnregistrationParams } from '../protocol'
import { Client } from './client'
import { DynamicFeature, RegistrationData, StaticFeature } from './features/common'

class TestClient extends Client {
    constructor() {
        super('', '', {
            root: null,
            createMessageTransports: () => {
                throw new Error('connection is not used in unit test')
            },
        })
    }

    public readonly features!: Client['features']
    public handleRegistrationRequest(params: RegistrationParams): void {
        super.handleRegistrationRequest(params)
    }
    public handleUnregistrationRequest(params: UnregistrationParams): void {
        super.handleUnregistrationRequest(params)
    }
}

const FIXTURE_STATIC_FEATURE: StaticFeature = {
    fillInitializeParams: (params: InitializeParams) => (params.initializationOptions = 'test'),
    fillClientCapabilities: (capabilities: ClientCapabilities) => (capabilities.experimental = 'test'),
    initialize: () => void 0,
}

const FIXTURE_DYNAMIC_FEATURE: DynamicFeature<any> = {
    messages: { method: 'm' },
    fillInitializeParams: (params: InitializeParams) => (params.initializationOptions = 'test'),
    fillClientCapabilities: (capabilities: ClientCapabilities) => (capabilities.experimental = 'test'),
    initialize: () => void 0,
    register: () => void 0,
    unregister: () => void 0,
    unregisterAll: () => void 0,
}

describe('Client', () => {
    describe('features', () => {
        it('registers static feature', () => {
            const client = new TestClient()
            client.registerFeature(FIXTURE_STATIC_FEATURE)
            assert.deepStrictEqual(client.features, [FIXTURE_STATIC_FEATURE])
        })

        it('registers dynamic feature', () => {
            const client = new TestClient()
            client.registerFeature(FIXTURE_DYNAMIC_FEATURE)
            assert.deepStrictEqual(client.features, [FIXTURE_DYNAMIC_FEATURE])
        })

        it('prevents dynamic feature registration with conflicting method', () => {
            const client = new TestClient()
            client.registerFeature(FIXTURE_DYNAMIC_FEATURE)
            assert.throws(() => client.registerFeature(FIXTURE_DYNAMIC_FEATURE))
            assert.deepStrictEqual(client.features, [FIXTURE_DYNAMIC_FEATURE])
        })

        describe('dynamic (un)registration', () => {
            interface RegisterCallArgs {
                message: RPCMessageType
                data: RegistrationData<any>
            }

            it('handles a registration request for a dynamic feature', () => {
                const client = new TestClient()
                const registerCalls: RegisterCallArgs[] = []
                const unregisterCalls: string[] = []
                client.registerFeature({
                    ...FIXTURE_DYNAMIC_FEATURE,
                    register: (message, data) => registerCalls.push({ message, data }),
                    unregister: id => unregisterCalls.push(id),
                })

                // Request registration.
                client.handleRegistrationRequest({
                    registrations: [{ id: 'a', method: 'm', registerOptions: { a: 1 } }],
                })
                assert.deepStrictEqual(registerCalls, [
                    { message: { method: 'm' }, data: { id: 'a', registerOptions: { a: 1 } } },
                ] as typeof registerCalls)

                // Request unregistration.
                client.handleUnregistrationRequest({
                    unregisterations: [{ id: 'a', method: 'm' }],
                })
                assert.deepStrictEqual(unregisterCalls, ['a'] as typeof unregisterCalls)
            })

            it('rejects registration requests for unknown dynamic features', () => {
                const client = new TestClient()
                assert.throws(() => client.handleRegistrationRequest({ registrations: [{ id: 'a', method: 'x' }] }))
            })

            it('rejects unregistration requests for unknown dynamic features', () => {
                const client = new TestClient()
                assert.throws(() =>
                    client.handleUnregistrationRequest({ unregisterations: [{ id: 'a', method: 'x' }] })
                )
            })
        })
    })
})
