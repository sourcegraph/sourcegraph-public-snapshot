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

/** The parameters for window/showInput. */
export interface ShowInputParams {
    /** The message to display in the input dialog. */
    message: string

    /** The default value to display in the input field. */
    defaultValue?: string
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
