import { BehaviorSubject, NextObserver, Observable } from 'rxjs'
import storage from './storage'
import { StorageItems } from './types'

/**
 * An RxJS subject that is backed by an extension storage item.
 */
export class ExtensionStorageSubject<T extends keyof StorageItems> extends Observable<StorageItems[T]>
    implements NextObserver<StorageItems[T]>, Pick<BehaviorSubject<StorageItems[T]>, 'value'> {
    constructor(private key: T, defaultValue: StorageItems[T]) {
        super(subscriber => {
            subscriber.next(this.value)
            return storage.observeLocal(this.key).subscribe(item => {
                this.value = item
                subscriber.next(item)
            })
        })
        this.value = defaultValue
    }

    public next(value: StorageItems[T]): void {
        storage.setLocal({ [this.key]: value })
    }

    public value: StorageItems[T]
}
