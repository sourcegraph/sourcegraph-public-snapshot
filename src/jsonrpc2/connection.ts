import { Unsubscribable } from 'rxjs'
import { isFunction } from '../util'
import { CancelNotification, CancelParams } from './cancel'
import { CancellationToken, CancellationTokenSource } from './cancel'
import { ConnectionStrategy } from './connectionStrategy'
import { Emitter, Event } from './events'
import {
    GenericNotificationHandler,
    GenericRequestHandler,
    NotificationHandler,
    RequestHandler,
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
    MessageType,
    NotificationMessage,
    NotificationType,
    RequestMessage,
    RequestType,
    ResponseError,
    ResponseMessage,
} from './messages'
import { LogTraceNotification, SetTraceNotification, Trace, Tracer } from './trace'
import { DataCallback, MessageReader, MessageWriter } from './transport'

// Copied from vscode-languageserver to avoid adding extraneous dependencies.

export interface Logger {
    error(message: string): void
    warn(message: string): void
    info(message: string): void
    log(message: string): void
}

export const NullLogger: Logger = Object.freeze({
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

export type MessageQueue = LinkedMap<string, Message>

export interface MessageConnection extends Unsubscribable {
    sendRequest<P, R, E, RO>(type: RequestType<P, R, E, RO>, params: P, token?: CancellationToken): Promise<R>
    sendRequest<R>(method: string, ...params: any[]): Promise<R>

    onRequest<P, R, E, RO>(type: RequestType<P, R, E, RO>, handler: RequestHandler<P, R, E>): void
    onRequest<R, E>(method: string, handler: GenericRequestHandler<R, E>): void
    onRequest(handler: StarRequestHandler): void

    sendNotification<P, RO>(type: NotificationType<P, RO>, params?: P): void
    sendNotification(method: string, ...params: any[]): void

    onNotification<P, RO>(type: NotificationType<P, RO>, handler: NotificationHandler<P>): void
    onNotification(method: string, handler: GenericNotificationHandler): void
    onNotification(handler: StarNotificationHandler): void

    trace(value: Trace, tracer: Tracer, sendNotification?: boolean): void

    onError: Event<[Error, Message | undefined, number | undefined]>
    onClose: Event<void>
    onUnhandledNotification: Event<NotificationMessage>
    listen(): void
    onUnsubscribe: Event<void>
    inspect(): void
}

export interface MessageTransports {
    reader: MessageReader
    writer: MessageWriter
}

export function createMessageConnection(
    transports: MessageTransports,
    logger?: Logger,
    strategy?: ConnectionStrategy
): MessageConnection {
    if (!logger) {
        logger = NullLogger
    }
    return _createMessageConnection(transports, logger, strategy)
}

interface ResponsePromise {
    method: string
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
    type: MessageType | undefined
    handler: GenericRequestHandler<any, any>
}

interface NotificationHandlerElement {
    type: MessageType | undefined
    handler: GenericNotificationHandler
}

function _createMessageConnection(
    transports: MessageTransports,
    logger: Logger,
    strategy?: ConnectionStrategy
): MessageConnection {
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
    let tracer: Tracer | undefined

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
            // We have received a cancellation message. Check if the message is still in the queue
            // and cancel it if allowed to do so.
            if (isNotificationMessage(message) && message.method === CancelNotification.type.method) {
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
                        traceSendingResponse(response, message.method, Date.now())
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

        function reply(resultOrError: any | ResponseError<any>, method: string, startTime: number): void {
            const message: ResponseMessage = {
                jsonrpc: version,
                id: requestMessage.id,
            }
            if (resultOrError instanceof ResponseError) {
                message.error = (resultOrError as ResponseError<any>).toJson()
            } else {
                message.result = resultOrError === undefined ? null : resultOrError
            }
            traceSendingResponse(message, method, startTime)
            transports.writer.write(message)
        }
        function replyError(error: ResponseError<any>, method: string, startTime: number): void {
            const message: ResponseMessage = {
                jsonrpc: version,
                id: requestMessage.id,
                error: error.toJson(),
            }
            traceSendingResponse(message, method, startTime)
            transports.writer.write(message)
        }
        function replySuccess(result: any, method: string, startTime: number): void {
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
            traceSendingResponse(message, method, startTime)
            transports.writer.write(message)
        }

        traceReceivedRequest(requestMessage)

        const element = requestHandlers[requestMessage.method]
        const requestHandler: GenericRequestHandler<any, any> | undefined = element && element.handler
        const startTime = Date.now()
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
                    replySuccess(handlerResult, requestMessage.method, startTime)
                } else if (promise.then) {
                    promise.then(
                        (resultOrError): any | ResponseError<any> => {
                            delete requestTokens[tokenKey]
                            reply(resultOrError, requestMessage.method, startTime)
                        },
                        error => {
                            delete requestTokens[tokenKey]
                            if (error instanceof ResponseError) {
                                replyError(error as ResponseError<any>, requestMessage.method, startTime)
                            } else if (error && typeof error.message === 'string') {
                                replyError(
                                    new ResponseError<void>(
                                        ErrorCodes.InternalError,
                                        `Request ${requestMessage.method} failed with message: ${error.message}`
                                    ),
                                    requestMessage.method,
                                    startTime
                                )
                            } else {
                                replyError(
                                    new ResponseError<void>(
                                        ErrorCodes.InternalError,
                                        `Request ${
                                            requestMessage.method
                                        } failed unexpectedly without providing any details.`
                                    ),
                                    requestMessage.method,
                                    startTime
                                )
                            }
                        }
                    )
                } else {
                    delete requestTokens[tokenKey]
                    reply(handlerResult, requestMessage.method, startTime)
                }
            } catch (error) {
                delete requestTokens[tokenKey]
                if (error instanceof ResponseError) {
                    reply(error as ResponseError<any>, requestMessage.method, startTime)
                } else if (error && typeof error.message === 'string') {
                    replyError(
                        new ResponseError<void>(
                            ErrorCodes.InternalError,
                            `Request ${requestMessage.method} failed with message: ${error.message}`
                        ),
                        requestMessage.method,
                        startTime
                    )
                } else {
                    replyError(
                        new ResponseError<void>(
                            ErrorCodes.InternalError,
                            `Request ${requestMessage.method} failed unexpectedly without providing any details.`
                        ),
                        requestMessage.method,
                        startTime
                    )
                }
            }
        } else {
            replyError(
                new ResponseError<void>(ErrorCodes.MethodNotFound, `Unhandled method ${requestMessage.method}`),
                requestMessage.method,
                startTime
            )
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
            traceReceivedResponse(responseMessage, responsePromise)
            if (responsePromise) {
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
            }
        }
    }

    function handleNotification(message: NotificationMessage): void {
        if (isUnsubscribed()) {
            // See handle request.
            return
        }
        let notificationHandler: GenericNotificationHandler | undefined
        if (message.method === CancelNotification.type.method) {
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
                traceReceivedNotification(message)
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

    function traceSendingRequest(message: RequestMessage): void {
        if (trace === Trace.Off || !tracer) {
            return
        }
        let data: string | undefined
        if (trace === Trace.Verbose && message.params) {
            data = `Params: ${JSON.stringify(message.params, null, 4)}\n\n`
        }
        tracer.log(`Sending request '${message.method} - (${message.id})'.`, data)
    }

    function traceSendNotification(message: NotificationMessage): void {
        if (trace === Trace.Off || !tracer) {
            return
        }
        let data: string | undefined
        if (trace === Trace.Verbose) {
            if (message.params) {
                data = `Params: ${JSON.stringify(message.params, null, 4)}\n\n`
            } else {
                data = 'No parameters provided.\n\n'
            }
        }
        tracer.log(`Sending notification '${message.method}'.`, data)
    }

    function traceSendingResponse(message: ResponseMessage, method: string, startTime: number): void {
        if (trace === Trace.Off || !tracer) {
            return
        }
        let data: string | undefined
        if (trace === Trace.Verbose) {
            if (message.error && message.error.data) {
                data = `Error data: ${JSON.stringify(message.error.data, null, 4)}\n\n`
            } else {
                if (message.result) {
                    data = `Result: ${JSON.stringify(message.result, null, 4)}\n\n`
                } else if (message.error === undefined) {
                    data = 'No result returned.\n\n'
                }
            }
        }
        tracer.log(
            `Sending response '${method} - (${message.id})'. Processing request took ${Date.now() - startTime}ms`,
            data
        )
    }

    function traceReceivedRequest(message: RequestMessage): void {
        if (trace === Trace.Off || !tracer) {
            return
        }
        let data: string | undefined
        if (trace === Trace.Verbose && message.params) {
            data = `Params: ${JSON.stringify(message.params, null, 4)}\n\n`
        }
        tracer.log(`Received request '${message.method} - (${message.id})'.`, data)
    }

    function traceReceivedNotification(message: NotificationMessage): void {
        if (trace === Trace.Off || !tracer || message.method === LogTraceNotification.type.method) {
            return
        }
        let data: string | undefined
        if (trace === Trace.Verbose) {
            if (message.params) {
                data = `Params: ${JSON.stringify(message.params, null, 4)}\n\n`
            } else {
                data = 'No parameters provided.\n\n'
            }
        }
        tracer.log(`Received notification '${message.method}'.`, data)
    }

    function traceReceivedResponse(message: ResponseMessage, responsePromise: ResponsePromise): void {
        if (trace === Trace.Off || !tracer) {
            return
        }
        let data: any
        if (trace === Trace.Verbose) {
            if (message.error && message.error.data) {
                data = message.error.data
            } else {
                if (message.result) {
                    data = message.result
                } else if (message.error === undefined) {
                    data = 'No result returned.\n\n'
                }
            }
        }
        if (responsePromise) {
            const error = message.error ? ` Request failed: ${message.error.message} (${message.error.code}).` : ''
            tracer.log(
                `RESP ${responsePromise.method} - (${message.id}) in ${Date.now() -
                    responsePromise.timerStart}ms.${error}`,
                data
            )
        } else {
            tracer.log(`Received response ${message.id} without active response promise.`, data)
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

    const connection: MessageConnection = {
        sendNotification: (type: string | MessageType, params: any): void => {
            throwIfClosedOrUnsubscribed()

            let method: string
            if (typeof type === 'string') {
                method = type
            } else {
                method = type.method
            }
            const notificationMessage: NotificationMessage = {
                jsonrpc: version,
                method,
                params,
            }
            traceSendNotification(notificationMessage)
            transports.writer.write(notificationMessage)
        },
        onNotification: (
            type: string | MessageType | StarNotificationHandler,
            handler?: GenericNotificationHandler
        ): void => {
            throwIfClosedOrUnsubscribed()
            if ((isFunction as (v: any) => v is StarNotificationHandler)(type)) {
                starNotificationHandler = type as StarNotificationHandler
            } else if (handler) {
                if (typeof type === 'string') {
                    notificationHandlers[type] = { type: undefined, handler }
                } else {
                    notificationHandlers[type.method] = { type, handler }
                }
            }
        },
        sendRequest: <R, E>(type: string | MessageType, params: any, token?: CancellationToken) => {
            throwIfClosedOrUnsubscribed()
            throwIfNotListening()

            let method: string
            if (typeof type === 'string') {
                method = type
            } else {
                method = type.method
                token = CancellationToken.is(token) ? token : undefined
            }

            const id = sequenceNumber++
            const result = new Promise<R | ResponseError<E>>((resolve, reject) => {
                const requestMessage: RequestMessage = {
                    jsonrpc: version,
                    id,
                    method,
                    params,
                }
                let responsePromise: ResponsePromise | null = {
                    method,
                    timerStart: Date.now(),
                    resolve,
                    reject,
                }
                traceSendingRequest(requestMessage)
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
        onRequest: <R, E>(
            type: string | MessageType | StarRequestHandler,
            handler?: GenericRequestHandler<R, E>
        ): void => {
            throwIfClosedOrUnsubscribed()

            if ((isFunction as (v: any) => v is StarRequestHandler)(type)) {
                starRequestHandler = type as StarRequestHandler
            } else if (handler) {
                if (typeof type === 'string') {
                    requestHandlers[type] = { type: undefined, handler }
                } else {
                    requestHandlers[type.method] = { type, handler }
                }
            }
        },
        trace: (_value: Trace, _tracer: Tracer, sendNotification = false) => {
            trace = _value
            if (trace === Trace.Off) {
                tracer = undefined
            } else {
                tracer = _tracer
            }
            if (sendNotification && !isClosed() && !isUnsubscribed()) {
                connection.sendNotification(SetTraceNotification.type, { value: Trace.toString(_value) })
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
            const error = new Error('Connection got unsubscribed.')
            for (const key of Object.keys(responsePromises)) {
                responsePromises[key].reject(error)
            }
            responsePromises = Object.create(null)
            requestTokens = Object.create(null)
            messageQueue = new LinkedMap<string, Message>()
            // Test for backwards compatibility
            // tslint:disable-next-line:no-unbound-method
            if (isFunction(transports.writer.unsubscribe)) {
                transports.writer.unsubscribe()
            }
            // tslint:disable-next-line:no-unbound-method
            if (isFunction(transports.reader.unsubscribe)) {
                transports.reader.unsubscribe()
            }
        },
        listen: () => {
            throwIfClosedOrUnsubscribed()
            throwIfListening()

            state = ConnectionState.Listening
            transports.reader.listen(callback)
        },
        inspect: (): void => {
            console.log('inspect')
        },
    }

    connection.onNotification(LogTraceNotification.type, params => {
        if (trace === Trace.Off || !tracer) {
            return
        }
        tracer.log(params.message, trace === Trace.Verbose ? params.verbose : undefined)
    })

    return connection
}

/** Support browser and node environments without needing a transpiler. */
function setImmediateCompat(f: () => void): NodeJS.Timer {
    if (typeof setImmediate !== 'undefined') {
        return setImmediate(f)
    }
    return setTimeout(f, 0)
}
