import NodeWebSocket from 'ws'
import { MessageTransports } from '../connection'
import { Message } from '../messages'
import { AbstractMessageReader, AbstractMessageWriter, DataCallback, MessageReader, MessageWriter } from '../transport'

class WebSocketMessageReader extends AbstractMessageReader implements MessageReader {
    private pending: Message[] = []
    private callback: DataCallback | null = null

    constructor(socket: NodeWebSocket) {
        super()

        socket.on('message', data => {
            try {
                this.processMessage(data)
            } catch (err) {
                this.fireError(err)
            }
        })
        socket.on('error', err => this.fireError(err))
        socket.on('close', () => this.fireClose())
    }

    private processMessage(data: NodeWebSocket.Data): void {
        const message: Message = JSON.parse(data.toString())
        if (this.callback) {
            this.callback(message)
        } else {
            this.pending.push(message)
        }
    }

    public listen(callback: DataCallback): void {
        if (this.callback) {
            throw new Error('callback is already set')
        }
        this.callback = callback
        while (this.pending.length !== 0) {
            callback(this.pending.pop()!)
        }
    }

    public stop(): void {
        this.callback = null
    }
}

class WebSocketMessageWriter extends AbstractMessageWriter implements MessageWriter {
    private errorCount = 0

    constructor(private socket: NodeWebSocket) {
        super()
        socket.on('close', () => this.fireClose())
    }

    public write(message: Message): void {
        try {
            this.socket.send(JSON.stringify(message), err => {
                if (err) {
                    this.fireError(err, message, ++this.errorCount)
                }
            })
        } catch (err) {
            this.fireError(err, message, ++this.errorCount)
        }
    }
}

/** Creates JSON-RPC2 message transports for a NodeWebSocket. */
export function createWebSocketMessageTransports(socket: NodeWebSocket): Promise<MessageTransports> {
    const transports: MessageTransports = {
        reader: new WebSocketMessageReader(socket),
        writer: new WebSocketMessageWriter(socket),
    }
    if (socket.readyState === socket.OPEN) {
        return Promise.resolve(transports)
    }
    return new Promise<MessageTransports>((resolve, reject) => {
        socket.prependOnceListener('open', () => {
            socket.off('error', reject)
            resolve(transports)
        })
        socket.prependOnceListener('error', reject)
    })
}
