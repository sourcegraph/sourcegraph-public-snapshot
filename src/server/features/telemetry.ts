import { InitializeParams, ServerCapabilities, TelemetryEventNotification } from '../../protocol'
import { IConnection } from '../server'
import { Remote } from './common'

/**
 * Interface to log telemetry events. The events are actually send to the client
 * and the client needs to feed the event into a proper telemetry system.
 */
export interface Telemetry extends Remote {
    /**
     * Log the given data to telemetry.
     *
     * @param data The data to log. Must be a JSON serializable object.
     */
    logEvent(data: any): void
}

export class TelemetryImpl implements Telemetry {
    private _connection?: IConnection

    public attach(connection: IConnection): void {
        this._connection = connection
    }

    public get connection(): IConnection {
        if (!this._connection) {
            throw new Error('Remote is not attached to a connection yet.')
        }
        return this._connection
    }

    public initialize(_params: InitializeParams): void {
        /* noop */
    }

    public fillServerCapabilities(_capabilities: ServerCapabilities): void {
        /* noop */
    }

    public logEvent(data: any): void {
        this.connection.sendNotification(TelemetryEventNotification.type, data)
    }
}
