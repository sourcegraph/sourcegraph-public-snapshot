import { BehaviorSubject, NextObserver, Observable } from 'rxjs'
import { observeStorageKey, storage } from './storage'
import { StorageItems } from './types'

/**
 * An RxJS subject that is backed by an extension storage item.
 */
export class ExtensionStorageSubject<T extends keyof StorageItems> extends Observable<StorageItems[T]>
    implements NextObserver<StorageItems[T]>, Pick<BehaviorSubject<StorageItems[T]>, 'value'> {
    constructor(private key: T, defaultValue: StorageItems[T]) {
        super(subscriber => {
            subscriber.next(this.value)
            return observeStorageKey('local', this.key).subscribe((item = defaultValue) => {
                this.value = item
                subscriber.next(item)
            })
        })
        this.value = defaultValue
    }

    public async next(value: StorageItems[T]): Promise<void> {
        await storage.local.set({ [this.key]: value })
    }

    public value: StorageItems[T]
}
