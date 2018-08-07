// We want to polyfill first.
// prettier-ignore
import '../../app/util/polyfill'

import { without } from 'lodash'

import initializeCli from '../../app/cli'
import { setServerUrls, setSourcegraphUrl } from '../../app/util/context'
import * as browserAction from '../../extension/browserAction'
import * as omnibox from '../../extension/omnibox'
import * as permissions from '../../extension/permissions'
import * as runtime from '../../extension/runtime'
import storage, { defaultStorageItems } from '../../extension/storage'
import * as tabs from '../../extension/tabs'

let customServerOrigins: string[] = []

const contentScripts = runtime.getContentScripts()

// jsContentScriptOrigins are the required URLs inside of the manifest. When checking for permissions to inject
// the content script on optional pages (inside browser.tabs.onUpdated) we need to skip manual injection of the
// script since the browser extension will automatically inject it.
const jsContentScriptOrigins: string[] = []
if (contentScripts) {
    for (const contentScript of contentScripts) {
        if (!contentScript || !contentScript.js || !contentScript.matches) {
            continue
        }
        jsContentScriptOrigins.push(...contentScript.matches)
    }
}

const configureOmnibox = (serverUrl: string) => {
    omnibox.setDefaultSuggestion({
        description: `Search code on ${serverUrl}`,
    })
}

initializeCli(omnibox)

storage.getSync(({ sourcegraphURL }) => configureOmnibox(sourcegraphURL))

storage.getManaged(items => {
    if (!items.enterpriseUrls || !items.enterpriseUrls.length) {
        setDefaultBrowserAction()
        return
    }
    const urls = items.enterpriseUrls.map(item => {
        if (item.endsWith('/')) {
            return item.substr(item.length - 1)
        }
        return item
    })
    handleManagedPermissionRequest(urls)
})

storage.onChanged((changes, areaName) => {
    if (areaName === 'managed') {
        storage.getSync(items => {
            if (changes.serverUrls && changes.serverUrls.newValue) {
                const serverUrls = [...new Set([...items.serverUrls, ...changes.serverUrls.newValue])]
                setServerUrls(serverUrls)
                if (serverUrls.length) {
                    storage.setSync({ serverUrls, sourcegraphURL: serverUrls[0] })
                }
            }
            if (changes.enterpriseUrls && changes.enterpriseUrls.newValue) {
                handleManagedPermissionRequest(changes.enterpriseUrls.newValue)
            }
        })
        return
    }

    if (changes.sourcegraphURL && changes.sourcegraphURL.newValue) {
        setSourcegraphUrl(changes.sourcegraphURL.newValue)
        configureOmnibox(changes.sourcegraphURL.newValue)
    }
    if (changes.serverUrls && changes.serverUrls.newValue) {
        setServerUrls(changes.serverUrls.newValue)
    }
})

permissions.getAll().then(permissions => {
    if (!permissions.origins) {
        customServerOrigins = []
        return
    }
    customServerOrigins = without(permissions.origins, ...jsContentScriptOrigins)
})

permissions.onAdded(permissions => {
    if (!permissions.origins) {
        return
    }
    storage.getSync(items => {
        const enterpriseUrls = items.enterpriseUrls || []
        for (const url of permissions.origins as string[]) {
            enterpriseUrls.push(url.replace('/*', ''))
        }
        storage.setSync({ enterpriseUrls })
    })
    const origins = without(permissions.origins, ...jsContentScriptOrigins)
    customServerOrigins.push(...origins)
})

permissions.onRemoved(permissions => {
    if (!permissions.origins) {
        return
    }
    customServerOrigins = without(customServerOrigins, ...permissions.origins)
    storage.getSync(items => {
        const enterpriseUrls = items.enterpriseUrls || []
        const urlsToRemove: string[] = []
        for (const url of permissions.origins as string[]) {
            urlsToRemove.push(url.replace('/*', ''))
        }
        storage.setSync({ enterpriseUrls: without(enterpriseUrls, ...urlsToRemove) })
    })
})

storage.addSyncMigration((items, set, remove) => {
    if (items.phabricatorURL) {
        remove('phabricatorURL')

        const newItems: {
            enterpriseUrls?: string[]
        } = {}

        if (items.enterpriseUrls && !items.enterpriseUrls.find(u => u === items.phabricatorURL)) {
            newItems.enterpriseUrls = items.enterpriseUrls.concat(items.phabricatorURL)
        } else if (!items.enterpriseUrls) {
            newItems.enterpriseUrls = [items.phabricatorURL]
        }

        set(newItems)
    }

    if (!items.repoLocations) {
        set({ repoLocations: {} })
    }

    if (items.openFileOnSourcegraph === undefined) {
        set({ openFileOnSourcegraph: true })
    }

    if (items.featureFlags.newTooltips) {
        set({ featureFlags: { ...items.featureFlags, newTooltips: true } })
    }

    if (!items.inlineSymbolSearchEnabled) {
        set({ inlineSymbolSearchEnabled: true })
    }
})

