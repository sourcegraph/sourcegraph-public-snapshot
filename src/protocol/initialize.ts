import { ClientCapabilities } from './capabilities'
import { ConfigurationCascade } from './configuration'
import { NotificationType, RequestType } from './jsonrpc2/messages'

/**
 * The initialize request is sent from the client to the server. It is sent once as the request after starting up
 * the server. The requests parameter is of type [InitializeParams](#InitializeParams) the response if of type
 * [InitializeResult](#InitializeResult) of a Thenable that resolves to such.
 */
export namespace InitializeRequest {
    export const type = new RequestType<InitializeParams, InitializeResult, void, void>('initialize')
}

/**
 * The initialize parameters
 */
// tslint:disable-next-line:class-name
export interface _InitializeParams {
    /**
     * The capabilities provided by the client (editor or tool)
     */
    capabilities: ClientCapabilities

    /**
     * The configuration at initialization time. If the configuration changes on the client, the client will report
     * the update to the extension by sending a `workspace/didChangeConfiguration`
     * ({@link DidChangeConfigurationNotification}) notification.
     */
    configurationCascade: ConfigurationCascade

    /**
     * The initial trace setting. If omitted trace is disabled ('off').
     */
    trace?: 'off' | 'messages' | 'verbose'
}

export type InitializeParams = _InitializeParams

/**
 * The result returned from an initialize request.
 */
export interface InitializeResult {}

export interface InitializedParams {}

/**
 * The intialized notification is sent from the client to the
 * server after the client is fully initialized and the server
 * is allowed to send requests from the server to the client.
 */
export namespace InitializedNotification {
    export const type = new NotificationType<InitializedParams, void>('initialized')
}

/**
 * A shutdown request is sent from the client to the server.
 * It is sent once when the client decides to shutdown the
 * server. The only notification that is sent after a shutdown request
 * is the exit event.
 */
export namespace ShutdownRequest {
    export const type = new RequestType<null, void, void, void>('shutdown')
}

/**
 * The exit event is sent from the client to the server to
 * ask the server to exit its process.
 */
export namespace ExitNotification {
    export const type = new NotificationType<null, void>('exit')
}
