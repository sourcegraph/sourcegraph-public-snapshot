import { fromEventPattern, Observable } from 'rxjs'

/**
 * Returns an Observable for a WebExtension API event listener.
 * The handler will always return `void`.
 */
export const fromBrowserEvent = <T extends (...args: any[]) => any>(
    emitter: browser.CallbackEventEmitter<T>
): Observable<Parameters<T>> =>
    fromEventPattern(handler => emitter.addListener(handler as T), handler => emitter.removeListener(handler as T))
