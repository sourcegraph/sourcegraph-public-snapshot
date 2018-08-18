import { Logger, MessageConnection } from '../../jsonrpc2/connection'
import { InitializeParams, LogMessageNotification, MessageType, ServerCapabilities } from '../../protocol'
import { Connection } from '../server'
import { Remote } from './common'

/**
 * The RemoteConsole interface contains all functions to interact with
 * the tools / clients console or log system. Interally it used `window/logMessage`
 * notifications.
 */
export interface RemoteConsole extends Remote {
    /**
     * Show an error message.
     *
     * @param message The message to show.
     */
    error(message: string): void

    /**
     * Show a warning message.
     *
     * @param message The message to show.
     */
    warn(message: string): void

    /**
     * Show an information message.
     *
     * @param message The message to show.
     */
    info(message: string): void

    /**
     * Log a message.
     *
     * @param message The message to log.
     */
    log(message: string): void
}

export class ConnectionLogger implements Logger, RemoteConsole {
    private _rawConnection?: MessageConnection
    private _connection?: Connection

    public rawAttach(connection: MessageConnection): void {
        this._rawConnection = connection
    }

    public attach(connection: Connection): void {
        this._connection = connection
    }

    public get connection(): Connection {
        if (!this._connection) {
            throw new Error('Remote is not attached to a connection yet.')
        }
        return this._connection
    }

    public fillServerCapabilities(_capabilities: ServerCapabilities): void {
        /* noop */
    }

    public initialize(_params: InitializeParams): void {
        /* noop */
    }

    public error(message: string): void {
        this.send(MessageType.Error, message)
    }

    public warn(message: string): void {
        this.send(MessageType.Warning, message)
    }

    public info(message: string): void {
        this.send(MessageType.Info, message)
    }

    public log(message: string): void {
        this.send(MessageType.Log, message)
    }

    private send(type: MessageType, message: string): void {
        if (this._rawConnection) {
            this._rawConnection.sendNotification(LogMessageNotification.type, { type, message })
        }
    }
}
