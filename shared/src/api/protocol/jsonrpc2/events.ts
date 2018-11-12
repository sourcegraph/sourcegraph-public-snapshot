import { Unsubscribable } from 'rxjs'

/**
 * Represents a typed event.
 */
export type Event<T> = (listener: (e: T) => any, thisArgs?: any) => Unsubscribable

export namespace Event {
    const _unsubscribable = {
        unsubscribe(): void {
            /* noop */
        },
    }
    export const None: Event<any> = () => _unsubscribable
}

class CallbackList {
    private _callbacks: ((...args: any[]) => any)[] | undefined
    private _contexts: any[] | undefined

    public add(callback: (...args: any[]) => any, context: any = null, bucket?: Unsubscribable[]): void {
        if (!this._callbacks) {
            this._callbacks = []
            this._contexts = []
        }
        this._callbacks.push(callback)
        this._contexts!.push(context)

        if (Array.isArray(bucket)) {
            bucket.push({ unsubscribe: () => this.remove(callback, context) })
        }
    }

    public remove(callback: (...args: any[]) => any, context: any = null): void {
        if (!this._callbacks) {
            return
        }

        let foundCallbackWithDifferentContext = false
        for (let i = 0, len = this._callbacks.length; i < len; i++) {
            if (this._callbacks[i] === callback) {
                if (this._contexts![i] === context) {
                    // callback & context match => remove it
                    this._callbacks.splice(i, 1)
                    this._contexts!.splice(i, 1)
                    return
                } else {
                    foundCallbackWithDifferentContext = true
                }
            }
        }

        if (foundCallbackWithDifferentContext) {
            throw new Error('When adding a listener with a context, you should remove it with the same context')
        }
    }

    public invoke(...args: any[]): any[] {
        if (!this._callbacks) {
            return []
        }

        const ret: any[] = []
        const callbacks = this._callbacks.slice(0)
        const contexts = this._contexts!.slice(0)
        for (let i = 0; i < callbacks.length; i++) {
            try {
                ret.push(callbacks[i].apply(contexts[i], args))
            } catch (e) {
                console.error(e)
            }
        }
        return ret
    }

    public isEmpty(): boolean {
        return !this._callbacks || this._callbacks.length === 0
    }

    public unsubscribe(): void {
        this._callbacks = undefined
        this._contexts = undefined
    }
}

export class Emitter<T> {
    private static _noop = () => void 0

    private _event?: Event<T>
    private _callbacks: CallbackList | undefined

    /**
     * For the public to allow to subscribe
     * to events from this Emitter
     */
    public get event(): Event<T> {
        if (!this._event) {
            this._event = (listener: (e: T) => any, thisArgs?: any, Unsubscribables?: Unsubscribable[]) => {
                if (!this._callbacks) {
                    this._callbacks = new CallbackList()
                }
                this._callbacks.add(listener, thisArgs)

                let result: Unsubscribable
                result = {
                    unsubscribe: () => {
                        this._callbacks!.remove(listener, thisArgs)
                        result.unsubscribe = Emitter._noop
                    },
                }
                if (Array.isArray(Unsubscribables)) {
                    Unsubscribables.push(result)
                }

                return result
            }
        }
        return this._event
    }

    /**
     * To be kept private to fire an event to
     * subscribers
     */
    public fire(event: T): any {
        if (this._callbacks) {
            this._callbacks.invoke.call(this._callbacks, event)
        }
    }

    public unsubscribe(): void {
        if (this._callbacks) {
            this._callbacks.unsubscribe()
            this._callbacks = undefined
        }
    }
}
