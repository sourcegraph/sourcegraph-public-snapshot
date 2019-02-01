import { toPromise } from 'abortable-rx'
import { from, fromEvent, isObservable, Observable, Observer, Subject, Unsubscribable } from 'rxjs'
import { takeUntil } from 'rxjs/operators'
import { isPromise } from '../../util'
import { Emitter, Event } from './events'
import { LinkedMap } from './linkedMap'
import {
    ErrorCodes,
    isNotificationMessage,
    isRequestMessage,
    isResponseMessage,
    Message,
    NotificationMessage,
    RequestID,
    RequestMessage,
    ResponseError,
    ResponseMessage,
} from './messages'
import { noopTracer, Tracer } from './trace'
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

type HandlerResult<R, E> =
    | R
    | ResponseError<E>
    | Promise<R>
    | Promise<ResponseError<E>>
    | Promise<R | ResponseError<E>>
    | Observable<R>

type StarRequestHandler = (method: string, params?: any, signal?: AbortSignal) => HandlerResult<any, any>

type GenericRequestHandler<R, E> = (params?: any, signal?: AbortSignal) => HandlerResult<R, E>

type StarNotificationHandler = (method: string, params?: any) => void

type GenericNotificationHandler = (params: any) => void

export interface Connection extends Unsubscribable {
    sendRequest<R>(method: string, params?: any[], signal?: AbortSignal): Promise<R>
    observeRequest<R>(method: string, params?: any[]): Observable<R>

    onRequest<R, E>(method: string, handler: GenericRequestHandler<R, E>): void
    onRequest(handler: StarRequestHandler): void

    sendNotification(method: string, params?: any[]): void

    onNotification(method: string, handler: GenericNotificationHandler): void
    onNotification(handler: StarNotificationHandler): void

    trace(tracer: Tracer | null): void

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

export function createConnection(transports: MessageTransports, logger?: Logger): Connection {
    if (!logger) {
        logger = NullLogger
    }
    return _createConnection(transports, logger)
}

interface ResponseObserver {
    /** The request's method. */
    method: string

    /** Only set in Trace.Verbose mode. */
    request?: RequestMessage

    /** The timestamp when the request was received. */
    timerStart: number

    /** Whether the request was aborted by the client. */
    complete: boolean

