import { RequestType } from '../jsonrpc2/messages'

/**
 * General parameters to to register for an notification or to register a provider.
 */
export interface Registration {
    /**
     * The id used to register the request. The id can be used to deregister
     * the request again.
     */
    id: string

    /**
     * The method to register for.
     */
    method: string

    /**
     * Options necessary for the registration.
     */
    registerOptions?: any

    /**
     * If true, this registration overwrites an existing registration with the same ID and method. It is an error
     * if overwriteExisting is true and there is no such existing registration.
     *
     * NOTE: This is currently only supported for contributions (when method is "window/contribution"). It is most
     * useful for contributions because it makes a separate unregister message unnecessary and allows the update to
     * be atomic (with no brief period of time in between the unregister and register when the contribution would
     * disappear).
     */
    overwriteExisting?: boolean
}

export interface RegistrationParams {
    registrations: Registration[]
}

/**
 * The `client/registerCapability` request is sent from the server to the client to register a new capability
 * handler on the client side.
 */
export namespace RegistrationRequest {
    export const type = new RequestType<RegistrationParams, void, void, void>('client/registerCapability')
}

/**
 * General parameters to unregister a request or notification.
 */
export interface Unregistration {
    /**
     * The id used to unregister the request or notification. Usually an id
     * provided during the register request.
     */
    id: string

    /**
     * The method to unregister for.
     */
    method: string
}

export interface UnregistrationParams {
    /**
     * NOTE: This typo exists in LSP ("unregisterations", the commonly used English spelling is "unregistrations").
     */
    unregisterations: Unregistration[]
}

/**
 * The `client/unregisterCapability` request is sent from the server to the client to unregister a previously registered capability
 * handler on the client side.
 */
export namespace UnregistrationRequest {
    export const type = new RequestType<UnregistrationParams, void, void, void>('client/unregisterCapability')
}
