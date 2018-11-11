import { Event } from './events';
import { Message } from './messages';
export declare type DataCallback = (data: Message) => void;
export interface MessageReader {
    readonly onError: Event<Error>;
    readonly onClose: Event<void>;
    listen(callback: DataCallback): void;
    unsubscribe(): void;
}
export declare abstract class AbstractMessageReader {
    private errorEmitter;
    private closeEmitter;
    constructor();
    unsubscribe(): void;
    readonly onError: Event<Error>;
    protected fireError(error: any): void;
    readonly onClose: Event<void>;
    protected fireClose(): void;
    private asError;
}
export interface MessageWriter {
    readonly onError: Event<[Error, Message | undefined, number | undefined]>;
    readonly onClose: Event<void>;
    write(msg: Message): void;
    unsubscribe(): void;
}
export declare abstract class AbstractMessageWriter {
    private errorEmitter;
    private closeEmitter;
    unsubscribe(): void;
    readonly onError: Event<[Error, Message | undefined, number | undefined]>;
    protected fireError(error: any, message?: Message, count?: number): void;
    readonly onClose: Event<void>;
    protected fireClose(): void;
    private asError;
}
