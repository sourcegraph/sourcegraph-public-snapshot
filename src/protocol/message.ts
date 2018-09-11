/**
 * The message type
 */
export namespace MessageType {
    /**
     * An error message.
     */
    export const Error = 1
    /**
     * A warning message.
     */
    export const Warning = 2
    /**
     * An information message.
     */
    export const Info = 3
    /**
     * A log message.
     */
    export const Log = 4
}

export type MessageType = 1 | 2 | 3 | 4

/**
 * The parameters of a notification message.
 */
export interface ShowMessageParams {
    /**
     * The message type. See {@link MessageType}
     */
    type: MessageType

    /**
     * The actual message
     */
    message: string
}

/**
 * The show message notification is sent from a server to a client to ask
 * the client to display a particular message in the user interface.
 */
export namespace ShowMessageNotification {
    export const type = 'window/showMessage'
}

export interface MessageActionItem {
    /**
     * A short title like 'Retry', 'Open Log' etc.
     */
    title: string
}

export interface ShowMessageRequestParams {
    /**
     * The message type. See {@link MessageType}
     */
    type: MessageType

    /**
     * The actual message
     */
    message: string

    /**
     * The message action items to present.
     */
    actions?: MessageActionItem[]
}

/**
 * The show message request is sent from the server to the client to show a message and a set of actions to the
 * user.
 */
export namespace ShowMessageRequest {
    export const type = 'window/showMessageRequest'
}

/** The parameters for window/showInput. */
export interface ShowInputParams {
    /** The message to display in the input dialog. */
    message: string

    /** The default value to display in the input field. */
    defaultValue?: string
}

/**
 * The show input request is sent from the server to the client to show a message and prompt the user for input.
 */
export namespace ShowInputRequest {
    export const type = 'window/showInput'
}

/**
 * The log message notification is sent from the server to the client to ask
 * the client to log a particular message.
 */
export namespace LogMessageNotification {
    export const type = 'window/logMessage'
}

/**
 * The log message parameters.
 */
export interface LogMessageParams {
    /**
     * The message type. See {@link MessageType}
     */
    type: MessageType

    /**
     * The actual message
     */
    message: string
}
