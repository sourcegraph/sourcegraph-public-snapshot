import { NotificationType } from './jsonrpc2/messages'

/**
 * The telemetry event notification is sent from the server to the client to ask
 * the client to log telemetry data.
 */
export namespace TelemetryEventNotification {
    export const type = new NotificationType<any, void>('telemetry/event')
}
