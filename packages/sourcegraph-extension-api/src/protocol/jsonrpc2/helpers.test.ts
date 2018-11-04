import { MessageTransports } from './connection'
import { Message } from './messages'
import { AbstractMessageReader, AbstractMessageWriter, DataCallback, MessageReader, MessageWriter } from './transport'

/**
 * Creates a pair of message transports that are connected to each other. One can be used as the server and the
 * other as the client.
 */
export function createMessageTransports(): [MessageTransports, MessageTransports] {
    const { reader: reader1, writer: writer1 } = createMessagePipe()
    const { reader: reader2, writer: writer2 } = createMessagePipe()
    return [{ reader: reader1, writer: writer2 }, { reader: reader2, writer: writer1 }]
}

/** Creates a single set of transports that are connected to each other. */
export function createMessagePipe(): MessageTransports {
    let readerCallback: DataCallback | undefined
    const reader: MessageReader = new class extends AbstractMessageReader implements MessageReader {
        public listen(callback: DataCallback): void {
            readerCallback = callback
        }
    }()
    const writer: MessageWriter = new class extends AbstractMessageWriter implements MessageWriter {
        public write(msg: Message): void {
            if (readerCallback) {
                readerCallback(msg)
            } else {
                throw new Error('reader has no listener')
            }
        }
    }()
    return { reader, writer }
}
