import { BehaviorSubject, fromEvent, NextObserver, Observable } from 'rxjs'
import { filter } from 'rxjs/operators'

/**
 * An RxJS subject that is backed by a localStorage item.
 *
 * It does not emit when the localStorage item is changed in the same window (because that does not trigger the
 * "storage" event per the Web Storage API specification). To emit for changes in the same window, call
 * {@link LocalStorageSubject#next}.
 */
export class LocalStorageSubject<T> extends Observable<T>
    implements NextObserver<T>, Pick<BehaviorSubject<T>, 'value'> {
    constructor(private key: string, private defaultValue: T) {
        super(subscriber => {
            subscriber.next(this.value)
            return fromEvent<StorageEvent>(window, 'storage')
                .pipe(filter(event => event.key === key))
                .subscribe(event => {
                    subscriber.next(parseValue(event.newValue, defaultValue))
                })
        })
    }

    public next(value: T): void {
        const v = JSON.stringify(value)
        localStorage.setItem(this.key, v)
        // Does not set oldValue or other StorageEventInit keys because we don't need them.
        window.dispatchEvent(new StorageEvent('storage', { key: this.key, newValue: v }))
    }

    public get value(): T {
        return parseValue(localStorage.getItem(this.key), this.defaultValue)
    }
}

function parseValue<T>(v: string | null, defaultValue: T): T {
    if (v === null) {
        return defaultValue
    }
    try {
        // False positive https://github.com/typescript-eslint/typescript-eslint/issues/1269
        // eslint-disable-next-line @typescript-eslint/return-await
        return JSON.parse(v)
    } catch (err) {
        return defaultValue
    }
}
