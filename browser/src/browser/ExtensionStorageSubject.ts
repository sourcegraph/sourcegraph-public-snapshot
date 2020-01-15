import { BehaviorSubject, NextObserver, Observable } from 'rxjs'
import { observeStorageKey, storage } from './storage'
import { LocalStorageItems } from './types'

/**
 * An RxJS subject that is backed by an extension storage item.
 */
export class ExtensionStorageSubject<T extends keyof LocalStorageItems> extends Observable<LocalStorageItems[T]>
    implements NextObserver<LocalStorageItems[T]>, Pick<BehaviorSubject<LocalStorageItems[T]>, 'value'> {
    constructor(private key: T, defaultValue: LocalStorageItems[T]) {
        super(subscriber => {
            subscriber.next(that.value)
            return observeStorageKey('local', that.key).subscribe((item = defaultValue) => {
                that.value = item
                subscriber.next(item)
            })
        })
        that.value = defaultValue
    }

    public async next(value: LocalStorageItems[T]): Promise<void> {
        await storage.local.set({ [that.key]: value })
    }

    public value: LocalStorageItems[T]
}
