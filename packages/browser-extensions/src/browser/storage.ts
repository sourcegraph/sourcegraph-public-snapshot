import { EMPTY, Observable } from 'rxjs'
import { shareReplay } from 'rxjs/operators'
import SafariStorageArea, { SafariSettingsChangeMessage, stringifyStorageArea } from './safari/StorageArea'
import { StorageChange, StorageItems } from './types'

export { StorageItems, defaultStorageItems } from './types'

type MigrateFunc = (
    items: StorageItems,
    set: (items: Partial<StorageItems>) => void,
    remove: (key: keyof StorageItems) => void
) => void

export interface Storage {
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
    addSyncMigration: (migrate: MigrateFunc) => void
    addLocalMigration: (migrate: MigrateFunc) => void
    onChanged: (listener: (changes: Partial<StorageChange>, areaName: string) => void) => void
}

const get = (area: chrome.storage.StorageArea) => (callback: (items: StorageItems) => void) => area.get(callback)
const set = (area: chrome.storage.StorageArea) => (items: Partial<StorageItems>, callback?: () => void) => {
    area.set(items, callback)
}
const getItem = (area: chrome.storage.StorageArea) => (
    key: keyof StorageItems,
    callback: (items: StorageItems) => void
) => area.get(key, callback)

const onChanged = (listener: (changes: Partial<StorageChange>, areaName: string) => void) => {
    if (chrome && chrome.storage) {
        chrome.storage.onChanged.addListener(listener)
    } else if (safari && safari.application) {
        const extension = safari.extension as SafariExtension

        extension.settings.addEventListener(
            'change',
            ({ key, newValue, oldValue }: SafariExtensionSettingsChangeEvent) => {
                const k = key as keyof StorageItems

                listener({ [k]: { newValue, oldValue } }, 'sync')
            }
        )
    } else if (safari && !safari.application) {
        const page = safari.self as SafariContentWebPage

        const handleChanges = (event: SafariExtensionMessageEvent) => {
            if (event.name === 'settings-change') {
                const { changes, areaName } = event.message as SafariSettingsChangeMessage
                const c = changes as { [key in keyof StorageItems]: chrome.storage.StorageChange }

                listener(c, areaName)
            }
        }

        page.addEventListener('message', handleChanges, false)
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

const addMigration = (
    get: Storage['getSync'],
    set: Storage['setSync'],
    remove: (keys: string | string[], callback?: (() => void) | undefined) => void
) => (migrate: MigrateFunc) => {
    get(items => {
        migrate(items as StorageItems, set, remove)
    })
}

export default ((): Storage => {
    if (window.SG_ENV === 'EXTENSION') {
        const chrome = global.chrome
        const safari = window.safari

        const syncStorageArea: chrome.storage.StorageArea =
            typeof chrome !== 'undefined'
                ? chrome.storage.sync
                : new SafariStorageArea((safari.extension as SafariExtension).settings, 'sync')

        const managedStorageArea: chrome.storage.StorageArea =
            typeof chrome !== 'undefined'
                ? chrome.storage.managed
                : new SafariStorageArea((safari.extension as SafariExtension).settings, 'managed')

        const localStorageArea: chrome.storage.StorageArea =
            typeof chrome !== 'undefined'
                ? chrome.storage.local
                : new SafariStorageArea(stringifyStorageArea(window.localStorage), 'local')

        return {
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

            addSyncMigration: addMigration(get(syncStorageArea), set(syncStorageArea), (keys, callback) =>
                syncStorageArea.remove(keys, callback)
            ),
            addLocalMigration: addMigration(get(localStorageArea), set(localStorageArea), (keys, callback) =>
                localStorageArea.remove(keys, callback)
            ),

            onChanged,
        }
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
        addSyncMigration: throwNoopErr,
        addLocalMigration: throwNoopErr,
    }
})()
