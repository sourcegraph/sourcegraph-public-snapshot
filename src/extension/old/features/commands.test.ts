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
import { MockConnection } from '../../../protocol/jsonrpc2/test/mockConnection'
import { Commands } from '../api'
import { createExtCommands } from './commands'

describe('ExtCommands', () => {
    function create(): {
        extCommands: Commands
        mockConnection: MockConnection
    } {
        const mockConnection = new MockConnection()
        const extCommands = createExtCommands(mockConnection)
        return { extCommands, mockConnection }
    }

    describe('register', () => {
        describe('registering a command', async () => {
            const { extCommands, mockConnection } = create()
            mockConnection.mockResults.set(RegistrationRequest.type, void 0)

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
                        method: RegistrationRequest.type,
                        params: {
                            registrations: [
                                {
                                    id: (mockConnection.sentMessages[0].params as RegistrationParams).registrations[0]
                                        .id,
                                    method: ExecuteCommandRequest.type,
                                    registerOptions: { commands: ['a'] } as ExecuteCommandRegistrationOptions,
                                },
                            ],
                        } as RegistrationParams,
                    },
                ]))

            it('invokes the run function when the client executes the command', async () => {
                await mockConnection.recvRequest(ExecuteCommandRequest.type, {
                    command: 'a',
                    arguments: [1],
                } as ExecuteCommandParams)
                assert.ok(called)
                assert.deepStrictEqual(callArgs, [1])
            })
        })

        it('unregisters a command', () => {
            const { extCommands, mockConnection } = create()
            mockConnection.mockResults.set(RegistrationRequest.type, void 0)

            let called = false
            const subscription = extCommands.register('a', () => (called = true))

            mockConnection.sentMessages = [] // clear
            mockConnection.mockResults.set(UnregistrationRequest.type, void 0)
            subscription.unsubscribe()
            const id = (mockConnection.sentMessages[0].params as UnregistrationParams).unregisterations[0].id
            assert.deepStrictEqual(mockConnection.sentMessages, [
                {
                    method: UnregistrationRequest.type,
                    params: {
                        unregisterations: [{ id, method: ExecuteCommandRequest.type }],
                    } as UnregistrationParams,
                },
            ])

            assert.throws(() =>
                mockConnection.recvRequest(ExecuteCommandRequest.type, {
                    command: 'a',
                } as ExecuteCommandParams)
            )
            assert.strictEqual(called, false)
        })
    })
})