tabs.onUpdated((tabId, changeInfo, tab) => {
    if (changeInfo.status === 'complete') {
        for (const origin of customServerOrigins) {
            if (origin !== '<all_urls>' && (!tab.url || !tab.url.startsWith(origin.replace('/*', '')))) {
                continue
            }
            tabs.executeScript(tabId, { file: 'js/inject.bundle.js', runAt: 'document_end', origin })
        }
    }
})

runtime.onMessage((message, _, cb) => {
    switch (message.type) {
        case 'setIdentity':
            storage.setLocal({ identity: message.payload.identity })
            return

        case 'getIdentity':
            storage.getLocalItem('identity', obj => {
                const { identity } = obj

                cb(identity)
            })
            return true

        case 'setEnterpriseUrl':
            requestPermissionsForEnterpriseUrls([message.payload], cb)
            return true

        case 'setSourcegraphUrl':
            requestPermissionsForSourcegraphUrl(message.payload)
            return true

        case 'removeEnterpriseUrl':
            removeEnterpriseUrl(message.payload, cb)
            return true

        // We should only need to do this on safari
        case 'insertCSS':
            const details = message.payload as { file: string; origin: string }
            storage.getSyncItem('serverUrls', ({ serverUrls }) =>
                tabs.insertCSS(0, {
                    ...details,
                    whitelist: details.origin ? [details.origin] : [],
                    blacklist: serverUrls || [],
                })
            )
            return

        case 'setBadgeText':
            browserAction.setBadgeText({ text: message.payload })
    }

    return
})

function requestPermissionsForEnterpriseUrls(urls: string[], cb: (res?: any) => void): void {
    storage.getSync(items => {
        const enterpriseUrls = items.enterpriseUrls || []
        storage.setSync(
            {
                enterpriseUrls: [...new Set([...enterpriseUrls, ...urls])],
            },
            cb
        )
    })
}

function requestPermissionsForSourcegraphUrl(url: string): void {
    permissions.request([url]).then(granted => {
        if (granted) {
            storage.setSync({ sourcegraphURL: url })
        }
    })
}

function removeEnterpriseUrl(url: string, cb: (res?: any) => void): void {
    permissions.remove(url)

    storage.getSyncItem('enterpriseUrls', ({ enterpriseUrls }) => {
        storage.setSync({ enterpriseUrls: without(enterpriseUrls, url) }, cb)
    })
}

runtime.setUninstallURL('https://about.sourcegraph.com/uninstall/')

runtime.onInstalled(() => {
    setDefaultBrowserAction()
    storage.getSync(items => {
        if (window.safari) {
            // Safari settings returns null for getters of values that don't exist so
            // values get returned with null and we can't override then with defaults like below.
            storage.setSync({
                ...defaultStorageItems,
            })
        } else {
            // Enterprise deployments of Sourcegraph are passed a configuration file.
            storage.getManaged(managedItems => {
                storage.setSync(
                    {
                        ...defaultStorageItems,
                        ...items,
                        ...managedItems,
                    },
                    () => {
                        if (managedItems && managedItems.enterpriseUrls && managedItems.enterpriseUrls.length) {
                            handleManagedPermissionRequest(managedItems.enterpriseUrls)
                        } else {
                            setDefaultBrowserAction()
                        }
                    }
                )
            })
        }
    })
})

function handleManagedPermissionRequest(managedUrls: string[]): void {
    setDefaultBrowserAction()
    if (managedUrls.length === 0) {
        return
    }
    permissions.getAll().then(perms => {
        const origins = perms.origins || []
        if (managedUrls.every(val => origins.indexOf(`${val}/*`) >= 0)) {
            setDefaultBrowserAction()
            return
        }
        browserAction.setPopup({ popup: '' })
        browserAction.setBadgeText({ text: '1' })
        browserAction.onClicked(() => {
            permissions.request(managedUrls).then(added => {
                if (!added) {
                    return
                }
                setDefaultBrowserAction()
            })
        })
    })
}

function setDefaultBrowserAction(): void {
    browserAction.setBadgeText({ text: '' })
    browserAction.setPopup({ popup: 'options.html?popup=true' })
}
