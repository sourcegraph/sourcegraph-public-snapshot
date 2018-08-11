import NodeWebSocket from 'ws'
import { createMessageConnection } from '../../src/jsonrpc2/connection'
import { createWebSocketMessageTransports } from '../../src/jsonrpc2/transports/nodeWebSocket'
import { InitializeParams, InitializeRequest, InitializeResult } from '../../src/protocol'
import config from './config'

async function run(): Promise<void> {
    console.log('Open.')
    const transports = await createWebSocketMessageTransports(
        new NodeWebSocket(config.url, {
            headers: { Authorization: `token ${config.accessToken}` },
        })
    )
    const connection = createMessageConnection(transports)
    connection.listen()
    connection.onError(err => console.error('Error:', err))
    connection.onClose(() => console.error('Connection closed.'))

    console.log('Initializing...')
    try {
        const initResult: InitializeResult = await connection.sendRequest(InitializeRequest.type, {
            root: config.root,
            initializationOptions: config.initializationOptions,
        } as InitializeParams)
        console.log('Initialized:', initResult)
    } catch (err) {
        console.error('initialize failed:', err.message)
        process.exit(1)
    }

    console.log('textDocument/hover...')
    try {
        const result = await connection.sendRequest<any>('textDocument/hover', {
            textDocument: { uri: `${config.root}#mux.go` },
            position: { character: 5, line: 23 },
        })
        console.log('textDocument/hover result:', result)
    } catch (err) {
        console.error('textDocument/hover failed:', err.message)
    }

    connection.unsubscribe()
    process.exit(0)
}
run().then(null, err => {
    console.error(err)
    process.exit(1)
})
