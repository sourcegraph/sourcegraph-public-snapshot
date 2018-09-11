import * as assert from 'assert'
import * as sourcegraph from 'sourcegraph'
import {
    InitializedNotification,
    InitializeParams,
    InitializeRequest,
    InitializeResult,
    RegistrationRequest,
} from '../protocol'
import { createMessageConnection } from '../protocol/jsonrpc2/connection'
import { createMessageTransports } from '../test/integration/helpers'
import { createExtensionHost } from './extensionHost'

describe('createExtensionHost', () => {
    it('initialize request parameters and result', async () => {
        const [clientTransports, serverTransports] = createMessageTransports()
        const clientConnection = createMessageConnection(clientTransports)
        clientConnection.listen()

        const initParams: InitializeParams = {
            capabilities: { decoration: true, experimental: { a: 1 } },
            configurationCascade: { merged: {} },
        }
        const initResult: InitializeResult = {}

        clientConnection.onRequest(RegistrationRequest.type, () => void 0)

        const [, result] = await Promise.all([
            createExtensionHost(serverTransports).then((extensionHost: typeof sourcegraph) => {
                assert.ok(extensionHost)
            }),
            clientConnection.sendRequest(InitializeRequest.type, initParams).then(result => {
                clientConnection.sendNotification(InitializedNotification.type, initParams)
                return result
            }),
        ])
        assert.deepStrictEqual(result, initResult)
    })
})
