import { NotificationType } from '../jsonrpc2/messages'

/**
 * The watched files notification is sent from the client to the server when
 * the client detects changes to file watched by the language client.
 */
export namespace DidChangeWatchedFilesNotification {
    export const type = new NotificationType<DidChangeWatchedFilesParams, DidChangeWatchedFilesRegistrationOptions>(
        'workspace/didChangeWatchedFiles'
    )
}

/**
 * The watched files change notification's parameters.
 */
export interface DidChangeWatchedFilesParams {
    /**
     * The actual file events.
     */
    changes: FileEvent[]
}

/**
 * The file event type
 */
export namespace FileChangeType {
    /**
     * The file got created.
     */
    export const Created = 1
    /**
     * The file got changed.
     */
    export const Changed = 2
    /**
     * The file got deleted.
     */
    export const Deleted = 3
}

export type FileChangeType = 1 | 2 | 3

/**
 * An event describing a file change.
 */
export interface FileEvent {
    /**
     * The file's uri.
     */
    uri: string
    /**
     * The change type.
     */
    type: FileChangeType
}

/**
 * Describe options to be used when registered for text document change events.
 */
export interface DidChangeWatchedFilesRegistrationOptions {
    /**
     * The watchers to register.
     */
    watchers: FileSystemWatcher[]
}

export interface FileSystemWatcher {
    /**
     * The  glob pattern to watch
     */
    globPattern: string

    /**
     * The kind of events of interest. If omitted it defaults
     * to WatchKind.Create | WatchKind.Change | WatchKind.Delete
     * which is 7.
     */
    kind?: number
}

export namespace WatchKind {
    /**
     * Interested in create events.
     */
    export const Create = 1

    /**
     * Interested in change events
     */
    export const Change = 2

    /**
     * Interested in delete events
     */
    export const Delete = 4
}
