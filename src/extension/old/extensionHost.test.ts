import * as assert from 'assert'
import {
    InitializedNotification,
    InitializeParams,
    InitializeRequest,
    InitializeResult,
    RegistrationRequest,
} from '../../protocol'
import { createConnection } from '../../protocol/jsonrpc2/connection'
import { createMessageTransports } from '../../test/integration/helpers'
import { activateExtension } from './extensionHost'

describe('activateExtension (old)', () => {
    it('initialize request parameters and result', async () => {
        const [clientTransports, serverTransports] = createMessageTransports()
        const clientConnection = createConnection(clientTransports)
        clientConnection.listen()

        const initParams: InitializeParams = {
            capabilities: { decoration: true },
            configurationCascade: { merged: {} },
        }
        const initResult: InitializeResult = {}

        clientConnection.onRequest(RegistrationRequest.type, () => void 0)

        const [, result] = await Promise.all([
            activateExtension<{}>(sourcegraph => {
                assert.deepStrictEqual(sourcegraph.initializeParams, initParams)
            }, serverTransports),
            clientConnection.sendRequest(InitializeRequest.type, initParams).then(result => {
                clientConnection.sendNotification(InitializedNotification.type, initParams)
                return result
            }),
        ])
        assert.deepStrictEqual(result, initResult)
    })
})
