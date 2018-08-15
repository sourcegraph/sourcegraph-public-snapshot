import { RequestHandler } from '../jsonrpc2/handlers'
import { NotificationType, RequestType } from '../jsonrpc2/messages'
import { NextSignature } from '../types/middleware'

/**
 * A key path that refers to a location in a JSON document.
 *
 * Each successive array element specifies an index in an object or array to descend into. For example, in the
 * object `{"a": ["x", "y"]}`, the key path `["a", 1]` refers to the value `"y"`.
 */
export type KeyPath = (string | number)[]

/** The parameters for the configuration/update request. */
export interface ConfigurationUpdateParams {
    /** The key path to the value. */
    path: KeyPath

    /** The new value to insert at the key path. */
    value: any
}

/**
 * The configuration/update request, which the server sends to the client to update the client's configuration.
 */
export namespace ConfigurationUpdateRequest {
    export const type = new RequestType<ConfigurationUpdateParams, void, void, void>('configuration/update')
}

export interface ConfigurationClientCapabilities {
    configuration?: { update?: boolean }

    /**
     * The workspace client capabilities
     */
    workspace?: {
        /**
         * The client supports `workspace/configuration` requests.
         */
        configuration?: boolean
    }
}

/**
 * The 'workspace/configuration' request is sent from the server to the client to fetch a certain
 * configuration setting.
 */
export namespace ConfigurationRequest {
    export const type = new RequestType<ConfigurationParams, any[], void, void>('workspace/configuration')
    export type HandlerSignature = RequestHandler<ConfigurationParams, any[], void>
    export type MiddlewareSignature = NextSignature<ConfigurationParams, any[]>
}

export interface ConfigurationItem {
    /**
     * The scope to get the configuration section for.
     */
    scopeUri?: string

    /**
     * The configuration section asked for.
     */
    section?: string
}

/**
 * The parameters of a configuration request.
 */
export interface ConfigurationParams {
    items: ConfigurationItem[]
}

/**
 * The configuration change notification is sent from the client to the server
 * when the client's configuration has changed. The notification contains
 * the changed configuration as defined by the language client.
 */
export namespace DidChangeConfigurationNotification {
    export const type = new NotificationType<DidChangeConfigurationParams, DidChangeConfigurationRegistrationOptions>(
        'workspace/didChangeConfiguration'
    )
}

export interface DidChangeConfigurationRegistrationOptions {
    section?: string | string[]
}

/**
 * The parameters of a change configuration notification.
 */
export interface DidChangeConfigurationParams {
    /**
     * The new configuration cascade, after the change was applied.
     */
    configurationCascade: ConfigurationCascade
}

/**
 * The merged configuration from a configuration cascade.
 */
export interface ConfigurationCascade {
    /** The final settings, merged from all levels in the cascade. */
    merged: any
}
