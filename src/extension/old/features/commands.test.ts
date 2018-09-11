import assert from 'assert'
import {
    ExecuteCommandParams,
    ExecuteCommandRegistrationOptions,
    ExecuteCommandRequest,
    RegistrationParams,
    RegistrationRequest,
    UnregistrationParams,
    UnregistrationRequest,
} from '../../../protocol'
import { MockMessageConnection } from '../../../protocol/jsonrpc2/test/mockMessageConnection'
import { Commands } from '../api'
import { createExtCommands } from './commands'

describe('ExtCommands', () => {
    function create(): {
        extCommands: Commands
        mockConnection: MockMessageConnection
    } {
        const mockConnection = new MockMessageConnection()
        const extCommands = createExtCommands(mockConnection)
        return { extCommands, mockConnection }
    }

    describe('register', () => {
        describe('registering a command', async () => {
            const { extCommands, mockConnection } = create()
            mockConnection.mockResults.set(RegistrationRequest.type.method, void 0)

            let called = false
            let callArgs: any
            extCommands.register('a', (...args: any[]) => {
                called = true
                callArgs = args
            })
            assert.strictEqual(called, false)

            it('sends a request', () =>
                assert.deepStrictEqual(mockConnection.sentMessages, [
                    {
                        method: RegistrationRequest.type.method,
                        params: {
                            registrations: [
                                {
                                    id: (mockConnection.sentMessages[0].params as RegistrationParams).registrations[0]
                                        .id,
                                    method: ExecuteCommandRequest.type.method,
                                    registerOptions: { commands: ['a'] } as ExecuteCommandRegistrationOptions,
                                },
                            ],
                        } as RegistrationParams,
                    },
                ]))

            it('invokes the run function when the client executes the command', async () => {
                await mockConnection.recvRequest(ExecuteCommandRequest.type.method, {
                    command: 'a',
                    arguments: [1],
                } as ExecuteCommandParams)
                assert.ok(called)
                assert.deepStrictEqual(callArgs, [1])
            })
        })

        it('unregisters a command', () => {
            const { extCommands, mockConnection } = create()
            mockConnection.mockResults.set(RegistrationRequest.type.method, void 0)

            let called = false
            const subscription = extCommands.register('a', () => (called = true))

            mockConnection.sentMessages = [] // clear
            mockConnection.mockResults.set(UnregistrationRequest.type.method, void 0)
            subscription.unsubscribe()
            const id = (mockConnection.sentMessages[0].params as UnregistrationParams).unregisterations[0].id
            assert.deepStrictEqual(mockConnection.sentMessages, [
                {
                    method: UnregistrationRequest.type.method,
                    params: {
                        unregisterations: [{ id, method: ExecuteCommandRequest.type.method }],
                    } as UnregistrationParams,
                },
            ])

            assert.throws(() =>
                mockConnection.recvRequest(ExecuteCommandRequest.type.method, {
                    command: 'a',
                } as ExecuteCommandParams)
            )
            assert.strictEqual(called, false)
        })
    })
})
