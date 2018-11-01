import { Unsubscribable } from 'rxjs'
import { CancelNotification, CancelParams } from './cancel'
import { CancellationToken, CancellationTokenSource } from './cancel'
import { ConnectionStrategy } from './connectionStrategy'
import { Emitter, Event } from './events'
import {
    GenericNotificationHandler,
    GenericRequestHandler,
    StarNotificationHandler,
    StarRequestHandler,
} from './handlers'
import { LinkedMap } from './linkedMap'
import {
    ErrorCodes,
    isNotificationMessage,
    isRequestMessage,
    isResponseMessage,
    Message,
    NotificationMessage,
    RequestMessage,
    ResponseError,
    ResponseMessage,
} from './messages'
import { noopTracer, Trace, Tracer } from './trace'
import { DataCallback, MessageReader, MessageWriter } from './transport'

// Copied from vscode-languageserver to avoid adding extraneous dependencies.

export interface Logger {
    error(message: string): void
    warn(message: string): void
    info(message: string): void
    log(message: string): void
}

const NullLogger: Logger = Object.freeze({
    error: () => {
        /* noop */
    },
    warn: () => {
        /* noop */
    },
    info: () => {
        /* noop */
    },
    log: () => {
        /* noop */
    },
})

export enum ConnectionErrors {
    /**
     * The connection is closed.
     */
    Closed = 1,
    /**
     * The connection got unsubscribed (i.e., disposed).
     */
    Unsubscribed = 2,
    /**
     * The connection is already in listening mode.
     */
    AlreadyListening = 3,
}

export class ConnectionError extends Error {
    public readonly code: ConnectionErrors

    constructor(code: ConnectionErrors, message: string) {
        super(message)
        this.code = code
        Object.setPrototypeOf(this, ConnectionError.prototype)
    }
}

type MessageQueue = LinkedMap<string, Message>

export interface Connection extends Unsubscribable {
    sendRequest<R>(method: string, params?: any): Promise<R>
    sendRequest<R>(method: string, ...params: any[]): Promise<R>

    onRequest<R, E>(method: string, handler: GenericRequestHandler<R, E>): void
    onRequest(handler: StarRequestHandler): void

    sendNotification(method: string, ...params: any[]): void

    onNotification(method: string, handler: GenericNotificationHandler): void
    onNotification(handler: StarNotificationHandler): void

    trace(value: Trace, tracer: Tracer): void

    onError: Event<[Error, Message | undefined, number | undefined]>
    onClose: Event<void>
    onUnhandledNotification: Event<NotificationMessage>
    listen(): void
    onUnsubscribe: Event<void>
}

export interface MessageTransports {
    reader: MessageReader
    writer: MessageWriter
}

export function createConnection(
    transports: MessageTransports,
    logger?: Logger,
    strategy?: ConnectionStrategy
): Connection {
    if (!logger) {
        logger = NullLogger
    }
    return _createConnection(transports, logger, strategy)
}

interface ResponsePromise {
    /** The request's method. */
    method: string

    /** Only set in Trace.Verbose mode. */
    request?: RequestMessage

    /** The timestamp when the request was received. */
    timerStart: number

    resolve: (response: any) => void
    reject: (error: any) => void
}

enum ConnectionState {
    New = 1,
    Listening = 2,
    Closed = 3,
    Unsubscribed = 4,
}

interface RequestHandlerElement {
    type: string | undefined
    handler: GenericRequestHandler<any, any>
}

interface NotificationHandlerElement {
    type: string | undefined
    handler: GenericNotificationHandler
}

