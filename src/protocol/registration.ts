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

/**
 * Static registration options to be returned in the initialize
 * request.
 */
export interface StaticRegistrationOptions {
    /**
     * The id used to register the request. The id can be used to deregister
     * the request again. See also Registration#id.
     */
    id?: string
}
