/**
 * A language server message
 */
export interface Message {
    jsonrpc: string
}

/**
 * Request message
 */
export interface RequestMessage extends Message {
    /**
     * The request id.
     */
    id: number | string

    /**
     * The method to be invoked.
     */
    method: string

    /**
     * The method's params.
     */
    params?: any
}

/**
 * Predefined error codes.
 */
export namespace ErrorCodes {
    // Defined by JSON-RPC 2.0.
    export const ParseError = -32700
    export const InvalidRequest = -32600
    export const MethodNotFound = -32601
    export const InvalidParams = -32602
    export const InternalError = -32603
    export const serverErrorStart = -32099
    export const serverErrorEnd = -32000
    export const ServerNotInitialized = -32002
    export const UnknownErrorCode = -32001

    // Defined by the protocol.
    export const RequestCancelled = -32800

    // Defined by this library.
    export const MessageWriteError = 1
    export const MessageReadError = 2
}

interface ResponseErrorLiteral<D> {
    /**
     * A number indicating the error type that occured.
     */
    code: number

    /**
     * A string providing a short decription of the error.
     */
    message: string

    /**
     * A Primitive or Structured value that contains additional
     * information about the error. Can be omitted.
     */
    data?: D
}

/**
 * An error object return in a response in case a request
 * has failed.
 */
export class ResponseError<D> extends Error {
    public readonly code: number
    public readonly data: D | undefined

    constructor(code: number, message: string, data?: D) {
        super(message)
        this.code = typeof code === 'number' ? code : ErrorCodes.UnknownErrorCode
        this.data = data
        Object.setPrototypeOf(this, ResponseError.prototype)
    }

    public toJSON(): ResponseErrorLiteral<D> {
        return {
            code: this.code,
            message: this.message,
            data: this.data,
        }
    }
}

/**
 * A response message.
 */
export interface ResponseMessage extends Message {
    /**
     * The request id.
     */
    id: number | string | null

    /**
     * The result of a request. This can be omitted in
     * the case of an error.
     */
    result?: any

    /**
     * The error object in case a request fails.
     */
    error?: ResponseErrorLiteral<any>
}

/**
 * Notification Message
 */
export interface NotificationMessage extends Message {
    /**
     * The method to be invoked.
     */
    method: string

    /**
     * The notification's params.
     */
    params?: any
}

/**
 * Tests if the given message is a request message
 */
export function isRequestMessage(message: Message | undefined): message is RequestMessage {
    const candidate = message as RequestMessage
    return (
        candidate &&
        typeof candidate.method === 'string' &&
        (typeof candidate.id === 'string' || typeof candidate.id === 'number')
    )
}

/**
 * Tests if the given message is a notification message
 */
export function isNotificationMessage(message: Message | undefined): message is NotificationMessage {
    const candidate = message as NotificationMessage
    return candidate && typeof candidate.method === 'string' && (message as any).id === void 0
}

/**
 * Tests if the given message is a response message
 */
export function isResponseMessage(message: Message | undefined): message is ResponseMessage {
    const candidate = message as ResponseMessage
    return (
        candidate &&
        (candidate.result !== void 0 || !!candidate.error) &&
        (typeof candidate.id === 'string' || typeof candidate.id === 'number' || candidate.id === null)
    )
}
