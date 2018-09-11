import { Message } from '../messages'
import { AbstractMessageReader, AbstractMessageWriter, DataCallback, MessageReader, MessageWriter } from '../transport'

const DefaultSize = 8192
const CR: number = Buffer.from('\r', 'ascii')[0]
const LF: number = Buffer.from('\n', 'ascii')[0]
const CRLF = '\r\n'
const ContentLength = 'Content-Length: '

class MessageBuffer {
    private encoding: string
    private index: number
    private buffer: Buffer

    constructor(encoding = 'utf8') {
        this.encoding = encoding
        this.index = 0
        this.buffer = Buffer.alloc(DefaultSize)
    }

    public append(chunk: Buffer | string): void {
        let toAppend: Buffer = chunk as Buffer
        if (typeof chunk === 'string') {
            const str = chunk as string
            const bufferLen = Buffer.byteLength(str, this.encoding)
            toAppend = Buffer.alloc(bufferLen)
            toAppend.write(str, 0, bufferLen, this.encoding)
        }
        if (this.buffer.length - this.index >= toAppend.length) {
            toAppend.copy(this.buffer, this.index, 0, toAppend.length)
        } else {
            const newSize = (Math.ceil((this.index + toAppend.length) / DefaultSize) + 1) * DefaultSize
            if (this.index === 0) {
                this.buffer = Buffer.alloc(newSize)
                toAppend.copy(this.buffer, 0, 0, toAppend.length)
            } else {
                this.buffer = Buffer.concat([this.buffer.slice(0, this.index), toAppend], newSize)
            }
        }
        this.index += toAppend.length
    }

    public tryReadHeaders(): { [key: string]: string } | undefined {
        let result: { [key: string]: string } | undefined
        let current = 0
        while (
            current + 3 < this.index &&
            (this.buffer[current] !== CR ||
                this.buffer[current + 1] !== LF ||
                this.buffer[current + 2] !== CR ||
                this.buffer[current + 3] !== LF)
        ) {
            current++
        }
        // No header / body separator found (e.g CRLFCRLF)
        if (current + 3 >= this.index) {
            return result
        }
        result = Object.create(null)
        const headers = this.buffer.toString('ascii', 0, current).split(CRLF)
        for (const header of headers) {
            const index: number = header.indexOf(':')
            if (index === -1) {
                throw new Error('Message header must separate key and value using :')
            }
            const key = header.substr(0, index)
            const value = header.substr(index + 1).trim()
            result![key] = value
        }

        const nextStart = current + 4
        this.buffer = this.buffer.slice(nextStart)
        this.index = this.index - nextStart
        return result
    }

    public tryReadContent(length: number): string | null {
        if (this.index < length) {
            return null
        }
        const result = this.buffer.toString(this.encoding, 0, length)
        const nextStart = length
        this.buffer.copy(this.buffer, 0, nextStart)
        this.index = this.index - nextStart
        return result
    }

    public get numberOfBytes(): number {
        return this.index
    }
}

export class StreamMessageReader extends AbstractMessageReader implements MessageReader {
    private readable: NodeJS.ReadableStream
    private callback?: DataCallback
    private buffer: MessageBuffer
    private nextMessageLength?: number
    private messageToken?: number
    private partialMessageTimer: NodeJS.Timer | undefined
    private _partialMessageTimeout: number

    public constructor(readable: NodeJS.ReadableStream, encoding = 'utf8') {
        super()
        this.readable = readable
        this.buffer = new MessageBuffer(encoding)
        this._partialMessageTimeout = 10000
    }

    public set partialMessageTimeout(timeout: number) {
        this._partialMessageTimeout = timeout
    }

    public get partialMessageTimeout(): number {
        return this._partialMessageTimeout
    }

    public listen(callback: DataCallback): void {
        this.nextMessageLength = -1
        this.messageToken = 0
        this.partialMessageTimer = undefined
        this.callback = callback
        this.readable.on('data', (data: Buffer) => {
            this.onData(data)
        })
        this.readable.on('error', (error: any) => this.fireError(error))
        this.readable.on('close', () => this.fireClose())
    }

    private onData(data: Buffer | string): void {
        this.buffer.append(data)
        while (true) {
            if (this.nextMessageLength === -1) {
                const headers = this.buffer.tryReadHeaders()
                if (!headers) {
                    return
                }
                const contentLength = headers['Content-Length']
                if (!contentLength) {
                    throw new Error('Header must provide a Content-Length property.')
                }
                const length = parseInt(contentLength, 10)
                if (isNaN(length)) {
                    throw new Error('Content-Length value must be a number.')
                }
                this.nextMessageLength = length
                // Take the encoding form the header. For compatibility
                // treat both utf-8 and utf8 as node utf8
            }
            const msg = this.buffer.tryReadContent(this.nextMessageLength!)
            if (msg === null) {
                /** We haven't recevied the full message yet. */
                this.setPartialMessageTimer()
                return
            }
            this.clearPartialMessageTimer()
            this.nextMessageLength = -1
            this.messageToken!++
            const json = JSON.parse(msg)
            this.callback!(json)
        }
    }

    private clearPartialMessageTimer(): void {
        if (this.partialMessageTimer) {
            clearTimeout(this.partialMessageTimer)
            this.partialMessageTimer = undefined
        }
    }

    private setPartialMessageTimer(): void {
        this.clearPartialMessageTimer()
        if (this._partialMessageTimeout <= 0) {
            return
        }
        this.partialMessageTimer = setTimeout(
            (token, timeout) => {
                this.partialMessageTimer = undefined
                if (token === this.messageToken) {
                    this.firePartialMessage({ messageToken: token, waitingTime: timeout })
                    this.setPartialMessageTimer()
                }
            },
            this._partialMessageTimeout,
            this.messageToken,
            this._partialMessageTimeout
        )
    }

    public unsubscribe(): void {
        super.unsubscribe()
        this.clearPartialMessageTimer()
    }
}

export class StreamMessageWriter extends AbstractMessageWriter implements MessageWriter {
    private writable: NodeJS.WritableStream
    private encoding: string
    private errorCount: number

    public constructor(writable: NodeJS.WritableStream, encoding = 'utf8') {
        super()
        this.writable = writable
        this.encoding = encoding
        this.errorCount = 0
        this.writable.on('error', (error: any) => this.fireError(error))
        this.writable.on('close', () => this.fireClose())
    }

    public write(msg: Message): void {
        const json = JSON.stringify(msg)
        const contentLength = Buffer.byteLength(json, this.encoding)

        const headers: string[] = [ContentLength, contentLength.toString(), CRLF, CRLF]
        try {
            // Header must be written in ASCII encoding
            this.writable.write(headers.join(''), 'ascii')
            // Now write the content. This can be written in any encoding
            this.writable.write(json, this.encoding)
            this.errorCount = 0
        } catch (error) {
            this.errorCount++
            this.fireError(error, msg, this.errorCount)
        }
    }
}
