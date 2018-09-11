/**
 * The parameters for the context/update notification, which is sent from the server to the client to update
 * context values.
 */
export interface ContextUpdateParams {
    /**
     * The updates to apply to the context. If a context property's value is null, it is deleted from the context.
     */
    updates: { [key: string]: string | number | boolean | null }
}

/**
 * The context/update notification, which is sent from the server to the client to update context values.
 */
export namespace ContextUpdateNotification {
    export const type = 'context/update'
}
