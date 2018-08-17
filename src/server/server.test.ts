import * as assert from 'assert'
import { createConnection as createClientConnection } from '../client/connection'
import { InitializeParams, InitializeRequest, InitializeResult } from '../protocol'
import { createMessageTransports } from '../test/integration/helpers'
import { createConnection as createServerConnection } from './server'

describe('Connection', () => {
    it('initialize request parameters and result', async () => {
        const [clientTransports, serverTransports] = createMessageTransports()
        const serverConnection = createServerConnection(serverTransports)
        const clientConnection = createClientConnection(clientTransports)
        serverConnection.listen()
        clientConnection.listen()

        const initParams: InitializeParams = {
            root: null,
            capabilities: { decoration: true },
            configurationCascade: { merged: {} },
            workspaceFolders: null,
        }
        const initResult: InitializeResult = {
            capabilities: { contributions: { actions: [{ id: 'c', command: 'c' }] } },
        }

        serverConnection.onRequest(InitializeRequest.type, params => {
            assert.deepStrictEqual(params, initParams)
            return initResult
        })
        const result = await clientConnection.sendRequest(InitializeRequest.type, initParams)
        assert.deepStrictEqual(result, initResult)
    })
})
