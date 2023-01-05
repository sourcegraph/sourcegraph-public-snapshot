import { Observable } from 'rxjs'

import type { NotificationType, Progress } from '../codeintel/legacy-extensions/api'

/**
 * A notification message to display to the user.
 */
export interface Notification {
    /** The message of the notification. */
    message?: string

    /**
     * The type of the message.
     */
    type: NotificationType

    /** The source of the notification.  */
    source?: string

    /**
     * Progress updates to show in this notification (progress bar and status messages).
     * If this Observable errors, the notification will be changed to an error type.
     */
    progress?: Observable<Progress>
}
