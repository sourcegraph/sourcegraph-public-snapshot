import assert from 'assert'
import { Readable, Writable } from 'stream'

import { Notifications, Requests } from './protocol'

export type RequestMethod = keyof Requests
export type NotificationMethod = keyof Notifications

export type Method = RequestMethod | keyof Notifications
export type ParamsOf<K extends Method> = (Requests & Notifications)[K][0]
export type ResultOf<K extends RequestMethod> = Requests[K][1]

export type Id = string | number

export enum ErrorCode {
    ParseError = -32700,
    InvalidRequest = -32600,
    MethodNotFound = -32601,
    InvalidParams = -32602,
    InternalError = -32603,
}

export interface ErrorInfo<T> {
    code: ErrorCode
    message: string
    data: T
}

export interface RequestMessage<M extends RequestMethod> {
    jsonrpc: '2.0'
    id: Id
    method: M
    params?: ParamsOf<M>
}

export interface ResponseMessage<M extends RequestMethod> {
    jsonrpc: '2.0'
    id: Id
    result?: ResultOf<M>
    error?: ErrorInfo<any>
}

export interface NotificationMessage<M extends NotificationMethod> {
    jsonrpc: '2.0'
    method: M
    params?: ParamsOf<M>
}

export type Message = RequestMessage<any> & ResponseMessage<any> & NotificationMessage<any>

export type MessageHandlerCallback = (err: Error | null, msg: Message | null) => void

export class MessageDecoder extends Writable {
    private buffer: Buffer = Buffer.alloc(0)
    private contentLengthRemaining: number | null = null
    private contentBuffer: Buffer = Buffer.alloc(0)

    constructor(public callback: MessageHandlerCallback) {
        super()
    }

    public _write(chunk: Buffer, encoding: string, callback: (error?: Error | null) => void): void {
        this.buffer = Buffer.concat([this.buffer, chunk])

        // We loop through as we could have a double message that requires processing twice
        read: while (true) {
            if (this.contentLengthRemaining === null) {
                const headerString = this.buffer.toString()

                let startIndex = 0
                let endIndex

                // We create this as we might get partial messages
                // so we only want to set the content length
                // once we get the whole thing
                let newContentLength = 0

                while ((endIndex = headerString.indexOf('\r\n', startIndex)) !== -1) {
                    const entry = headerString.slice(startIndex, endIndex)
                    const [headerName, headerValue] = entry.split(':').map(_ => _.trim())

                    if (headerValue === undefined) {
                        this.buffer = this.buffer.slice(endIndex + 2)

                        // Asserts we actually have a valid header with a Content-Length
                        // This state is irrecoverable because the stream is polluted
                        // Also what is the client doing 😭
                        this.contentLengthRemaining = newContentLength
                        assert(this.contentLengthRemaining !== null)
                        continue read
                    }

                    switch (headerName) {
                        case 'Content-Length':
                            newContentLength = parseInt(headerValue, 10)
                            break

                        default:
                            console.error(`Unknown header '${headerName}': ignoring!`)
                            break
                    }

                    startIndex = endIndex + 2
                }

                break
            } else {
                if (this.contentLengthRemaining === 0) {
                    try {
                        const data = JSON.parse(this.contentBuffer.toString())
                        this.contentBuffer = Buffer.alloc(0)
                        this.contentLengthRemaining = null
                        this.callback(null, data)
                    } catch (error) {
                        this.callback(error, null)
                    }

                    continue
                }

                const data = this.buffer.slice(0, this.contentLengthRemaining)
                this.contentBuffer = Buffer.concat([this.contentBuffer, data])
                this.buffer = this.buffer.slice(this.contentLengthRemaining)

                this.contentLengthRemaining -= data.byteLength
            }
        }

        callback()
    }
}

export class MessageEncoder extends Readable {
    private buffer: Buffer = Buffer.alloc(0)

    public send(data: any): void {
        this.pause()

        const content = Buffer.from(JSON.stringify(data), 'utf-8')
        const header = Buffer.from(`Content-Length: ${content.byteLength}\r\n\r\n`, 'utf-8')
        this.buffer = Buffer.concat([this.buffer, header, content])

        this.resume()
    }

    public _read(size: number): void {
        this.push(this.buffer.slice(0, size))
        this.buffer = this.buffer.slice(size)
    }
}

type RequestCallback<M extends RequestMethod> = (params: ParamsOf<M>) => Promise<ResultOf<M>>
type NotificationCallback<M extends NotificationMethod> = (params: ParamsOf<M>) => void

export class MessageHandler {
    private id = 0
    private requestHandlers: Map<RequestMethod, RequestCallback<any>> = new Map()
    private notificationHandlers: Map<NotificationMethod, NotificationCallback<any>> = new Map()
    private responseHandlers: Map<Id, (params: any) => void> = new Map()

    // TODO: RPC error handling
    public messageDecoder: MessageDecoder = new MessageDecoder((err: Error | null, msg: Message | null) => {
        if (err) {
            console.error(`Error: ${err}`)
        }
        if (!msg) {
            return
        }

        if (msg.id !== undefined && msg.method) {
            if (typeof msg.id === 'number' && msg.id > this.id) {
                this.id = msg.id + 1
            }

            // Requests have ids and methods
            const handler = this.requestHandlers.get(msg.method)
            if (handler) {
                handler(msg.params).then(
                    result => {
                        const data: ResponseMessage<any> = {
                            jsonrpc: '2.0',
                            id: msg.id,
                            // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment
                            result,
                        }
                        this.messageEncoder.send(data)
                    },
                    error => {
                        const message = error instanceof Error ? error.message : `${error}`
                        const data: ResponseMessage<any> = {
                            jsonrpc: '2.0',
                            id: msg.id,
                            error: {
                                code: ErrorCode.InternalError,
                                message,
                                data: JSON.stringify(error),
                            },
                        }
                        this.messageEncoder.send(data)
                    }
                )
            } else {
                console.error(`No handler for request with method ${msg.method}`)
            }
        } else if (msg.id !== undefined) {
            // Responses have ids
            const handler = this.responseHandlers.get(msg.id)
            if (handler) {
                handler(msg.result)
                this.responseHandlers.delete(msg.id)
            } else {
                console.error(`No handler for response with id ${msg.id}`)
            }
        } else if (msg.method) {
            // Notifications have methods
            const notificationHandler = this.notificationHandlers.get(msg.method)
            if (notificationHandler) {
                notificationHandler(msg.params)
            } else {
                console.error(`No handler for notification with method ${msg.method}`)
            }
        }
    })

    public messageEncoder: MessageEncoder = new MessageEncoder()

    public registerRequest<M extends RequestMethod>(method: M, callback: RequestCallback<M>): void {
        this.requestHandlers.set(method, callback)
    }

    public registerNotification<M extends NotificationMethod>(method: M, callback: NotificationCallback<M>): void {
        this.notificationHandlers.set(method, callback)
    }

    public request<M extends RequestMethod>(method: M, params: ParamsOf<M>): Promise<ResultOf<M>> {
        const id = this.id++

        const data: RequestMessage<M> = {
            jsonrpc: '2.0',
            id,
            method,
            params,
        }
        this.messageEncoder.send(data)

        return new Promise(resolve => {
            this.responseHandlers.set(id, resolve)
        })
    }

    public notify<M extends NotificationMethod>(method: M, params: ParamsOf<M>): void {
        const data: NotificationMessage<M> = {
            jsonrpc: '2.0',
            method,
            params,
        }
        this.messageEncoder.send(data)
    }
}
