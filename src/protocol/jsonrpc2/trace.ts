import { NotificationMessage, RequestMessage, ResponseMessage } from './messages'

// Copied from vscode-jsonrpc to avoid adding extraneous dependencies.

export enum Trace {
    Off = 'off',
    Messages = 'messages',
    Verbose = 'verbose',
}

/** Records messages sent and received on a JSON-RPC 2.0 connection. */
export interface Tracer {
    log(message: string, details?: string): void
    requestSent(message: RequestMessage): void
    requestReceived(message: RequestMessage): void
    notificationSent(message: NotificationMessage): void
    notificationReceived(message: NotificationMessage): void
    responseSent(message: ResponseMessage, request: RequestMessage, startTime: number): void
    responseCanceled(message: ResponseMessage, request: RequestMessage, cancelMessage: NotificationMessage): void
    responseReceived(message: ResponseMessage, request: RequestMessage | string, startTime: number): void
    unknownResponseReceived(message: ResponseMessage): void
}

/** A tracer that implements the Tracer interface with noop methods. */
export const noopTracer: Tracer = {
    log: () => void 0,
    requestSent: () => void 0,
    requestReceived: () => void 0,
    notificationSent: () => void 0,
    notificationReceived: () => void 0,
    responseSent: () => void 0,
    responseCanceled: () => void 0,
    responseReceived: () => void 0,
    unknownResponseReceived: () => void 0,
}

/** A tracer that implements the Tracer interface with console API calls, intended for a web browser. */
export class BrowserConsoleTracer implements Tracer {
    public constructor(private name: string) {}

    private prefix(level: 'info' | 'error', label: string, title: string): string[] {
        let color: string
        let backgroundColor: string
        if (level === 'info') {
            color = '#000'
            backgroundColor = '#eee'
        } else {
            color = 'white'
            backgroundColor = 'red'
        }
        return [
            '%c%s%c %s%c%s%c',
            `font-weight:bold;background-color:#d8f7ff;color:black`,
            this.name,
            '',
            label,
            `background-color:${backgroundColor};color:${color};font-weight:bold`,
            title,
            '',
        ]
    }

    public log(message: string, details?: string): void {
        if (details) {
            ;(console.groupCollapsed as any)(...this.prefix('info', 'log', ''), message)
            console.log(details)
            console.groupEnd()
        } else {
            console.log(...this.prefix('info', 'log', ''), message)
        }
    }

    public requestSent(message: RequestMessage): void {
        console.log(...this.prefix('info', `◀◀ sent request #${message.id}: `, message.method), message.params)
    }

    public requestReceived(message: RequestMessage): void {
        console.log(...this.prefix('info', `▶▶ recv request #${message.id}: `, message.method), message.params)
    }

    public notificationSent(message: NotificationMessage): void {
        console.log(...this.prefix('info', `◀◀ sent notif: `, message.method), message.params)
    }

    public notificationReceived(message: NotificationMessage): void {
        console.log(...this.prefix('info', `▶▶ recv notif: `, message.method), message.params)
    }

    public responseSent(message: ResponseMessage, request: RequestMessage, startTime: number): void {
        const prefix = this.prefix(
            message.error ? 'error' : 'info',
            `◀▶ sent response #${message.id}: `,
            typeof request === 'string' ? request : request.method
        )
        ;(console.groupCollapsed as any)(...prefix)
        if (message.error) {
            console.log('Error:', message.error)
        } else {
            console.log('Result:', message.result)
        }
        console.log('Request:', request)
        console.log('Duration: %d msec', Date.now() - startTime)
        console.groupEnd()
    }

    public responseCanceled(
        _message: ResponseMessage,
        request: RequestMessage,
        _cancelMessage: NotificationMessage
    ): void {
        console.log(...this.prefix('info', '× cancel: ', request.method))
    }

    public responseReceived(message: ResponseMessage, request: RequestMessage | string, startTime: number): void {
        const prefix = this.prefix(
            message.error ? 'error' : 'info',
            `◀▶ recv response #${message.id}: `,
            typeof request === 'string' ? request : request.method
        )
        if (typeof request === 'string') {
            console.log(...prefix, message.error || message.result)
        } else {
            ;(console.groupCollapsed as any)(...prefix)
            if (message.error) {
                console.log('Error:', message.error)
            } else {
                console.log('Result:', message.result)
            }
            console.log('Request:', request)
            console.log('Duration: %d msec', Date.now() - startTime)
            console.groupEnd()
        }
    }

    public unknownResponseReceived(message: ResponseMessage): void {
        console.log(...this.prefix('error', 'UNKNOWN', ''), message)
    }
}
