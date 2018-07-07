import { MessageTransports } from '../connection'
import { Message } from '../messages'
import { AbstractMessageReader, AbstractMessageWriter, DataCallback, MessageReader, MessageWriter } from '../transports'

// Copied subset of WebSocket from the TypeScript "dom" core library to avoid needing to add that lib to
// tsconfig.json.
interface MessageEvent {
    data: string
}
interface WebSocketEventMap {
    close: never
    error: any
    message: MessageEvent
    open: {}
}
interface WebSocket {
    readonly readyState: number
    send(data: string): void
    addEventListener<K extends keyof WebSocketEventMap>(
        type: K,
        listener: (this: WebSocket, ev: WebSocketEventMap[K]) => any
    ): void
    removeEventListener<K extends keyof WebSocketEventMap>(
        type: K,
        listener: (this: WebSocket, ev: WebSocketEventMap[K]) => any
    ): void
    readonly OPEN: number
}

class WebSocketMessageReader extends AbstractMessageReader implements MessageReader {
    private pending: Message[] = []
    private callback: DataCallback | null = null

    constructor(socket: WebSocket) {
        super()

        socket.addEventListener('message', (e: MessageEvent) => {
            try {
                this.processMessage(e)
            } catch (err) {
                this.fireError(err)
            }
        })
        socket.addEventListener('error', err => this.fireError(err))
        socket.addEventListener('close', () => this.fireClose())
    }

    private processMessage(e: MessageEvent): void {
        const message: Message = JSON.parse(e.data)
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

    constructor(private socket: WebSocket) {
        super()
        socket.addEventListener('close', () => this.fireClose())
    }

    public write(message: Message): void {
        try {
            this.socket.send(JSON.stringify(message))
        } catch (err) {
            this.fireError(err, message, ++this.errorCount)
        }
    }
}

/** Creates JSON-RPC2 message transports for a WebSocket. */
export function createWebSocketMessageTransports(socket: WebSocket): Promise<MessageTransports> {
    const transports: MessageTransports = {
        reader: new WebSocketMessageReader(socket),
        writer: new WebSocketMessageWriter(socket),
    }
    if (socket.readyState === socket.OPEN) {
        return Promise.resolve(transports)
    }
    return new Promise<MessageTransports>((resolve, reject) => {
        function cleanup(): void {
            socket.removeEventListener('open', doResolve)
            socket.removeEventListener('close', doReject)
            socket.removeEventListener('error', doReject)
        }
        function doResolve(): void {
            cleanup()
            resolve(transports)
        }
        function doReject(err: any): void {
            cleanup()
            reject(err)
        }
        socket.addEventListener('open', doResolve)
        socket.addEventListener('close', doResolve)
        socket.addEventListener('error', doResolve)
    })
}
