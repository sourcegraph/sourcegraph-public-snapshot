import * as assert from 'assert'
import { createConnection as createClientConnection } from '../client/connection'
import { InitializedNotification, InitializeParams, InitializeRequest, InitializeResult } from '../protocol'
import { createMessageTransports } from '../test/integration/helpers'
import { activateExtension } from './extension'

describe('activateExtension', () => {
    it('initialize request parameters and result', async () => {
        const [clientTransports, serverTransports] = createMessageTransports()
        const clientConnection = createClientConnection(clientTransports)
        clientConnection.listen()

        const initParams: InitializeParams = {
            capabilities: { decoration: true },
            configurationCascade: { merged: {} },
        }
        const initResult: InitializeResult = {}

        const [, result] = await Promise.all([
            activateExtension<{}>(serverTransports, sourcegraph => {
                assert.deepStrictEqual(sourcegraph.initializeParams, initParams)
            }),
            clientConnection.sendRequest(InitializeRequest.type, initParams).then(result => {
                clientConnection.sendNotification(InitializedNotification.type, initParams)
                return result
            }),
        ])
        assert.deepStrictEqual(result, initResult)
    })
})
