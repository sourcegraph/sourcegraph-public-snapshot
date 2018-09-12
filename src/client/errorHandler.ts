import { CloseAction, ErrorAction, ErrorHandler as _ErrorHandler } from 'sourcegraph/module/client/errorHandler'
import { Message } from 'sourcegraph/module/protocol/jsonrpc2/messages'
import { log } from './log'

/** The extension client initialization-failed and error handler. */
export class ErrorHandler implements _ErrorHandler {
    /** The number of connection times to record. */
    private static MAX_CONNECTION_TIMESTAMPS = 4

    /** The timestamps of the last connection initiation times, with the 0th element being the oldest. */
    private connectionTimestamps: number[] = [Date.now()]

    public constructor(private extensionID: string) {}

    public error(err: Error, message: Message, count: number): ErrorAction {
        log(
            'error',
            `${this.extensionID}${count > 1 ? ` (count: ${count})` : ''}`,
            err,
            message ? { message } : undefined
        )

        if (err.message && err.message.includes('got unsubscribed')) {
            return ErrorAction.ShutDown
        }

        // Language servers differ in when they decide to return an error vs. just return an empty result. This
        // constant here is a guess that should be adjusted.
        if (count && count <= 5) {
            return ErrorAction.Continue
        }
        return ErrorAction.ShutDown
    }

    private computeDelayBeforeRetry(): number {
        const lastRestart: number | undefined = this.connectionTimestamps[this.connectionTimestamps.length - 1]
        const now = Date.now()

        // Bound the size of the array.
        if (this.connectionTimestamps.length === ErrorHandler.MAX_CONNECTION_TIMESTAMPS) {
            this.connectionTimestamps.shift()
        }
        this.connectionTimestamps.push(now)

        const diff = now - (lastRestart || 0)
        if (diff <= 10 * 1000) {
            // If the connection was created less than 10 seconds ago, wait longer to restart to avoid excessive
            // attempts.
            return 2500
        }
        // Otherwise restart after a shorter period.
        return 500
    }

    public closed(): CloseAction | Promise<CloseAction> {
        if (this.connectionTimestamps.length === ErrorHandler.MAX_CONNECTION_TIMESTAMPS) {
            const diff = this.connectionTimestamps[this.connectionTimestamps.length - 1] - this.connectionTimestamps[0]
            if (diff <= 60 * 1000) {
                // Stop restarting the server if it has restarted n times in the last minute.
                return CloseAction.DoNotReconnect
            }
        }

        return delayed(CloseAction.Reconnect, this.computeDelayBeforeRetry())
    }
}

function delayed<T>(value: T, msec: number): Promise<T> {
    return new Promise(resolve => {
        setTimeout(() => resolve(value), msec)
    })
}
