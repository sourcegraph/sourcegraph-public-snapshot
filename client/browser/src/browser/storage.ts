import { concat, from, Observable } from 'rxjs'
import { filter, map } from 'rxjs/operators'
import { fromBrowserEvent } from '../shared/util/browser'
import { StorageItems } from './types'

/**
 * Type-safe access to browser extension storage.
 *
 * `undefined` in non-browser context (in-page integration). Make sure to check `isInPage`/`isExtension` first.
 */
export const storage: Record<browser.storage.AreaName, browser.storage.StorageArea<StorageItems>> & {
    onChanged: browser.CallbackEventEmitter<
        (changes: browser.storage.ChangeDict<StorageItems>, areaName: browser.storage.AreaName) => void
    >
} = globalThis.browser && browser.storage

export const observeStorageKey = <K extends keyof StorageItems>(
    areaName: browser.storage.AreaName,
    key: K
): Observable<StorageItems[K] | undefined> =>
    concat(
        // Start with current value of the item
        from(storage[areaName].get(key)).pipe(map(items => items[key])),
        // Emit every new value from change events that affect that item
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
