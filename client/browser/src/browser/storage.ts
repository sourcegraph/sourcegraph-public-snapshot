import { EMPTY, Observable } from 'rxjs'
import { shareReplay } from 'rxjs/operators'
import { MigratableStorageArea, noopMigration, provideMigrations } from './storage_migrations'
import { StorageChange, StorageItems } from './types'

export { StorageItems, defaultStorageItems } from './types'

interface Storage {
    getManaged: (callback: (items: StorageItems) => void) => void
    getManagedItem: (key: keyof StorageItems, callback: (items: StorageItems) => void) => void
    getSync: (callback: (items: StorageItems) => void) => void
    getSyncItem: (key: keyof StorageItems, callback: (items: StorageItems) => void) => void
    setSync: (items: Partial<StorageItems>, callback?: (() => void) | undefined) => void
    observeSync: <T extends keyof StorageItems>(key: T) => Observable<StorageItems[T]>
    getLocal: (callback: (items: StorageItems) => void) => void
    getLocalItem: (key: keyof StorageItems, callback: (items: StorageItems) => void) => void
    setLocal: (items: Partial<StorageItems>, callback?: (() => void) | undefined) => void
    observeLocal: <T extends keyof StorageItems>(key: T) => Observable<StorageItems[T]>
    setSyncMigration: MigratableStorageArea['setMigration']
    setLocalMigration: MigratableStorageArea['setMigration']
    onChanged: (listener: (changes: Partial<StorageChange>, areaName: string) => void) => void
}

const get = (area: chrome.storage.StorageArea) => (callback: (items: StorageItems) => void) =>
    area.get(items => callback(items as StorageItems))
const set = (area: chrome.storage.StorageArea) => (items: Partial<StorageItems>, callback?: () => void) => {
    area.set(items, callback)
}
const getItem = (area: chrome.storage.StorageArea) => (
    key: keyof StorageItems,
    callback: (items: StorageItems) => void
) => area.get(key, items => callback(items as StorageItems))

const onChanged = (listener: (changes: Partial<StorageChange>, areaName: string) => void) => {
    if (chrome && chrome.storage) {
        chrome.storage.onChanged.addListener(listener)
    }
}

const observe = (area: chrome.storage.StorageArea) => <T extends keyof StorageItems>(
    key: T
): Observable<StorageItems[T]> =>
    new Observable<StorageItems[T]>(observer => {
        get(area)(items => {
            const item = items[key]
            observer.next(item)
        })
        onChanged(changes => {
            const change = changes[key]
            if (change) {
                observer.next(change.newValue)
            }
        })
    }).pipe(shareReplay(1))

const noopObserve = () => EMPTY

const throwNoopErr = () => {
    throw new Error('do not call browser extension apis from an in page script')
}

export default ((): Storage => {
    if (window.SG_ENV === 'EXTENSION') {
        const chrome = global.chrome

        const syncStorageArea = provideMigrations(chrome.storage.sync)
        const localStorageArea = provideMigrations(chrome.storage.local)
        const managedStorageArea: chrome.storage.StorageArea = chrome.storage.managed

        const storage: Storage = {
            getManaged: get(managedStorageArea),
            getManagedItem: getItem(managedStorageArea),

            getSync: get(syncStorageArea),
            getSyncItem: getItem(syncStorageArea),
            setSync: set(syncStorageArea),
            observeSync: observe(syncStorageArea),

            getLocal: get(localStorageArea),
            getLocalItem: getItem(localStorageArea),
            setLocal: set(localStorageArea),
            observeLocal: observe(localStorageArea),

            setSyncMigration: syncStorageArea.setMigration,
            setLocalMigration: localStorageArea.setMigration,

            onChanged,
        }

        // Only background script should set migrations.
        if (window.EXTENSION_ENV !== 'BACKGROUND') {
            storage.setSyncMigration(noopMigration)
            storage.setLocalMigration(noopMigration)
        }

        return storage
    }

    // Running natively in the webpage(in Phabricator patch) so we don't need any storage.
    return {
        getManaged: throwNoopErr,
        getManagedItem: throwNoopErr,
        getSync: throwNoopErr,
        getSyncItem: throwNoopErr,
        setSync: throwNoopErr,
        observeSync: noopObserve,
        onChanged: throwNoopErr,
        getLocal: throwNoopErr,
        getLocalItem: throwNoopErr,
        setLocal: throwNoopErr,
        observeLocal: noopObserve,
        setSyncMigration: throwNoopErr,
        setLocalMigration: throwNoopErr,
    }
})()
