import { Duplex } from 'stream'
import { MessageTransports } from '../../protocol/jsonrpc2/connection'
import { StreamMessageReader, StreamMessageWriter } from '../../protocol/jsonrpc2/transports/stream'

/** A bidirectional pipe. */
class TestStream extends Duplex {
    public _write(chunk: string, _encoding: string, done: () => void): void {
        this.emit('data', chunk)
        done()
    }
    public _read(_size: number): void {
        /* noop */
    }
}

/**
 * Creates a pair of message transports that are connected to each other. One can be used as the server and the
 * other as the client.
 */
export function createMessageTransports(): [MessageTransports, MessageTransports] {
    const up = new TestStream()
    const down = new TestStream()
    return [
        { reader: new StreamMessageReader(up), writer: new StreamMessageWriter(down) },
        { reader: new StreamMessageReader(down), writer: new StreamMessageWriter(up) },
    ]
}
