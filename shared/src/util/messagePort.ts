/**
 * Compatible with {@link MessageEvent} but synthesizable.
 */
export interface MessageEventLike extends Pick<MessageEvent, 'data'> {}

/**
 * Compatible with {@link MessagePort} but synthesizable.
 */
export interface MessagePortLike extends Pick<MessagePort, 'postMessage' | 'start'> {
    addEventListener<E extends MessageEventLike>(
        type: 'message',
        listener: (ev: E) => any,
        options?: boolean | AddEventListenerOptions
    ): void
    removeEventListener<E extends MessageEventLike>(
        type: 'message',
        listener: (ev: E) => any,
        options?: boolean | EventListenerOptions
    ): void
}
