import { NotificationMessage, RequestMessage, ResponseMessage } from './messages';
export declare enum Trace {
    Off = "off",
    Messages = "messages",
    Verbose = "verbose"
}
/** Records messages sent and received on a JSON-RPC 2.0 connection. */
export interface Tracer {
    log(message: string, details?: string): void;
    requestSent(message: RequestMessage): void;
    requestReceived(message: RequestMessage): void;
    notificationSent(message: NotificationMessage): void;
    notificationReceived(message: NotificationMessage): void;
    responseSent(message: ResponseMessage, request: RequestMessage, startTime: number): void;
    responseCanceled(message: ResponseMessage, request: RequestMessage, cancelMessage: NotificationMessage): void;
    responseReceived(message: ResponseMessage, request: RequestMessage | string, startTime: number): void;
    unknownResponseReceived(message: ResponseMessage): void;
}
/** A tracer that implements the Tracer interface with noop methods. */
export declare const noopTracer: Tracer;
/** A tracer that implements the Tracer interface with console API calls, intended for a web browser. */
export declare class BrowserConsoleTracer implements Tracer {
    private name;
    constructor(name: string);
    private prefix;
    log(message: string, details?: string): void;
    requestSent(message: RequestMessage): void;
    requestReceived(message: RequestMessage): void;
    notificationSent(message: NotificationMessage): void;
    notificationReceived(message: NotificationMessage): void;
    responseSent(message: ResponseMessage, request: RequestMessage, startTime: number): void;
    responseCanceled(_message: ResponseMessage, request: RequestMessage, _cancelMessage: NotificationMessage): void;
    responseReceived(message: ResponseMessage, request: RequestMessage | string, startTime: number): void;
    unknownResponseReceived(message: ResponseMessage): void;
}
