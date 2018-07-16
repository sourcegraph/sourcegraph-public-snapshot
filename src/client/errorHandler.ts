import { Message, ResponseError } from '../jsonrpc2/messages'
import { InitializeError } from '../protocol'

export type InitializationFailedHandler = (error: ResponseError<InitializeError> | Error | any) => boolean

/**
 * A pluggable error handler that is invoked when the connection is either
 * producing errors or got closed.
 */
export interface ErrorHandler {
    /**
     * An error has occurred while writing or reading from the connection.
     *
     * @param error - the error received
     * @param message - the message to be delivered to the server if know.
     * @param count - a count indicating how often an error is received. Will
     *  be reset if a message got successfully send or received.
     */
    error(error: Error, message: Message, count: number): ErrorAction

    /**
     * The connection to the server got closed.
     */
    closed(): CloseAction
}

/** An action to be performed when the connection is producing errors. */
export enum ErrorAction {
    /** Continue running the server. */
    Continue = 1,
    /** Shut down the server. */
    Shutdown = 2,
}

/** An action to be performed when the connection to a server is closed. */
export enum CloseAction {
    /** Don't restart the server. The connection remains closed. */
    DoNotRestart = 1,
    /** Restart the server. */
    Restart = 2,
}

/** The default error handler. */
export class DefaultErrorHandler implements ErrorHandler {
    private restarts: number[] = []

    public error(_error: Error, _message: Message, count: number): ErrorAction {
        if (count && count <= 3) {
            return ErrorAction.Continue
        }
        return ErrorAction.Shutdown
    }

    public closed(): CloseAction {
        this.restarts.push(Date.now())
        if (this.restarts.length < 5) {
            return CloseAction.Restart
        } else {
            const diff = this.restarts[this.restarts.length - 1] - this.restarts[0]
            if (diff <= 3 * 60 * 1000) {
                // The server crashed 5 times in the last 3 minutes. The server will not be restarted.
                return CloseAction.DoNotRestart
            } else {
                this.restarts.shift()
                return CloseAction.Restart
            }
        }
    }
}
