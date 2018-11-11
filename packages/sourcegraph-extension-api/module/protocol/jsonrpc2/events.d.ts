import { Unsubscribable } from 'rxjs';
/**
 * Represents a typed event.
 */
export declare type Event<T> = (listener: (e: T) => any, thisArgs?: any) => Unsubscribable;
export declare namespace Event {
    const None: Event<any>;
}
export declare class Emitter<T> {
    private static _noop;
    private _event?;
    private _callbacks;
    /**
     * For the public to allow to subscribe
     * to events from this Emitter
     */
    readonly event: Event<T>;
    /**
     * To be kept private to fire an event to
     * subscribers
     */
    fire(event: T): any;
    unsubscribe(): void;
}
