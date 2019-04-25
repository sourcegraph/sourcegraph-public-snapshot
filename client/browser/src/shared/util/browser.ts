import { fromEventPattern, Observable } from 'rxjs'

/**
 * Returns an Observable for a WebExtension API event listener.
 * The handler will always return `void`.
 */
export const fromBrowserEvent = <T extends (...args: any[]) => any>(
    event: browser.CallbackEventEmitter<T>
): Observable<Parameters<T>> =>
    fromEventPattern(handler => event.addListener(handler as T), handler => event.removeListener(handler as T))
