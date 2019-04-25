import { BehaviorSubject, concat, NextObserver, Observable } from 'rxjs'
import { filter, map } from 'rxjs/operators'
import { fromBrowserEvent } from '../shared/util/browser'
import { StorageItems } from './types'

export { StorageItems, defaultStorageItems } from './types'

export const storage: Record<browser.storage.AreaName, browser.storage.StorageArea<StorageItems>> & {
    onChanged: browser.CallbackEventEmitter<
        (changes: browser.storage.ChangeDict<StorageItems>, areaName: browser.storage.AreaName) => void
    >
} = browser.storage

export const observeStorageKey = <K extends keyof StorageItems>(
    areaName: browser.storage.AreaName,
    key: K
): Observable<StorageItems[K] | undefined> =>
    concat(
        storage[areaName].get(key),
        fromBrowserEvent(storage.onChanged).pipe(
            filter(([, name]) => areaName === name),
            map(([changes]) => changes),
            filter(
                (changes): changes is typeof changes & { [k in K]-?: NonNullable<StorageItems[k]> } =>
                    changes.hasOwnProperty(key)
            ),
            map(changes => changes[key].newValue)
        )
    )

/**
 * An RxJS subject that is backed by an extension storage item.
 */
export class ExtensionStorageSubject<K extends keyof StorageItems> extends Observable<StorageItems[K]>
    implements NextObserver<StorageItems[K]>, Pick<BehaviorSubject<StorageItems[K]>, 'value'> {
    public value: StorageItems[K]

    /**
     * @param key The key of the storage item to watch
     * @param defaultValue The initial value that's emitted and readable through `value` and defaulted to when the storage item is removed.
     */
    constructor(private key: K, defaultValue: StorageItems[K]) {
        super(subscriber => {
            subscriber.next(this.value)
            return observeStorageKey('local', this.key).subscribe((item = defaultValue) => {
                this.value = item
                subscriber.next(item)
            })
        })
        this.value = defaultValue
    }

    public async next(value: StorageItems[K]): Promise<void> {
        await storage.local.set({ [this.key]: value })
    }
}
