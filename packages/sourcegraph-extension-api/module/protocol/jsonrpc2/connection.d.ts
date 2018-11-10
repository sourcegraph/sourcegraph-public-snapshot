import { Unsubscribable } from 'rxjs';
import { ConnectionStrategy } from './connectionStrategy';
import { Event } from './events';
import { GenericNotificationHandler, GenericRequestHandler, StarNotificationHandler, StarRequestHandler } from './handlers';
import { Message, NotificationMessage } from './messages';
import { Trace, Tracer } from './trace';
import { MessageReader, MessageWriter } from './transport';
export interface Logger {
    error(message: string): void;
    warn(message: string): void;
    info(message: string): void;
    log(message: string): void;
}
export declare enum ConnectionErrors {
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
    AlreadyListening = 3
}
export declare class ConnectionError extends Error {
    readonly code: ConnectionErrors;
    constructor(code: ConnectionErrors, message: string);
}
export interface Connection extends Unsubscribable {
    sendRequest<R>(method: string, params?: any): Promise<R>;
    sendRequest<R>(method: string, ...params: any[]): Promise<R>;
    onRequest<R, E>(method: string, handler: GenericRequestHandler<R, E>): void;
    onRequest(handler: StarRequestHandler): void;
    sendNotification(method: string, ...params: any[]): void;
    onNotification(method: string, handler: GenericNotificationHandler): void;
    onNotification(handler: StarNotificationHandler): void;
    trace(value: Trace, tracer: Tracer): void;
    onError: Event<[Error, Message | undefined, number | undefined]>;
    onClose: Event<void>;
    onUnhandledNotification: Event<NotificationMessage>;
    listen(): void;
    onUnsubscribe: Event<void>;
}
export interface MessageTransports {
    reader: MessageReader;
    writer: MessageWriter;
}
export declare function createConnection(transports: MessageTransports, logger?: Logger, strategy?: ConnectionStrategy): Connection;
