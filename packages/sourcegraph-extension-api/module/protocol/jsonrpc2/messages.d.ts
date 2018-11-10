/**
 * A language server message
 */
export interface Message {
    jsonrpc: string;
}
/**
 * Request message
 */
export interface RequestMessage extends Message {
    /**
     * The request id.
     */
    id: number | string;
    /**
     * The method to be invoked.
     */
    method: string;
    /**
     * The method's params.
     */
    params?: any;
}
/**
 * Predefined error codes.
 */
export declare namespace ErrorCodes {
    const ParseError = -32700;
    const InvalidRequest = -32600;
    const MethodNotFound = -32601;
    const InvalidParams = -32602;
    const InternalError = -32603;
    const serverErrorStart = -32099;
    const serverErrorEnd = -32000;
    const ServerNotInitialized = -32002;
    const UnknownErrorCode = -32001;
    const RequestCancelled = -32800;
    const MessageWriteError = 1;
    const MessageReadError = 2;
}
interface ResponseErrorLiteral<D> {
    /**
     * A number indicating the error type that occured.
     */
    code: number;
    /**
     * A string providing a short decription of the error.
     */
    message: string;
    /**
     * A Primitive or Structured value that contains additional
     * information about the error. Can be omitted.
     */
    data?: D;
}
/**
 * An error object return in a response in case a request
 * has failed.
 */
export declare class ResponseError<D> extends Error {
    readonly code: number;
    readonly data: D | undefined;
    constructor(code: number, message: string, data?: D);
    toJSON(): ResponseErrorLiteral<D>;
}
/**
 * A response message.
 */
export interface ResponseMessage extends Message {
    /**
     * The request id.
     */
    id: number | string | null;
    /**
     * The result of a request. This can be omitted in
     * the case of an error.
     */
    result?: any;
    /**
     * The error object in case a request fails.
     */
    error?: ResponseErrorLiteral<any>;
}
/**
 * Notification Message
 */
export interface NotificationMessage extends Message {
    /**
     * The method to be invoked.
     */
    method: string;
    /**
     * The notification's params.
     */
    params?: any;
}
/**
 * Tests if the given message is a request message
 */
export declare function isRequestMessage(message: Message | undefined): message is RequestMessage;
/**
 * Tests if the given message is a notification message
 */
export declare function isNotificationMessage(message: Message | undefined): message is NotificationMessage;
/**
 * Tests if the given message is a response message
 */
export declare function isResponseMessage(message: Message | undefined): message is ResponseMessage;
export {};
