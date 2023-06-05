import assert from 'assert'
import { Readable, Writable } from 'stream'

import { RecipeID } from '@sourcegraph/cody-shared/src/chat/recipes/recipe'
import { TranscriptJSON } from '@sourcegraph/cody-shared/src/chat/transcript'
import { ChatMessage } from '@sourcegraph/cody-shared/src/chat/transcript/messages'
import {
    ActiveTextEditor,
    ActiveTextEditorSelection,
    ActiveTextEditorVisibleContent,
} from '@sourcegraph/cody-shared/src/editor'

export interface RecipeInfo {
    id: RecipeID
    title: string
}

export interface StaticEditor {
    workspaceRoot: string | null
}

// Static recipe context that lots of recipes might need
// More context is obtained if necessary via server to client requests
export interface StaticRecipeContext {
    editor: StaticEditor
    firstInteraction: boolean
}

export interface ExecuteRecipeParams {
    id: RecipeID
    humanChatInput: string
    context: StaticRecipeContext
}

export interface ReplaceSelectionParams {
    fileName: string
    selectedText: string
    replacement: string
}

export interface ReplaceSelectionResult {
    applied: boolean
    failureReason: string
}

// TODO: Add some version info to prevent version incompatibilities
// TODO: Add capabilities so clients can announce what they can handle
export interface ClientInfo {
    name: string
}

export interface ServerInfo {
    name: string
}

// The RPC is packaged in the same way as LSP:
// Content-Length: lengthInBytes\r\n
// \r\n
// ...

// The RPC initialization process is the same as LSP:
// (-- Server process started; session begins --)
// Client: initialize request
// Server: initialize response
// Client: initialized notification
// Client and server send anything they want after this point
// The RPC shutdown process is the same as LSP:
// Client: shutdown request
// Server: shutdown response
// Client: exit notification
// (-- Server process exited; session ends --)
type Requests = {
    // Client -> Server
    initialize: [ClientInfo, ServerInfo]
    shutdown: [void, void]

    'recipes/list': [void, RecipeInfo[]]
    'recipes/execute': [ExecuteRecipeParams, void]

    // Server -> Client
    'editor/quickPick': [string[], string | null]
    'editor/prompt': [string, string | null]

    'editor/active': [void, ActiveTextEditor | null]
    'editor/selection': [void, ActiveTextEditorSelection | null]
    'editor/selectionOrEntireFile': [void, ActiveTextEditorSelection | null]
    'editor/visibleContent': [void, ActiveTextEditorVisibleContent | null]

    'intent/isCodebaseContextRequired': [string, boolean]
    'intent/isEditorContextRequired': [string, boolean]

    'editor/replaceSelection': [ReplaceSelectionParams, ReplaceSelectionResult]
}

type Notifications = {
    // Client -> Server
    initialized: [void]
    exit: [void]

    // Server -> Client
    'editor/warning': [string]

    'chat/updateMessageInProgress': [ChatMessage | null]
    'chat/updateTranscript': [TranscriptJSON]
}

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

    _write(chunk: Buffer, encoding: string, callback: (error?: Error | null) => void) {
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
                let newContentLength: number = 0

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
                            newContentLength = parseInt(headerValue)
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
                    } catch (err) {
                        this.callback(err, null)
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

    constructor() {
        super()
    }

    send(data: any) {
        this.pause()

        const content = Buffer.from(JSON.stringify(data), 'utf-8')
        const header = Buffer.from(`Content-Length: ${content.byteLength}\r\n\r\n`, 'utf-8')
        this.buffer = Buffer.concat([this.buffer, header, content])

        this.resume()
    }

    _read(size: number) {
        this.push(this.buffer.slice(0, size))
        this.buffer = this.buffer.slice(size)
    }
}

type RequestCallback<M extends RequestMethod> = (params: ParamsOf<M>) => Promise<ResultOf<M>>
type NotificationCallback<M extends NotificationMethod> = (params: ParamsOf<M>) => Promise<void>

export class MessageHandler {
    private id: number = 0
    private requestHandlers: Map<RequestMethod, RequestCallback<any>> = new Map()
    private notificationHandlers: Map<NotificationMethod, NotificationCallback<any>> = new Map()
    private responseHandlers: Map<Id, Function> = new Map()

    // TODO: RPC error handling
    public messageDecoder: MessageDecoder = new MessageDecoder((err: Error | null, msg: Message | null) => {
        if (err) {
            console.error(`Error: ${err}`)
        }
        if (!msg) return

        if (msg.id !== undefined && msg.method) {
            if (typeof msg.id === 'number' && msg.id > this.id) {
                this.id = msg.id + 1
            }

            // Requests have ids and methods
            const cb = this.requestHandlers.get(msg.method)
            if (cb) {
                cb(msg.params).then(result => {
                    this.messageEncoder.send({
                        jsonrpc: '2.0',
                        id: msg.id,
                        result,
                    } as ResponseMessage<any>)
                })
            } else {
                console.error(`No handler for request with method ${msg.method}`)
            }
        } else if (msg.id !== undefined) {
            // Responses have ids
            const cb = this.responseHandlers.get(msg.id)
            if (cb) {
                cb(msg.result)
                this.responseHandlers.delete(msg.id)
            } else {
                console.error(`No handler for response with id ${msg.id}`)
            }
        } else if (msg.method) {
            // Notifications have methods
            const cb = this.notificationHandlers.get(msg.method)
            if (cb) {
                cb(msg.params)
            } else {
                console.error(`No handler for notification with method ${msg.method}`)
            }
        }
    })

    public messageEncoder: MessageEncoder = new MessageEncoder()

    public registerRequest<M extends RequestMethod>(method: M, callback: RequestCallback<M>) {
        this.requestHandlers.set(method, callback)
    }

    public registerNotification<M extends NotificationMethod>(method: M, callback: NotificationCallback<M>) {
        this.notificationHandlers.set(method, callback)
    }

    public request<M extends RequestMethod>(method: M, params: ParamsOf<M>): Promise<ResultOf<M>> {
        const id = this.id++

        this.messageEncoder.send({
            jsonrpc: '2.0',
            id,
            method: method,
            params,
        } as RequestMessage<M>)

        return new Promise(resolve => {
            this.responseHandlers.set(id, resolve)
        })
    }

    public notify<M extends NotificationMethod>(method: M, params: ParamsOf<M>) {
        this.messageEncoder.send({
            jsonrpc: '2.0',
            method: method,
            params,
        } as NotificationMessage<M>)
    }
}
