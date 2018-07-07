import { Socket } from 'net'
import { Message } from '../messages'
import { AbstractMessageWriter, MessageWriter } from '../transports'
import { ContentLength, CRLF, StreamMessageReader } from './stream'

export class SocketMessageReader extends StreamMessageReader {
    public constructor(socket: Socket, encoding = 'utf-8') {
        super(socket as NodeJS.ReadableStream, encoding)
    }
}

export class SocketMessageWriter extends AbstractMessageWriter implements MessageWriter {
    private socket: Socket
    private queue: Message[]
    private sending: boolean
    private encoding: string
    private errorCount: number

    public constructor(socket: Socket, encoding = 'utf8') {
        super()
        this.socket = socket
        this.queue = []
        this.sending = false
        this.encoding = encoding
        this.errorCount = 0
        this.socket.on('error', (error: any) => this.fireError(error))
        this.socket.on('close', () => this.fireClose())
    }

    public write(msg: Message): void {
        if (!this.sending && this.queue.length === 0) {
            // See https://github.com/nodejs/node/issues/7657
            this.doWriteMessage(msg)
        } else {
            this.queue.push(msg)
        }
    }

    public doWriteMessage(msg: Message): void {
        const json = JSON.stringify(msg)
        const contentLength = Buffer.byteLength(json, this.encoding)

        const headers: string[] = [ContentLength, contentLength.toString(), CRLF, CRLF]
        try {
            // Header must be written in ASCII encoding
            this.sending = true
            this.socket.write(headers.join(''), 'ascii', (error: any) => {
                if (error) {
                    this.handleError(error, msg)
                }
                try {
                    // Now write the content. This can be written in any encoding
                    this.socket.write(json, this.encoding, (error: any) => {
                        this.sending = false
                        if (error) {
                            this.handleError(error, msg)
                        } else {
                            this.errorCount = 0
                        }
                        if (this.queue.length > 0) {
                            this.doWriteMessage(this.queue.shift()!)
                        }
                    })
                } catch (error) {
                    this.handleError(error, msg)
                }
            })
        } catch (error) {
            this.handleError(error, msg)
        }
    }

    private handleError(error: any, msg: Message): void {
        this.errorCount++
        this.fireError(error, msg, this.errorCount)
    }
}
