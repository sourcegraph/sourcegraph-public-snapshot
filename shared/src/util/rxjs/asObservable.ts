import { Observable, isObservable, of, throwError } from 'rxjs'

/**
 * Calls a function and returns the result as an Observable.
 */
export function asObservable<T>(fn: () => Observable<T> | T): Observable<T> {
    try {
        const value = fn()
        if (isObservable(value)) {
            return value
        }
        return of(value)
    } catch (error) {
        return throwError(error)
    }
}
