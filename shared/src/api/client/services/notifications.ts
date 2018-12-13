import { Observable, Subject } from 'rxjs'
import { Progress } from 'sourcegraph'

interface PromiseCallback<T> {
    resolve: (p: T | Promise<T>) => void
}

/**
 * The type of a message.
 */
export enum MessageType {
    /**
     * An error message.
     */
    Error,
    /**
     * A warning message.
     */
    Warning,
    /**
     * An information message.
     */
    Info,
    /**
     * A log message.
     */
    Log,
}

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
interface LogMessageParams {
    /**
     * The message type. See {@link MessageType}
     */
    type: MessageType

    /**
     * The actual message
     */
    message: string
}

type ShowMessageRequest = ShowMessageRequestParams & PromiseCallback<MessageActionItem | null>

type ShowInputRequest = ShowInputParams & PromiseCallback<string | null>

export class NotificationsService {
    /** Log messages from extensions. */
    public readonly logMessages = new Subject<LogMessageParams>()

    /** Messages from extensions intended for display to the user. */
    public readonly showMessages = new Subject<ShowMessageParams>()

    /** Messages from extensions requesting the user to select an action. */
    public readonly showMessageRequests = new Subject<ShowMessageRequest>()
    /** Messages from extensions requesting the user to select an action. */
    public readonly progresses = new Subject<{ title?: string; progress: Observable<Progress> }>()

    /** Messages from extensions requesting text input from the user. */
    public readonly showInputs = new Subject<ShowInputRequest>()
}
