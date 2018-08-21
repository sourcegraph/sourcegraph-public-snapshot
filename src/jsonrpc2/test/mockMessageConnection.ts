import { isFunction } from '../../util'
import { CancellationToken } from '../cancel'
import { MessageConnection } from '../connection'
import { Emitter } from '../events'
import {
    GenericNotificationHandler,
    GenericRequestHandler,
    StarNotificationHandler,
    StarRequestHandler,
} from '../handlers'
import { Message, MessageType, NotificationMessage, NotificationType, RequestType } from '../messages'
import { Trace, Tracer } from '../trace'

/**
 * A mock implementation of {@link MessageConnection} for use in tests.
 */
export class MockMessageConnection implements MessageConnection {
    /**
     * Messages sent by calls to {@link MockMessageConnection#sendRequest} or
     * {@link MockMessageConnection#sendNotification}.
     */
    public sentMessages: { method: string; params?: any }[] = []

    /**
     * Message handlers registered by calls to {@link MockMessageConnection#onRequest} or
     * {@link MockMessageConnection#onNotification}.
     */
    public registeredHandlers = new Map<string, GenericRequestHandler<any, any> | GenericNotificationHandler>()

    /**
     * Mock responses that are returned as results from {@link MockMessageConnection#sendRequest}.
     */
    public mockResults = new Map<string, any>()

    /**
     * Simulates receiving the given request by calling the handler registered for it.
     *
     * @return The result returned by the request handler.
     */
    public recvRequest<P, R, E, RO>(type: RequestType<P, R, E, RO>, params: P): Promise<R>
    public recvRequest<R>(method: string, params: any): Promise<R>
    public recvRequest<R>(type: MessageType | string, params: any): Promise<R> {
        const method = typeof type === 'string' ? type : type.method
        const handler = this.registeredHandlers.get(method)
        if (!handler) {
            throw new Error(`no mock request handler for method ${method}`)
        }
        return Promise.resolve(handler(params))
    }

    /**
     * Simulates receiving the given notification by calling the handler registered for it.
     */
    public recvNotification<P, RO>(type: NotificationType<P, RO>, params: P): void
    public recvNotification<R>(method: string, params: any): void
    public recvNotification<R>(type: MessageType | string, params: any): void {
        const method = typeof type === 'string' ? type : type.method
        const handler = this.registeredHandlers.get(method)
        if (!handler) {
            throw new Error(`no mock notification handler for method ${method}`)
        }
        handler(params)
    }

    public sendRequest(type: string | MessageType, params: any, token?: CancellationToken): Promise<any> {
        const method = typeof type === 'string' ? type : type.method
        this.sentMessages.push({ method, params })
        if (!this.mockResults.has(method)) {
            throw new Error(`no mock result for method ${method}`)
        }
        const resultOrError = this.mockResults.get(method)
        return resultOrError instanceof Error ? Promise.reject(resultOrError) : Promise.resolve(resultOrError)
    }

    public sendNotification(type: string | MessageType, params: any): void {
        this.sentMessages.push({ method: typeof type === 'string' ? type : type.method, params })
    }

    public onRequest(type: string | MessageType, handler: GenericRequestHandler<any, any>): void
    public onRequest(handler: StarRequestHandler): void
    public onRequest(arg1: string | MessageType | StarRequestHandler, arg2?: GenericRequestHandler<any, any>): void {
        return this.onHandler(arg1, arg2)
    }

    public onNotification(method: string | MessageType, handler: GenericNotificationHandler): void
    public onNotification(handler: StarNotificationHandler): void
    public onNotification(
        arg1: string | MessageType | StarNotificationHandler,
        arg2?: GenericNotificationHandler
    ): void {
        return this.onHandler(arg1, arg2)
    }

    private onHandler(
        arg1: string | MessageType | StarRequestHandler | StarNotificationHandler,
        arg2?: GenericRequestHandler<any, any> | GenericNotificationHandler
    ): void {
        let method: string
        let handler: GenericRequestHandler<any, any> | GenericNotificationHandler
        if (isFunction(arg1)) {
            method = '*'
            handler = arg1
        } else {
            method = typeof arg1 === 'string' ? arg1 : arg1.method
            handler = arg2!
        }
        this.registeredHandlers.set(method, handler)
    }

    public trace(value: Trace, tracer: Tracer, sendNotification?: boolean): void {
        /* noop */
    }

    public errorEmitter = new Emitter<[Error, Message | undefined, number | undefined]>()
    public onError = this.errorEmitter.event

    public closeEmitter = new Emitter<void>()
    public onClose = this.closeEmitter.event

    public unhandledNotificationEmitter = new Emitter<NotificationMessage>()
    public onUnhandledNotification = this.unhandledNotificationEmitter.event

    public listen(): void {
        /* noop */
    }

    public unsubscribeEmitter = new Emitter<void>()
    public onUnsubscribe = this.unsubscribeEmitter.event

    public unsubscribe(): void {
        /* noop */
    }
}
