import { NotificationType } from '../jsonrpc2/messages'

// Copied from vscode-jsonrpc to avoid adding extraneous dependencies.

export enum Trace {
    Off,
    Messages,
    Verbose,
}

export type TraceValues = 'off' | 'messages' | 'verbose'
export namespace Trace {
    export function fromString(value: string): Trace {
        value = value.toLowerCase()
        switch (value) {
            // tslint:disable:no-unnecessary-qualifier
            case 'off':
                return Trace.Off
            case 'messages':
                return Trace.Messages
            case 'verbose':
                return Trace.Verbose
            default:
                return Trace.Off
            // tslint:enable:no-unnecessary-qualifier
        }
    }

    export function toString(value: Trace): TraceValues {
        switch (value) {
            // tslint:disable:no-unnecessary-qualifier
            case Trace.Off:
                return 'off'
            case Trace.Messages:
                return 'messages'
            case Trace.Verbose:
                return 'verbose'
            default:
                return 'off'
            // tslint:enable:no-unnecessary-qualifier
        }
    }
}

export interface SetTraceParams {
    value: TraceValues
}

export namespace SetTraceNotification {
    export const type = new NotificationType<SetTraceParams, void>('$/setTraceNotification')
}

export interface LogTraceParams {
    message: string
    verbose?: string
}

export namespace LogTraceNotification {
    export const type = new NotificationType<LogTraceParams, void>('$/logTraceNotification')
}

export interface Tracer {
    log(message: string, data?: string): void
}
