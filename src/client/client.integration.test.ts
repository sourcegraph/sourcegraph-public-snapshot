import * as assert from 'assert'
import { ClientCapabilities, InitializeParams, InitializeRequest, InitializeResult } from '../protocol'
import { Connection, createConnection, MessageTransports } from '../protocol/jsonrpc2/connection'
import { Trace } from '../protocol/jsonrpc2/trace'
import { clientStateIsActive, getClientState } from '../test/helpers'
import { createMessageTransports } from '../test/integration/helpers'
import { Client, ClientState } from './client'

const createClientTransportsForTestServer = (registerServer: (server: Connection) => void): MessageTransports => {
    const [clientTransports, serverTransports] = createMessageTransports()
    const serverConnection = createConnection(serverTransports)
    serverConnection.listen()
    registerServer(serverConnection)
    return clientTransports
}

describe('Client', () => {
    it('registers features, activates, initializes, stops, and reactivates', async () => {
        const initResult: InitializeResult = {}
        const testNotificationParams = { a: 1 }
        const testRequestParams = { b: 2 }
        const testRequestResult = { c: 3 }

        // Create test server.
        let serverInitialized!: Promise<void>
        let serverReceivedTestNotification!: Promise<void>
        let serverReceivedTestRequest!: Promise<void>
        const createMessageTransports = () =>
            createClientTransportsForTestServer(server => {
                serverInitialized = new Promise<void>((resolve, reject) => {
                    server.onRequest(InitializeRequest.type, params => {
                        try {
                            assert.deepStrictEqual(params, {
                                capabilities: { experimental: 'test' },
                                configurationCascade: { merged: {} },
                                trace: Trace.toString(Trace.Off),
                            } as InitializeParams)
                            resolve()
                        } catch (err) {
                            reject(err)
                        }
                        return initResult
                    })
                })
                serverReceivedTestNotification = new Promise<void>((resolve, reject) => {
                    server.onNotification('test', params => {
                        try {
                            assert.deepStrictEqual(params, testNotificationParams)
                            resolve()
                        } catch (err) {
                            reject(err)
                        }
                    })
                })
                serverReceivedTestRequest = new Promise<void>((resolve, reject) => {
                    server.onRequest('test', params => {
                        try {
                            assert.deepStrictEqual(params, testRequestParams)
                            resolve()
                        } catch (err) {
                            reject(err)
                        }
                        return testRequestResult
                    })
                })
            })

        const checkClient = async (client: Client): Promise<void> => {
            assert.strictEqual(getClientState(client), ClientState.Connecting)

            await Promise.all([clientStateIsActive(client), serverInitialized])
            assert.deepStrictEqual(client.initializeResult, initResult)

            client.sendNotification('test', testNotificationParams)
            await serverReceivedTestNotification

            await client.sendRequest('test', testRequestParams)
            await serverReceivedTestRequest

            client.onNotification('test', () => void 0)
            client.onRequest('test', () => void 0)

            assert.ok(client.needsStop())
            client.trace = Trace.Messages
            client.trace = Trace.Verbose
            client.trace = Trace.Off
        }

        // Create test client.
        const client = new Client('', { createMessageTransports })
        client.registerFeature({
            fillClientCapabilities: (capabilities: ClientCapabilities) => (capabilities.experimental = 'test'),
        })
        assert.strictEqual(getClientState(client), ClientState.Initial)

        // Activate client.
        client.activate()
        assert.strictEqual(client.initializeResult, null)
        await checkClient(client)

        // Stop client and check that it reports itself as being stopped.
        await client.stop()
        assert.strictEqual(getClientState(client), ClientState.Stopped)
        assert.strictEqual(client.needsStop(), false)

        // Stop client again (noop because the client is already stopped).
        await client.stop()
        assert.strictEqual(getClientState(client), ClientState.Stopped)

        // Reactivate client.
        client.activate()
        await checkClient(client)

        client.unsubscribe()
    })
})
