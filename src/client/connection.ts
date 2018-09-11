import { Unsubscribable } from 'rxjs'
import { ExitNotification, InitializeParams, InitializeRequest, InitializeResult, ShutdownRequest } from '../protocol'
import { createMessageConnection, Logger, MessageTransports } from '../protocol/jsonrpc2/connection'
import { ConnectionStrategy } from '../protocol/jsonrpc2/connectionStrategy'
import {
    GenericNotificationHandler,
    GenericRequestHandler,
    NotificationHandler,
    RequestHandler,
} from '../protocol/jsonrpc2/handlers'
import { Message, MessageType as RPCMessageType, NotificationType, RequestType } from '../protocol/jsonrpc2/messages'
import { Trace, Tracer } from '../protocol/jsonrpc2/trace'

export interface Connection extends Unsubscribable {
    listen(): void
    sendRequest<P, R, E, RO>(type: RequestType<P, R, E, RO>, params: P): Promise<R>
    sendRequest<R>(method: string, param?: any): Promise<R>
    sendRequest<R>(type: string | RPCMessageType, ...params: any[]): Promise<R>
    onRequest<P, R, E, RO>(type: RequestType<P, R, E, RO>, handler: RequestHandler<P, R, E>): void
    onRequest<R, E>(method: string | RPCMessageType, handler: GenericRequestHandler<R, E>): void
    sendNotification<P, RO>(type: NotificationType<P, RO>, params?: P): void
    sendNotification(method: string | RPCMessageType, params?: any): void
    onNotification<P, RO>(type: NotificationType<P, RO>, handler: NotificationHandler<P>): void
    onNotification(method: string | RPCMessageType, handler: GenericNotificationHandler): void
    initialize(params: InitializeParams): Promise<InitializeResult>
    shutdown(): Promise<void>
    exit(): void
    trace(value: Trace, tracer: Tracer, sendNotification?: boolean): void
}

type ConnectionErrorHandler = (error: Error, message: Message | undefined, count: number | undefined) => void

type ConnectionCloseHandler = () => void

export function createConnection(
    transports: MessageTransports,
    errorHandler?: ConnectionErrorHandler,
    closeHandler?: ConnectionCloseHandler,
    logger?: Logger,
    strategy?: ConnectionStrategy
): Connection {
    const connection = createMessageConnection(transports, logger, strategy)
    if (errorHandler) {
        connection.onError(data => errorHandler(data[0], data[1], data[2]))
    }
    if (closeHandler) {
        connection.onClose(closeHandler)
    }
    return {
        listen: (): void => connection.listen(),
        sendRequest: <R>(type: string | RPCMessageType, ...params: any[]): Promise<R> =>
            connection.sendRequest(typeof type === 'string' ? type : type.method, ...params),
        onRequest: <R, E>(type: string | RPCMessageType, handler: GenericRequestHandler<R, E>): void =>
            connection.onRequest(typeof type === 'string' ? type : type.method, handler),
        sendNotification: (type: string | RPCMessageType, params?: any): void =>
            connection.sendNotification(typeof type === 'string' ? type : type.method, params),
        onNotification: (type: string | RPCMessageType, handler: GenericNotificationHandler): void =>
            connection.onNotification(typeof type === 'string' ? type : type.method, handler),
        initialize: (params: InitializeParams) => connection.sendRequest(InitializeRequest.type, params),
        shutdown: () => connection.sendRequest(ShutdownRequest.type, undefined),
        exit: () => connection.sendNotification(ExitNotification.type),
        trace: (value: Trace, tracer: Tracer, sendNotification = false): void =>
            connection.trace(value, tracer, sendNotification),
        unsubscribe: () => connection.unsubscribe(),
    }
}
