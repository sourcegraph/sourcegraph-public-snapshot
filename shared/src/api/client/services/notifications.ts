import { Observable, Subject } from 'rxjs'
import * as sourcegraph from 'sourcegraph'

/**
 * The type of a notification.
 * This is needed because if sourcegraph.NotificationType enum values are referenced,
 * the `sourcegraph` module import at the top of the file is emitted in the generated code.
 */
export const NotificationType: typeof sourcegraph.NotificationType = {
    Error: 1,
    Warning: 2,
    Info: 3,
    Log: 4,
    Success: 5,
}

interface PromiseCallback<T> {
    resolve: (p: T | Promise<T>) => void
}

/**
 * The parameters of a notification message.
 */
export interface ShowNotificationParams {
    /**
     * The notification type. See {@link NotificationType}
     */
    type: sourcegraph.NotificationType

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
     * The message type. See {@link NotificationType}
     */
    type: sourcegraph.NotificationType

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

type ShowMessageRequest = ShowMessageRequestParams & PromiseCallback<MessageActionItem | null>

type ShowInputRequest = ShowInputParams & PromiseCallback<string | null>

export class NotificationsService {
    /** Messages from extensions intended for display to the user. */
    public readonly showMessages = new Subject<ShowNotificationParams>()

    /** Messages from extensions requesting the user to select an action. */
    public readonly showMessageRequests = new Subject<ShowMessageRequest>()
    /** Messages from extensions requesting the user to select an action. */
    public readonly progresses = new Subject<{ title?: string; progress: Observable<sourcegraph.Progress> }>()

    /** Messages from extensions requesting text input from the user. */
    public readonly showInputs = new Subject<ShowInputRequest>()
}
