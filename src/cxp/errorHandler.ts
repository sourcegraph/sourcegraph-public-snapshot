import {
    CloseAction,
    ErrorAction,
    ErrorHandler as CXPErrorHandler,
    InitializationFailedHandler,
} from 'cxp/module/client/errorHandler'
import { Message, ResponseError } from 'cxp/module/jsonrpc2/messages'
import { InitializeError } from 'cxp/module/protocol'
import { log } from './log'

interface CXPInitializationFailedHandler {
    initializationFailed: InitializationFailedHandler
}

/** The CXP client initializion-failed and error handler. */
export class ErrorHandler implements CXPInitializationFailedHandler, CXPErrorHandler {
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

    public initializationFailed(err: ResponseError<InitializeError> | Error | any): boolean | Promise<boolean> {
        log('error', this.extensionID, err)

        const EINVALIDREQUEST = -32600 // JSON-RPC 2.0 error code

        if (
            isResponseError(err) &&
            ((err.message.includes('dial tcp') && err.message.includes('connect: connection refused')) ||
                (err.code === EINVALIDREQUEST && err.message.includes('client proxy handler is already initialized')))
        ) {
            return false
        }

        const retry = isResponseError(err) && !!err.data && err.data.retry && this.connectionTimestamps.length === 0
        return delayed(retry, this.computeDelayBeforeRetry())
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

function isResponseError(err: any): err is ResponseError<InitializeError> {
    return 'code' in err && 'message' in err
}

function delayed<T>(value: T, msec: number): Promise<T> {
    return new Promise(resolve => {
        setTimeout(() => resolve(value), msec)
    })
}
