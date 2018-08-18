import { LogTraceNotification, Trace } from '../../jsonrpc2/trace'
import { InitializeParams, ServerCapabilities } from '../../protocol'
import { Connection } from '../server'
import { Remote } from './common'

/**
 * Interface to log traces to the client. The events are sent to the client and the
 * client needs to log the trace events.
 */
export interface Tracer extends Remote {
    /**
     * Log the given data to the trace Log
     */
    log(message: string, verbose?: string): void
}

export class TracerImpl implements Tracer {
    private _trace: Trace
    private _connection?: Connection

    constructor() {
        this._trace = Trace.Off
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

    public initialize(_params: InitializeParams): void {
        /* noop */
    }

    public fillServerCapabilities(_capabilities: ServerCapabilities): void {
        /* noop */
    }

    public set trace(value: Trace) {
        this._trace = value
    }

    public log(message: string, verbose?: string): void {
        if (this._trace === Trace.Off) {
            return
        }
        this.connection.sendNotification(LogTraceNotification.type, {
            message,
            verbose: this._trace === Trace.Verbose ? verbose : undefined,
        })
    }
}
