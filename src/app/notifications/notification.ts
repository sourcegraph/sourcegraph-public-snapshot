import { MessageType } from 'sourcegraph/module/protocol'
import { ErrorLike } from '../../errors'

/**
 * A notification message to display to the user.
 */
export interface Notification {
    /** The message or error of the notification. */
    message: string | ErrorLike

    /**
     * The type of the message.
     *
     * @default MessageType.Info
     */
    type?: MessageType

    /** The source of the notification.  */
    source?: string
}
