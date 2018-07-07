import { LogMessageNotification, LogMessageParams } from '../../protocol'
import { Client } from '../client'
import { StaticFeature } from './common'

/**
 * Support for server log messages (window/logMessage notifications from the server).
 */
export class WindowLogMessagesFeature implements StaticFeature {
    constructor(
        private client: Client,
        /** Called when the client receives a window/logMessage notification. */
        private logMessage: (params: LogMessageParams) => void
    ) {}

    public initialize(): void {
        // TODO(sqs): no way to unregister this
        this.client.onNotification(LogMessageNotification.type, this.logMessage)
    }
}
