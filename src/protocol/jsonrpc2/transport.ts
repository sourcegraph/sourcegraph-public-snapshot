import { isFunction } from '../../util'
import { Emitter, Event } from './events'
import { Message } from './messages'

// Copied from vscode-jsonrpc to avoid adding extraneous dependencies.

export type DataCallback = (data: Message) => void

export interface PartialMessageInfo {
    readonly messageToken: number
    readonly waitingTime: number
}

export interface MessageReader {
    readonly onError: Event<Error>
    readonly onClose: Event<void>
    readonly onPartialMessage: Event<PartialMessageInfo>
    listen(callback: DataCallback): void
    unsubscribe(): void
}

export namespace MessageReader {
    export function is(value: any): value is MessageReader {
        const candidate: MessageReader = value
        return (
            candidate &&
            // tslint:disable-next-line:no-unbound-method
            isFunction(candidate.listen) &&
            // tslint:disable-next-line:no-unbound-method
            isFunction(candidate.unsubscribe) &&
            isFunction(candidate.onError) &&
            isFunction(candidate.onClose) &&
            isFunction(candidate.onPartialMessage)
        )
    }
}

export abstract class AbstractMessageReader {
    private errorEmitter: Emitter<Error>
    private closeEmitter: Emitter<void>

    private partialMessageEmitter: Emitter<PartialMessageInfo>

    constructor() {
        this.errorEmitter = new Emitter<Error>()
        this.closeEmitter = new Emitter<void>()
        this.partialMessageEmitter = new Emitter<PartialMessageInfo>()
    }

    public unsubscribe(): void {
        this.errorEmitter.unsubscribe()
        this.closeEmitter.unsubscribe()
    }

    public get onError(): Event<Error> {
        return this.errorEmitter.event
    }

    protected fireError(error: any): void {
        this.errorEmitter.fire(this.asError(error))
    }

    public get onClose(): Event<void> {
        return this.closeEmitter.event
    }

    protected fireClose(): void {
        this.closeEmitter.fire(undefined)
    }

    public get onPartialMessage(): Event<PartialMessageInfo> {
        return this.partialMessageEmitter.event
    }

    protected firePartialMessage(info: PartialMessageInfo): void {
        this.partialMessageEmitter.fire(info)
    }

    private asError(error: any): Error {
        if (error instanceof Error) {
            return error
        }
        return new Error(
            `Reader received error. Reason: ${typeof error.message === 'string' ? error.message : 'unknown'}`
        )
    }
}

export interface MessageWriter {
    readonly onError: Event<[Error, Message | undefined, number | undefined]>
    readonly onClose: Event<void>
    write(msg: Message): void
    unsubscribe(): void
}

export namespace MessageWriter {
    export function is(value: any): value is MessageWriter {
        const candidate: MessageWriter = value
        return (
            candidate &&
            // tslint:disable-next-line:no-unbound-method
            isFunction(candidate.unsubscribe) &&
            isFunction(candidate.onClose) &&
            isFunction(candidate.onError) &&
            // tslint:disable-next-line:no-unbound-method
            isFunction(candidate.write)
        )
    }
}

export abstract class AbstractMessageWriter {
    private errorEmitter = new Emitter<[Error, Message | undefined, number | undefined]>()
    private closeEmitter = new Emitter<void>()

    public unsubscribe(): void {
        this.errorEmitter.unsubscribe()
        this.closeEmitter.unsubscribe()
    }

    public get onError(): Event<[Error, Message | undefined, number | undefined]> {
        return this.errorEmitter.event
    }

    protected fireError(error: any, message?: Message, count?: number): void {
        this.errorEmitter.fire([this.asError(error), message, count])
    }

    public get onClose(): Event<void> {
        return this.closeEmitter.event
    }

    protected fireClose(): void {
        this.closeEmitter.fire(undefined)
    }

    private asError(error: any): Error {
        if (error instanceof Error) {
            return error
        }
        return new Error(
            `Writer received error. Reason: ${typeof error.message === 'string' ? error.message : 'unknown'}`
        )
    }
}
