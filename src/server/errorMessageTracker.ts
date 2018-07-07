import { RemoteWindow } from './features/window'

/**
 * Helps tracking error message. Equal occurences of the same
 * message are only stored once. This class is for example
 * useful if text documents are validated in a loop and equal
 * error message should be folded into one.
 */
export class ErrorMessageTracker {
    private _messages: { [key: string]: number }

    constructor() {
        this._messages = Object.create(null)
    }

    /**
     * Add a message to the tracker.
     *
     * @param message The message to add.
     */
    public add(message: string): void {
        let count: number = this._messages[message]
        if (!count) {
            count = 0
        }
        count++
        this._messages[message] = count
    }

    /**
     * Send all tracked messages to the connection's window.
     *
     * @param connection The connection established between client and server.
     */
    public sendErrors(connection: { window: RemoteWindow }): void {
        for (const message of Object.keys(this._messages)) {
            connection.window.showErrorMessage(message)
        }
    }
}
