import { Observable } from './api'

/**
 * Returns the (synchronously available) value that the Observable emits upon subscription, or throws an error if
 * there is none.
 */
export function observableValue<T>(observable: Observable<T>): T {
    let value!: T
    let set = false
    observable
        .subscribe(v => {
            value = v
            set = true
        })
        .unsubscribe()
    if (!set) {
        // If this error is thrown, there is a bug. Check the Observable's original sources; they should all be
        // available synchronously (e.g., using a BehaviorSubject).
        throw new Error('Observable value was not available synchronously')
    }
    return value
}
