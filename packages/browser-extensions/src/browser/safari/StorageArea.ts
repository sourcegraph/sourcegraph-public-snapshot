import { defaultStorageItems, StorageItems } from '../types'

interface ISafariStorageArea {
    getItem: (key: string) => any | null
    setItem: (key: string, value: any) => void
    removeItem: (key: string) => void
    clear: () => void
}

export interface SafariSettingsChangeMessage {
    changes: { [key: string]: chrome.storage.StorageChange }
    areaName: string
}

interface SetItemPayload {
    key: string
    value: any
}

interface SettingsHandlerMessage {
    type: keyof ISafariStorageArea | 'get' | 'init'
    payload?: string | SetItemPayload
}

const safari = window.safari

class ContentStorageArea implements ISafariStorageArea {
    private page = safari.self as SafariContentWebPage
    private items: StorageItems = defaultStorageItems

    public isReady = false
    private pendingGetCallbacks: ((items: any) => void)[] = []

    constructor(name: string) {
        if (safari.application) {
            return
        }

        this.page.addEventListener('message', this.init, false)

        this.page.tab.dispatchMessage('settings', { type: 'init' })

        this.page.addEventListener('message', this.handleChanges, false)
    }

    private handleChanges = (event: SafariExtensionMessageEvent) => {
        if (event.name === 'settings' && event.message.type === 'change') {
            const { changes } = event.message as SafariSettingsChangeMessage

            for (const key of Object.keys(changes)) {
                this.items[key] = changes[key].newValue
            }
        }
    }

    private init = (event: SafariExtensionMessageEvent) => {
        if (event.name === 'settings' && event.message.type === 'init') {
            const items = event.message.payload as { [key in keyof StorageItems]: any }

            // TODO(isaac): Figure out why sometimes this gets fired when we don't have any values
            //
            // We should always have a sourcegraphURL
            if (items.sourcegraphURL) {
                this.items = items
                document.dispatchEvent(new CustomEvent<{}>('sourcegraph:storage-init'))

                for (const cb of this.pendingGetCallbacks) {
                    cb(this.items)
                }

                this.isReady = true
                this.pendingGetCallbacks = []
            }
        }
    }

    public addPendingGetCallback(cb: (items: any) => void): void {
        if (this.isReady) {
            cb(this.items)
            return
        }

        this.pendingGetCallbacks.push(cb)
    }

    public getItem(key: string): any | null {
        return this.items[key]
    }

    public setItem(key: string, value: any): void {
        this.page.tab.dispatchMessage('settings', { type: 'setItem', payload: { key, value } })
    }

    public removeItem(key: string): void {
        this.page.tab.dispatchMessage('settings', { type: 'removeItem', payload: key })
    }

    public clear(): void {
        this.page.tab.dispatchMessage('settings', { type: 'clear' })
    }
}

export default class SafariStorageArea implements chrome.storage.StorageArea {
    private area: ISafariStorageArea | null = null
    private keys: string[] = []
    private isBackground = !!safari.application

    constructor(area: ISafariStorageArea, areaName: string) {
        if (area) {
            this.area = area
        } else if (!this.isBackground && areaName === 'sync') {
            this.area = new ContentStorageArea(areaName)
        } else if (!this.isBackground && areaName === 'local') {
            this.area = localStorage
        }

        this.keys = Object.keys(defaultStorageItems)

        this.initBackground()
    }

    private initBackground(): void {
        if (!safari.application) {
            return
        }

        safari.application.addEventListener('message', this.handleStorageMessages, false)
    }

    private handleStorageMessages = (event: SafariExtensionMessageEvent) => {
        if (event.name !== 'settings' || !this.area) {
            return
        }

        const message = event.message as SettingsHandlerMessage
        if (message.type === 'setItem') {
            const { key, value } = message.payload as SetItemPayload

            this.area.setItem(key, value)
        } else if (message.type === 'removeItem') {
            const key = message.payload as string

            this.area.removeItem(key)
        } else if (message.type === 'clear') {
            this.area.clear()
        } else if (message.type === 'get' || message.type === 'init') {
            const tab = event.target as SafariBrowserTab

            this.get(items => tab.page.dispatchMessage('settings', { type: message.type, payload: items }))
        }
    }

    public getBytesInUse(
        keysOrCallback: string | string[] | ((bytesInUse: number) => void) | null,
        callback?: (bytesInUse: number) => void
    ): void {
        throw new Error('SafariStorageArea.prototype.getBytesInUse not implemented')
    }

    public clear = (callback?: () => void) => {
        if (this.area) {
            this.area.clear()
        }
    }

    public set = (items: Partial<StorageItems>, callback?: () => void) => {
        if (this.area) {
            for (const key of Object.keys(items)) {
                this.area.setItem(key, items[key])
            }

            if (callback) {
                if (this.isBackground) {
                    callback()
                } else {
                    console.log('SAFARI impl storage.set callback handler')
                }
            }
        }
    }

    public remove = (keys: string | string[], callback?: () => void) => {
        if (this.area) {
            const arr: string[] = Array.isArray(keys) ? keys : [keys]

            for (const key of arr) {
                this.area.removeItem(key)
            }
        }
    }

    public get = (
        keyOrCallback: keyof StorageItems | ((items: Partial<StorageItems>) => void),
        callback?: (items: Partial<StorageItems>) => void
    ) => {
        if (!this.area) {
            return
        }

        if (!this.isBackground) {
            const area = this.area as ContentStorageArea

            if (!area.isReady) {
                area.addPendingGetCallback(typeof keyOrCallback === 'string' ? callback! : keyOrCallback)
                return
            }
        }

        if (typeof keyOrCallback === 'string' && callback) {
            const key = keyOrCallback as keyof StorageItems
            callback({
                [key]: this.area.getItem(key),
            })
            return
        }

        const items = {}

        for (const key of this.keys) {
            items[key] = this.area.getItem(key)
        }

        const cb = keyOrCallback as (items: { [key: string]: any }) => void
        cb(items)
    }
}

export const stringifyStorageArea = (area: ISafariStorageArea): ISafariStorageArea => ({
    ...area,
    getItem: (key: string) => JSON.parse(area.getItem(key)),
    setItem: (key: string, value: any) => area.setItem(key, JSON.stringify(value)),
})
