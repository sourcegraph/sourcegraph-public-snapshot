import { Emitter, Event } from './events'

// Copied from vscode-jsonrpc to avoid adding extraneous dependencies.

export interface CancelParams {
    /**
     * The request id to cancel.
     */
    id: number | string
}

export namespace CancelNotification {
    export const type = '$/cancelRequest'
}

/**
 * Defines a CancellationToken. This interface is not
 * intended to be implemented. A CancellationToken must
 * be created via a CancellationTokenSource.
 */
export interface CancellationToken {
    /**
     * Is `true` when the token has been cancelled, `false` otherwise.
     */
    readonly isCancellationRequested: boolean

    /**
     * An [event](#Event) which fires upon cancellation.
     */
    readonly onCancellationRequested: Event<any>
}

export namespace CancellationToken {
    export const None: CancellationToken = Object.freeze({
        isCancellationRequested: false,
        onCancellationRequested: Event.None,
    })

    export const Cancelled: CancellationToken = Object.freeze({
        isCancellationRequested: true,
        onCancellationRequested: Event.None,
    })

    export function is(value: any): value is CancellationToken {
        const candidate = value as CancellationToken
        return (
            candidate &&
            // tslint:disable-next-line:no-unnecessary-qualifier
            (candidate === CancellationToken.None ||
                // tslint:disable-next-line:no-unnecessary-qualifier
                candidate === CancellationToken.Cancelled ||
                (typeof candidate.isCancellationRequested === 'boolean' && !!candidate.onCancellationRequested))
        )
    }
}

const shortcutEvent: Event<any> = Object.freeze(
    (callback: (...args: any[]) => any, context?: any): any => {
        const handle = setTimeout(callback.bind(context), 0)
        return {
            unsubscribe(): void {
                clearTimeout(handle)
            },
        }
    }
)

class MutableToken implements CancellationToken {
    private _isCancelled = false
    private _emitter: Emitter<any> | undefined

    public cancel(): void {
        if (!this._isCancelled) {
            this._isCancelled = true
            if (this._emitter) {
                this._emitter.fire(undefined)
                this._emitter = undefined
            }
        }
    }

    public get isCancellationRequested(): boolean {
        return this._isCancelled
    }

    public get onCancellationRequested(): Event<any> {
        if (this._isCancelled) {
            return shortcutEvent
        }
        if (!this._emitter) {
            this._emitter = new Emitter<any>()
        }
        return this._emitter.event
    }
}

export class CancellationTokenSource {
    private _token?: CancellationToken

    public get token(): CancellationToken {
        if (!this._token) {
            // be lazy and create the token only when
            // actually needed
            this._token = new MutableToken()
        }
        return this._token
    }

    public cancel(): void {
        if (!this._token) {
            // save an object by returning the default
            // cancelled token when cancellation happens
            // before someone asks for the token
            this._token = CancellationToken.Cancelled
        } else {
            ;(this._token as MutableToken).cancel()
        }
    }

    public unsubscribe(): void {
        this.cancel()
    }
}