function _createConnection(transports: MessageTransports, logger: Logger, strategy?: ConnectionStrategy): Connection {
    let sequenceNumber = 0
    let notificationSquenceNumber = 0
    let unknownResponseSquenceNumber = 0
    const version = '2.0'

    let starRequestHandler: StarRequestHandler | undefined
    const requestHandlers: { [name: string]: RequestHandlerElement | undefined } = Object.create(null)
    let starNotificationHandler: StarNotificationHandler | undefined
    const notificationHandlers: { [name: string]: NotificationHandlerElement | undefined } = Object.create(null)

    let timer: NodeJS.Timer | undefined
    let messageQueue: MessageQueue = new LinkedMap<string, Message>()
    let responsePromises: { [name: string]: ResponsePromise } = Object.create(null)
    let requestTokens: { [id: string]: CancellationTokenSource } = Object.create(null)

    let trace: Trace = Trace.Off
    let tracer: Tracer = noopTracer

    let state: ConnectionState = ConnectionState.New
    const errorEmitter = new Emitter<[Error, Message | undefined, number | undefined]>()
    const closeEmitter: Emitter<void> = new Emitter<void>()
    const unhandledNotificationEmitter: Emitter<NotificationMessage> = new Emitter<NotificationMessage>()

    const unsubscribeEmitter: Emitter<void> = new Emitter<void>()

    function createRequestQueueKey(id: string | number): string {
        return 'req-' + id.toString()
    }

    function createResponseQueueKey(id: string | number | null): string {
        if (id === null) {
            return 'res-unknown-' + (++unknownResponseSquenceNumber).toString()
        } else {
            return 'res-' + id.toString()
        }
    }

    function createNotificationQueueKey(): string {
        return 'not-' + (++notificationSquenceNumber).toString()
    }

    function addMessageToQueue(queue: MessageQueue, message: Message): void {
        if (isRequestMessage(message)) {
            queue.set(createRequestQueueKey(message.id), message)
        } else if (isResponseMessage(message)) {
            queue.set(createResponseQueueKey(message.id), message)
        } else {
            queue.set(createNotificationQueueKey(), message)
        }
    }

    function cancelUndispatched(_message: Message): ResponseMessage | undefined {
        return undefined
    }

    function isListening(): boolean {
        return state === ConnectionState.Listening
    }

    function isClosed(): boolean {
        return state === ConnectionState.Closed
    }

    function isUnsubscribed(): boolean {
        return state === ConnectionState.Unsubscribed
    }

    function closeHandler(): void {
        if (state === ConnectionState.New || state === ConnectionState.Listening) {
            state = ConnectionState.Closed
            closeEmitter.fire(undefined)
        }
        // If the connection is unsubscribed don't sent close events.
    }

    function readErrorHandler(error: Error): void {
        errorEmitter.fire([error, undefined, undefined])
    }

    function writeErrorHandler(data: [Error, Message | undefined, number | undefined]): void {
        errorEmitter.fire(data)
    }

    transports.reader.onClose(closeHandler)
    transports.reader.onError(readErrorHandler)

    transports.writer.onClose(closeHandler)
    transports.writer.onError(writeErrorHandler)

    function triggerMessageQueue(): void {
        if (timer || messageQueue.size === 0) {
            return
        }
        timer = setImmediateCompat(() => {
            timer = undefined
            processMessageQueue()
        })
    }

    function processMessageQueue(): void {
        if (messageQueue.size === 0) {
            return
        }
        const message = messageQueue.shift()!
        try {
            if (isRequestMessage(message)) {
                handleRequest(message)
            } else if (isNotificationMessage(message)) {
                handleNotification(message)
            } else if (isResponseMessage(message)) {
                handleResponse(message)
            } else {
                handleInvalidMessage(message)
            }
        } finally {
            triggerMessageQueue()
        }
    }

    const callback: DataCallback = message => {
        try {
            // We have received a cancellation message. Check if the message is still in the queue and cancel it if
            // allowed to do so.
            if (isNotificationMessage(message) && message.method === CancelNotification.type) {
                const key = createRequestQueueKey((message.params as CancelParams).id)
                const toCancel = messageQueue.get(key)
                if (isRequestMessage(toCancel)) {
                    const response =
                        strategy && strategy.cancelUndispatched
                            ? strategy.cancelUndispatched(toCancel, cancelUndispatched)
                            : cancelUndispatched(toCancel)
                    if (response && (response.error !== undefined || response.result !== undefined)) {
                        messageQueue.delete(key)
                        response.id = toCancel.id
                        tracer.responseCanceled(response, toCancel, message)
                        transports.writer.write(response)
                        return
                    }
                }
            }
            addMessageToQueue(messageQueue, message)
        } finally {
            triggerMessageQueue()
        }
    }

    function handleRequest(requestMessage: RequestMessage): void {
        if (isUnsubscribed()) {
            // we return here silently since we fired an event when the
            // connection got unsubscribed.
            return
        }

        const startTime = Date.now()

        function reply(resultOrError: any | ResponseError<any>): void {
            const message: ResponseMessage = {
                jsonrpc: version,
                id: requestMessage.id,
            }
            if (resultOrError instanceof ResponseError) {
                message.error = (resultOrError as ResponseError<any>).toJSON()
            } else {
                message.result = resultOrError === undefined ? null : resultOrError
            }
            tracer.responseSent(message, requestMessage, startTime)
            transports.writer.write(message)
        }
        function replyError(error: ResponseError<any>): void {
            const message: ResponseMessage = {
                jsonrpc: version,
                id: requestMessage.id,
                error: error.toJSON(),
            }
            tracer.responseSent(message, requestMessage, startTime)
            transports.writer.write(message)
        }
        function replySuccess(result: any): void {
            // The JSON RPC defines that a response must either have a result or an error
            // So we can't treat undefined as a valid response result.
            if (result === undefined) {
                result = null
            }
            const message: ResponseMessage = {
                jsonrpc: version,
                id: requestMessage.id,
                result,
            }
            tracer.responseSent(message, requestMessage, startTime)
            transports.writer.write(message)
        }

        tracer.requestReceived(requestMessage)

        const element = requestHandlers[requestMessage.method]
        const requestHandler: GenericRequestHandler<any, any> | undefined = element && element.handler
        if (requestHandler || starRequestHandler) {
            const cancellationSource = new CancellationTokenSource()
            const tokenKey = String(requestMessage.id)
            requestTokens[tokenKey] = cancellationSource
            try {
                const params = requestMessage.params !== undefined ? requestMessage.params : null
                const handlerResult = requestHandler
                    ? requestHandler(params, cancellationSource.token)
                    : starRequestHandler!(requestMessage.method, params, cancellationSource.token)

                const promise = handlerResult as Promise<any | ResponseError<any>>
                if (!handlerResult) {
                    delete requestTokens[tokenKey]
                    replySuccess(handlerResult)
                } else if (promise.then) {
                    promise.then(
                        (resultOrError): any | ResponseError<any> => {
                            delete requestTokens[tokenKey]
                            reply(resultOrError)
                        },
                        error => {
                            delete requestTokens[tokenKey]
                            if (error instanceof ResponseError) {
                                replyError(error as ResponseError<any>)
                            } else if (error && typeof error.message === 'string') {
                                replyError(
                                    new ResponseError<void>(ErrorCodes.InternalError, error.message, {
                                        stack: error.stack,
                                        ...error,
                                    })
                                )
                            } else {
                                replyError(
                                    new ResponseError<void>(
                                        ErrorCodes.InternalError,
                                        `Request ${
                                            requestMessage.method
                                        } failed unexpectedly without providing any details.`
                                    )
                                )
                            }
                        }
                    )
                } else {
                    delete requestTokens[tokenKey]
                    reply(handlerResult)
                }
            } catch (error) {
                delete requestTokens[tokenKey]
                if (error instanceof ResponseError) {
                    reply(error as ResponseError<any>)
                } else if (error && typeof error.message === 'string') {
                    replyError(
                        new ResponseError<void>(ErrorCodes.InternalError, error.message, {
                            stack: error.stack,
                            ...error,
                        })
                    )
                } else {
                    replyError(
                        new ResponseError<void>(
                            ErrorCodes.InternalError,
                            `Request ${requestMessage.method} failed unexpectedly without providing any details.`
                        )
                    )
                }
            }
        } else {
            replyError(new ResponseError<void>(ErrorCodes.MethodNotFound, `Unhandled method ${requestMessage.method}`))
        }
    }

    function handleResponse(responseMessage: ResponseMessage): void {
        if (isUnsubscribed()) {
            // See handle request.
            return
        }

        if (responseMessage.id === null) {
            if (responseMessage.error) {
                logger.error(
                    `Received response message without id: Error is: \n${JSON.stringify(
                        responseMessage.error,
                        undefined,
                        4
                    )}`
                )
            } else {
                logger.error(`Received response message without id. No further error information provided.`)
            }
        } else {
            const key = String(responseMessage.id)
            const responsePromise = responsePromises[key]
            if (responsePromise) {
                tracer.responseReceived(
                    responseMessage,
                    responsePromise.request || responsePromise.method,
                    responsePromise.timerStart
                )
                delete responsePromises[key]
                try {
                    if (responseMessage.error) {
                        const error = responseMessage.error
                        responsePromise.reject(new ResponseError(error.code, error.message, error.data))
                    } else if (responseMessage.result !== undefined) {
                        responsePromise.resolve(responseMessage.result)
                    } else {
                        throw new Error('Should never happen.')
                    }
                } catch (error) {
                    if (error.message) {
                        logger.error(
                            `Response handler '${responsePromise.method}' failed with message: ${error.message}`
                        )
                    } else {
                        logger.error(`Response handler '${responsePromise.method}' failed unexpectedly.`)
                    }
                }
            } else {
                tracer.unknownResponseReceived(responseMessage)
            }
        }
    }

    function handleNotification(message: NotificationMessage): void {
        if (isUnsubscribed()) {
            // See handle request.
            return
        }
        let notificationHandler: GenericNotificationHandler | undefined
        if (message.method === CancelNotification.type) {
            notificationHandler = (params: CancelParams) => {
                const id = params.id
                const source = requestTokens[String(id)]
                if (source) {
                    source.cancel()
                }
            }
        } else {
            const element = notificationHandlers[message.method]
            if (element) {
                notificationHandler = element.handler
            }
        }
        if (notificationHandler || starNotificationHandler) {
            try {
                tracer.notificationReceived(message)
                notificationHandler
                    ? notificationHandler(message.params)
                    : starNotificationHandler!(message.method, message.params)
            } catch (error) {
                if (error.message) {
                    logger.error(`Notification handler '${message.method}' failed with message: ${error.message}`)
                } else {
                    logger.error(`Notification handler '${message.method}' failed unexpectedly.`)
                }
            }
        } else {
            unhandledNotificationEmitter.fire(message)
        }
    }

    function handleInvalidMessage(message: Message): void {
        if (!message) {
            logger.error('Received empty message.')
            return
        }
        logger.error(
            `Received message which is neither a response nor a notification message:\n${JSON.stringify(
                message,
                null,
                4
            )}`
        )
        // Test whether we find an id to reject the promise
        const responseMessage: ResponseMessage = message as ResponseMessage
        if (typeof responseMessage.id === 'string' || typeof responseMessage.id === 'number') {
            const key = String(responseMessage.id)
            const responseHandler = responsePromises[key]
            if (responseHandler) {
                responseHandler.reject(new Error('The received response has neither a result nor an error property.'))
            }
        }
    }

    function throwIfClosedOrUnsubscribed(): void {
        if (isClosed()) {
            throw new ConnectionError(ConnectionErrors.Closed, 'Connection is closed.')
        }
        if (isUnsubscribed()) {
            throw new ConnectionError(ConnectionErrors.Unsubscribed, 'Connection is unsubscribed.')
        }
    }

    function throwIfListening(): void {
        if (isListening()) {
            throw new ConnectionError(ConnectionErrors.AlreadyListening, 'Connection is already listening')
        }
    }

    function throwIfNotListening(): void {
        if (!isListening()) {
            throw new Error('Call listen() first.')
        }
    }

    const connection: Connection = {
        sendNotification: (method: string, params: any): void => {
            throwIfClosedOrUnsubscribed()
            const notificationMessage: NotificationMessage = {
                jsonrpc: version,
                method,
                params,
            }
            tracer.notificationSent(notificationMessage)
            transports.writer.write(notificationMessage)
        },
        onNotification: (type: string | StarNotificationHandler, handler?: GenericNotificationHandler): void => {
            throwIfClosedOrUnsubscribed()
            if (typeof type === 'function') {
                starNotificationHandler = type
            } else if (handler) {
                notificationHandlers[type] = { type: undefined, handler }
            }
        },
        sendRequest: <R>(method: string, params: any, token?: CancellationToken) => {
            throwIfClosedOrUnsubscribed()
            throwIfNotListening()
            token = CancellationToken.is(token) ? token : undefined
            const id = sequenceNumber++
            const result = new Promise<R>((resolve, reject) => {
                const requestMessage: RequestMessage = {
                    jsonrpc: version,
                    id,
                    method,
                    params,
                }
                let responsePromise: ResponsePromise | null = {
                    method,
                    request: trace === Trace.Verbose ? requestMessage : undefined,
                    timerStart: Date.now(),
                    resolve,
                    reject,
                }
                tracer.requestSent(requestMessage)
                try {
                    transports.writer.write(requestMessage)
                } catch (e) {
                    // Writing the message failed. So we need to reject the promise.
                    responsePromise.reject(
                        new ResponseError<void>(ErrorCodes.MessageWriteError, e.message ? e.message : 'Unknown reason')
                    )
                    responsePromise = null
                }
                if (responsePromise) {
                    responsePromises[String(id)] = responsePromise
                }
            })
            if (token) {
                token.onCancellationRequested(() => {
                    connection.sendNotification(CancelNotification.type, { id })
                })
            }
            return result
        },
        onRequest: <R, E>(type: string | StarRequestHandler, handler?: GenericRequestHandler<R, E>): void => {
            throwIfClosedOrUnsubscribed()

            if (typeof type === 'function') {
                starRequestHandler = type
            } else if (handler) {
                requestHandlers[type] = { type: undefined, handler }
            }
        },
        trace: (value: Trace, _tracer: Tracer, sendNotification = false) => {
            trace = value
            if (trace === Trace.Off) {
                tracer = noopTracer
            } else {
                tracer = _tracer
            }
        },
        onError: errorEmitter.event,
        onClose: closeEmitter.event,
        onUnhandledNotification: unhandledNotificationEmitter.event,
        onUnsubscribe: unsubscribeEmitter.event,
        unsubscribe: () => {
            if (isUnsubscribed()) {
                return
            }
            state = ConnectionState.Unsubscribed
            unsubscribeEmitter.fire(undefined)
            for (const key of Object.keys(responsePromises)) {
                responsePromises[key].reject(
                    new ConnectionError(
                        ConnectionErrors.Unsubscribed,
                        `The underlying JSON-RPC connection got unsubscribed while responding to this ${
                            responsePromises[key].method
                        } request.`
                    )
                )
            }
            responsePromises = Object.create(null)
            requestTokens = Object.create(null)
            messageQueue = new LinkedMap<string, Message>()
            transports.writer.unsubscribe()
            transports.reader.unsubscribe()
        },
        listen: () => {
            throwIfClosedOrUnsubscribed()
            throwIfListening()

            state = ConnectionState.Listening
            transports.reader.listen(callback)
        },
    }

    return connection
}

/** Support browser and node environments without needing a transpiler. */
function setImmediateCompat(f: () => void): NodeJS.Timer {
    if (typeof setImmediate !== 'undefined') {
        const immediate = setImmediate(f)
        return {
            ref: () => immediate.ref(),
            unref: () => immediate.ref(),
            refresh: () => {
                // noop
            },
        }
    }
    return setTimeout(f, 0)
}
