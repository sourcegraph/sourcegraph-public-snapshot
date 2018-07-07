import { HandlerResult, NotificationHandler, RequestHandler } from '../jsonrpc2/handlers'
import { NotificationType, RequestType } from '../jsonrpc2/messages'

export interface WorkspaceFoldersInitializeParams {
    /**
     * The actual configured workspace folders.
     */
    workspaceFolders: WorkspaceFolder[] | null
}

export interface WorkspaceFoldersClientCapabilities {
    /**
     * The workspace client capabilities
     */
    workspace?: {
        /**
         * The client has support for workspace folders
         */
        workspaceFolders?: boolean
    }
}

export interface WorkspaceFoldersServerCapabilities {
    /**
     * The workspace server capabilities
     */
    workspace?: {
        workspaceFolders?: {
            /**
             * The Server has support for workspace folders
             */
            supported?: boolean

            /**
             * Whether the server wants to receive workspace folder
             * change notifications.
             *
             * If a strings is provided the string is treated as a ID
             * under which the notification is registed on the client
             * side. The ID can be used to unregister for these events
             * using the `client/unregisterCapability` request.
             */
            changeNotifications?: string | boolean
        }
    }
}

export interface WorkspaceFolder {
    /**
     * The associated URI for this workspace folder.
     */
    uri: string

    /**
     * The name of the workspace folder. Defaults to the
     * uri's basename.
     */
    name: string
}

/**
 * The `workspace/workspaceFolders` is sent from the server to the client to fetch the open workspace folders.
 */
export namespace WorkspaceFoldersRequest {
    export const type = new RequestType<null, WorkspaceFolder[] | null, void, void>('workspace/workspaceFolders')
    export type HandlerSignature = RequestHandler<null, WorkspaceFolder[] | null, void>
    export type MiddlewareSignature = (next: HandlerSignature) => HandlerResult<WorkspaceFolder[] | null, void>
}

/**
 * The `workspace/didChangeWorkspaceFolders` notification is sent from the client to the server when the workspace
 * folder configuration changes.
 */
export namespace DidChangeWorkspaceFoldersNotification {
    export const type = new NotificationType<DidChangeWorkspaceFoldersParams, void>(
        'workspace/didChangeWorkspaceFolders'
    )
    export type HandlerSignature = NotificationHandler<DidChangeWorkspaceFoldersParams>
    export type MiddlewareSignature = (params: DidChangeWorkspaceFoldersParams, next: HandlerSignature) => void
}

/**
 * The parameters of a `workspace/didChangeWorkspaceFolders` notification.
 */
export interface DidChangeWorkspaceFoldersParams {
    /**
     * The actual workspace folder change event.
     */
    event: WorkspaceFoldersChangeEvent
}

/**
 * The workspace folder change event.
 */
export interface WorkspaceFoldersChangeEvent {
    /**
     * The array of added workspace folders
     */
    added: WorkspaceFolder[]

    /**
     * The array of the removed workspace folders
     */
    removed: WorkspaceFolder[]
}