    /** The observable containing the result value(s) and state. */
    observer: Observer<any>
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

const ABORT_REQUEST_METHOD = '$/abortRequest'

function _createConnection(transports: MessageTransports, logger: Logger): Connection {
    let sequenceNumber = 0
    let notificationSquenceNumber = 0
    let unknownResponseSquenceNumber = 0
    const version = '2.0'

    let starRequestHandler: StarRequestHandler | undefined
    const requestHandlers: { [name: string]: RequestHandlerElement | undefined } = Object.create(null)
    let starNotificationHandler: StarNotificationHandler | undefined
    const notificationHandlers: { [name: string]: NotificationHandlerElement | undefined } = Object.create(null)

    let timer = false
    let messageQueue: MessageQueue = new LinkedMap<string, Message>()
    let responseObservables: { [name: string]: ResponseObserver } = Object.create(null)
    let requestAbortControllers: { [id: string]: AbortController } = Object.create(null)

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
            const key = createResponseQueueKey(message.id) + Math.random() // TODO(sqs)
            queue.set(key, message)
        } else {
            queue.set(createNotificationQueueKey(), message)
        }
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
        timer = true
        setImmediateCompat(() => {
            timer = false
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
            // We have received an abort signal. Check if the message is still in the queue and abort it if allowed
            // to do so.
            if (isNotificationMessage(message) && message.method === ABORT_REQUEST_METHOD) {
                const key = createRequestQueueKey(message.params[0])
                const toAbort = messageQueue.get(key)
                if (isRequestMessage(toAbort)) {
                    messageQueue.delete(key)
                    const response: ResponseMessage = {
                        jsonrpc: '2.0',
                        id: toAbort.id,
                        error: { code: ErrorCodes.RequestAborted, message: 'request aborted' },
                    }
                    tracer.responseAborted(response, toAbort, message)
                    transports.writer.write(response)
                    return
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

        function reply(resultOrError: any | ResponseError<any>, complete: boolean): void {
            const message: ResponseMessage = {
                jsonrpc: version,
                id: requestMessage.id,
                complete,
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
                complete: true,
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
                complete: true,
            }
            tracer.responseSent(message, requestMessage, startTime)
            transports.writer.write(message)
        }
        function replyComplete(): void {
            const message: ResponseMessage = {
                jsonrpc: version,
                id: requestMessage.id,
                complete: true,
            }
            tracer.responseSent(message, requestMessage, startTime)
            transports.writer.write(message)
        }

        tracer.requestReceived(requestMessage)

        const element = requestHandlers[requestMessage.method]
        const requestHandler: GenericRequestHandler<any, any> | undefined = element && element.handler
        if (requestHandler || starRequestHandler) {
            const abortController = new AbortController()
            const signalKey = String(requestMessage.id)
            requestAbortControllers[signalKey] = abortController
            try {
                const params = requestMessage.params !== undefined ? requestMessage.params : null
                const handlerResult = requestHandler
                    ? requestHandler(params, abortController.signal)
                    : starRequestHandler!(requestMessage.method, params, abortController.signal)

                if (!handlerResult) {
                    delete requestAbortControllers[signalKey]
                    replySuccess(handlerResult)
                } else if (isPromise(handlerResult) || isObservable(handlerResult)) {
                    const onComplete = () => {
                        delete requestAbortControllers[signalKey]
                    }
                    from(handlerResult)
                        .pipe(takeUntil(fromEvent(abortController.signal, 'abort')))
                        .subscribe(
                            value => reply(value, false),
                            error => {
                                onComplete()
                                if (error instanceof ResponseError) {
                                    replyError(error as ResponseError<any>)
                                } else if (error && typeof error.message === 'string') {
                                    replyError(
                                        new ResponseError<any>(ErrorCodes.InternalError, error.message, {
                                            stack: error.stack,
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
                            },
                            () => {
                                onComplete()
                                replyComplete()
                            }
                        )
                } else {
                    delete requestAbortControllers[signalKey]
                    reply(handlerResult, true)
                }
            } catch (error) {
                delete requestAbortControllers[signalKey]
                if (error instanceof ResponseError) {
                    reply(error as ResponseError<any>, true)
                } else if (error && typeof error.message === 'string') {
                    replyError(
                        new ResponseError<any>(ErrorCodes.InternalError, error.message, {
                            stack: error.stack,
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
            const responseObservable = responseObservables[key]
            if (responseObservable) {
                tracer.responseReceived(
                    responseMessage,
                    responseObservable.request || responseObservable.method,
                    responseObservable.timerStart
                )
                try {
                    if (responseMessage.error) {
                        const { code, message, data } = responseMessage.error
                        const err = new ResponseError(code, message, data)
                        if (data && data.stack) {
                            err.stack = data.stack
                        }
                        responseObservable.observer.error(err)
                    } else if (responseMessage.result !== undefined) {
                        responseObservable.observer.next(responseMessage.result)
                    }
                    if (responseMessage.complete) {
                        responseObservable.complete = true
                        responseObservable.observer.complete()
                    }
                } catch (error) {
                    if (error.message) {
                        logger.error(
                            `Response handler '${responseObservable.method}' failed with message: ${error.message}`
                        )
                    } else {
                        logger.error(`Response handler '${responseObservable.method}' failed unexpectedly.`)
                    }
                } finally {
                    if (responseMessage.complete) {
                        delete responseObservables[key]
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
        if (message.method === ABORT_REQUEST_METHOD) {
            notificationHandler = (params: [RequestID]) => {
                const id = params[0]
                const abortController = requestAbortControllers[String(id)]
                if (abortController) {
                    abortController.abort()
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
            const responseHandler = responseObservables[key]
            if (responseHandler) {
                responseHandler.observer.error(
                    new Error('The received response has neither a result nor an error property.')
                )
            }
        }
    }

    function throwIfClosedOrUnsubscribed(): void {
        if (isClosed()) {
            throw new ConnectionError(
                ConnectionErrors.Closed,
                'Extension host connection unexpectedly closed. Reload the page to resolve.'
            )
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

    const sendNotification = (method: string, params?: any[]): void => {
        throwIfClosedOrUnsubscribed()
        const notificationMessage: NotificationMessage = {
            jsonrpc: version,
            method,
            params,
        }
        tracer.notificationSent(notificationMessage)
        transports.writer.write(notificationMessage)
    }

    /**
     * @returns a hot observable with the result of sending the request
     */
    const requestHelper = <R>(method: string, params?: any[]): Observable<R> => {
        const id = sequenceNumber++
        const requestMessage: RequestMessage = {
            jsonrpc: version,
            id,
            method,
            params,
        }
        const subject = new Subject<R>()
        const responseObserver: ResponseObserver = {
            method,
            request: requestMessage,
            timerStart: Date.now(),
            complete: false,
            observer: subject,
        }
        tracer.requestSent(requestMessage)
        try {
            transports.writer.write(requestMessage)
            responseObservables[String(id)] = responseObserver
        } catch (e) {
            responseObserver.observer.error(
                new ResponseError<void>(ErrorCodes.MessageWriteError, e.message ? e.message : 'Unknown reason')
            )
        }
        return new Observable(observer => {
            subject.subscribe(observer).add(() => {
                if (
                    !isUnsubscribed() &&
                    responseObserver &&
                    !responseObserver.complete &&
                    !responseObserver.observer.closed
                ) {
                    sendNotification(ABORT_REQUEST_METHOD, [id])
                }
            })
        })
    }

    const connection: Connection = {
        sendNotification,
        onNotification: (type: string | StarNotificationHandler, handler?: GenericNotificationHandler): void => {
            throwIfClosedOrUnsubscribed()
            if (typeof type === 'function') {
                starNotificationHandler = type
            } else if (handler) {
                notificationHandlers[type] = { type: undefined, handler }
            }
        },
        sendRequest: <R>(method: string, params?: any[], signal?: AbortSignal): Promise<R> => {
            throwIfClosedOrUnsubscribed()
            throwIfNotListening()
            return toPromise(requestHelper<R>(method, params), signal)
        },
        observeRequest: <R>(method: string, params?: any[]): Observable<R> => {
            throwIfClosedOrUnsubscribed()
            throwIfNotListening()
            return requestHelper<R>(method, params)
        },
        onRequest: <R, E>(type: string | StarRequestHandler, handler?: GenericRequestHandler<R, E>): void => {
            throwIfClosedOrUnsubscribed()

            if (typeof type === 'function') {
                starRequestHandler = type
            } else if (handler) {
                requestHandlers[type] = { type: undefined, handler }
            }
        },
        trace: (_tracer: Tracer | null) => {
            if (_tracer) {
                tracer = _tracer
            } else {
                tracer = noopTracer
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
            for (const key of Object.keys(responseObservables)) {
                responseObservables[key].observer.error(
                    new ConnectionError(
                        ConnectionErrors.Unsubscribed,
                        `The underlying JSON-RPC connection got unsubscribed while responding to this ${
                            responseObservables[key].method
                        } request.`
                    )
                )
            }
            responseObservables = Object.create(null)
            requestAbortControllers = Object.create(null)
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
function setImmediateCompat(f: () => void): void {
    if (typeof setImmediate !== 'undefined') {
        setImmediate(f)
        return
    }
    setTimeout(f, 0)
}
