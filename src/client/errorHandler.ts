import { Message } from '../jsonrpc2/messages'

/**
 * Called by the client when initialization fails to determine how to proceed.
 *
 * @returns true to attempt reinitialization, false otherwise
 */
export type InitializationFailedHandler = (error: Error) => boolean | Promise<boolean>

/**
 * A pluggable error handler that is invoked when the connection encounters an error or is closed.
 */
export interface ErrorHandler {
    /**
     * An error has occurred while writing or reading from the connection.
     *
     * @param error - the error received
     * @param message - the message that was being delivered, if known
     * @param count - how many times this error has occurred (reset after success)
     */
    error(error: Error, message: Message | undefined, count: number | undefined): ErrorAction

    /**
     * The connection to the server got closed.
     */
    closed(): CloseAction | Promise<CloseAction>
}

/** An action to be performed when the connection is producing errors. */
export enum ErrorAction {
    /** Continue running the server. */
    Continue = 1,

    /** Shut down the server. */
    ShutDown = 2,
}

/** An action to be performed when the connection to a server is closed. */
export enum CloseAction {
    /** Don't reconnect to the server. The connection will remain closed. */
    DoNotReconnect = 1,

    /** Reconnect to the server. */
    Reconnect = 2,
}

/** The default error handler. */
export class DefaultErrorHandler implements ErrorHandler {
    private reconnects: number[] = []

    public error(_error: Error, _message: Message, count: number): ErrorAction {
        if (count && count <= 3) {
            return ErrorAction.Continue
        }
        return ErrorAction.ShutDown
    }

    public closed(): CloseAction {
        this.reconnects.push(Date.now())
        if (this.reconnects.length < 5) {
            return CloseAction.Reconnect
        } else {
            const diff = this.reconnects[this.reconnects.length - 1] - this.reconnects[0]
            if (diff <= 3 * 60 * 1000) {
                // The client reconnected 5 times in the last 3 minutes. Do not attempt to reconnect again.
                return CloseAction.DoNotReconnect
            } else {
                this.reconnects.shift()
                return CloseAction.Reconnect
            }
        }
    }
}
