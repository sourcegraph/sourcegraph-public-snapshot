import * as assert from 'assert'
import { Duplex } from 'stream'
import { createConnection as createClientConnection } from '../client/connection'
import { StreamMessageReader, StreamMessageWriter } from '../jsonrpc2/transports/stream'
import { InitializeParams, InitializeRequest, InitializeResult } from '../protocol'
import { createConnection as createServerConnection } from './server'

class TestStream extends Duplex {
    public _write(chunk: string, _encoding: string, done: () => void): void {
        this.emit('data', chunk)
        done()
    }
    public _read(_size: number): void {
        /* noop */
    }
}

describe('Connection', () => {
    it('initialize request parameters and result', async () => {
        const up = new TestStream()
        const down = new TestStream()
        const serverConnection = createServerConnection({
            reader: new StreamMessageReader(up),
            writer: new StreamMessageWriter(down),
        })
        const clientConnection = createClientConnection({
            reader: new StreamMessageReader(down),
            writer: new StreamMessageWriter(up),
        })
        serverConnection.listen()
        clientConnection.listen()

        const initParams: InitializeParams = {
            root: null,
            capabilities: { decoration: { static: true } },
            workspaceFolders: null,
        }
        const initResult: InitializeResult = { capabilities: { contributions: { commands: [{ command: 'c' }] } } }

        serverConnection.onRequest(InitializeRequest.type, params => {
            assert.deepStrictEqual(params, initParams)
            return initResult
        })
        const result = await clientConnection.sendRequest(InitializeRequest.type, initParams)
        assert.deepStrictEqual(result, initResult)
    })
})
