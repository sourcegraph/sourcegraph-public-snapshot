// We want to polyfill first.
// prettier-ignore
import '../../config/polyfill'

import { without } from 'lodash'
import { ajax } from 'rxjs/ajax'
import { InitData } from 'sourcegraph/module/extension/extensionHost'
import DPT from 'webext-domain-permission-toggle'
import ExtensionHostWorker from 'worker-loader?inline!./extensionHost.worker'
import * as browserAction from '../../browser/browserAction'
import * as omnibox from '../../browser/omnibox'
import * as permissions from '../../browser/permissions'
import * as runtime from '../../browser/runtime'
import storage, { defaultStorageItems } from '../../browser/storage'
import * as tabs from '../../browser/tabs'
import initializeCli from '../../libs/cli'
import { resolveClientConfiguration } from '../../shared/backend/server'
import { ExtensionConnectionInfo, onFirstMessage } from '../../shared/messaging'
import { DEFAULT_SOURCEGRAPH_URL, setSourcegraphUrl, sourcegraphUrl } from '../../shared/util/context'
import { assertEnv } from '../envAssertion'

assertEnv('BACKGROUND')

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

storage.getSync(({ sourcegraphURL }) => {
    // If no sourcegraphURL is set ensure we default back to https://sourcegraph.com.
    if (!sourcegraphURL) {
        storage.setSync({ sourcegraphURL: DEFAULT_SOURCEGRAPH_URL })
        setSourcegraphUrl(DEFAULT_SOURCEGRAPH_URL)
    }

    resolveClientConfiguration().subscribe(
        config => {
            // ClientConfiguration is the new storage option.
            // Request permissions for the urls.
            storage.setSync({
                clientConfiguration: {
                    parentSourcegraph: {
                        url: config.parentSourcegraph.url,
                    },
                    contentScriptUrls: config.contentScriptUrls,
                },
            })
        },
        () => {
            /* noop */
        }
    )
    configureOmnibox(sourcegraphURL)
})

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
            if (changes.enterpriseUrls && changes.enterpriseUrls.newValue) {
                handleManagedPermissionRequest(changes.enterpriseUrls.newValue)
            }
        })
        return
    }

    if (changes.sourcegraphURL && changes.sourcegraphURL.newValue) {
        setSourcegraphUrl(changes.sourcegraphURL.newValue)
        resolveClientConfiguration().subscribe(
            config => {
                // ClientConfiguration is the new storage option.
                // Request permissions for the urls.
                storage.setSync({
                    clientConfiguration: {
                        parentSourcegraph: {
                            url: config.parentSourcegraph.url,
                        },
                        contentScriptUrls: config.contentScriptUrls,
                    },
                })
            },
            () => {
                /* noop */
            }
        )
        configureOmnibox(changes.sourcegraphURL.newValue)
    }
})

permissions
    .getAll()
    .then(permissions => {
        if (!permissions.origins) {
            customServerOrigins = []
            return
        }
        customServerOrigins = without(permissions.origins, ...jsContentScriptOrigins)
    })
    .catch(err => console.error('could not get permissions:', err))

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
    if (!items.accessTokens) {
        set({ accessTokens: {} })
    }

    if (items.accessTokens) {
        const accessTokens = {}

        for (const url of Object.keys(items.accessTokens)) {
            const token = items.accessTokens[url]
            if (typeof token !== 'string' && token.id && token.token) {
                accessTokens[url] = token
            }
        }

        set({ accessTokens })
    }

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

    if (items.featureFlags && !items.featureFlags.newInject) {
        set({ featureFlags: { ...items.featureFlags, newInject: true } })
    }

    if (!items.inlineSymbolSearchEnabled) {
        set({ inlineSymbolSearchEnabled: true })
    }

    if (items.serverUrls) {
        if (items.sourcegraphURL) {
            if (items.sourcegraphURL === DEFAULT_SOURCEGRAPH_URL) {
                const urls = without(items.serverUrls, DEFAULT_SOURCEGRAPH_URL)
                if (urls.length) {
                    set({ sourcegraphURL: urls[0], serverUrls: [urls[0]] })
                }
            } else {
                set({ serverUrls: [items.sourcegraphURL] })
            }
        }
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
            return
        case 'openOptionsPage':
            runtime.openOptionsPage()
            return true
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
    permissions
        .request([url])
        .then(granted => {
            if (granted) {
                storage.setSync({ sourcegraphURL: url })
            }
        })
        .catch(err => console.error('Permissions request denied', err))
}

function removeEnterpriseUrl(url: string, cb: (res?: any) => void): void {
    permissions.remove(url).catch(err => console.error('could not remove permission', err))

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
    permissions
        .getAll()
        .then(perms => {
            const origins = perms.origins || []
            if (managedUrls.every(val => origins.indexOf(`${val}/*`) >= 0)) {
                setDefaultBrowserAction()
                return
            }
            browserAction.onClicked(() => {
                runtime.openOptionsPage()
            })
        })
        .catch(err => console.error('could not get all permissions', err))
}

function setDefaultBrowserAction(): void {
    browserAction.setBadgeText({ text: '' })
}

browserAction.onClicked(() => {
    runtime.openOptionsPage()
})

/**
 * Fetches JavaScript from a URL and runs it in a web worker.
 */
function spawnWebWorkerFromURL(url: string): Promise<Worker> {
    return ajax({
        url,
        crossDomain: true,
        responseType: 'blob',
    })
        .toPromise()
        .then(response => {
            const blobURL = window.URL.createObjectURL(response.response)
            try {
                const worker = new ExtensionHostWorker()
                const initData: InitData = {
                    bundleURL: blobURL,
                    sourcegraphURL: sourcegraphUrl,
                    clientApplication: 'other',
                }
                worker.postMessage(initData)
                return worker
            } catch (err) {
                console.error(err)
            }
            throw new Error('failed to initialize extension host')
        })
}

/**
 * Connects a port and worker by forwarding messages from one to the other and
 * vice versa.
 */
const connectPortAndWorker = (port: chrome.runtime.Port, worker: Worker) => {
    worker.addEventListener('message', m => {
        port.postMessage(m.data)
    })
    port.onMessage.addListener(m => {
        worker.postMessage(m)
    })
    port.onDisconnect.addListener(() => worker.terminate())
}

/**
 * Either creates a web worker or connects to a WebSocket based on the given
 * platform, then connects the given port to it.
 */
const spawnAndConnect = ({
    connectionInfo,
    port,
}: {
    connectionInfo: ExtensionConnectionInfo
    port: chrome.runtime.Port
}): Promise<void> =>
    spawnWebWorkerFromURL(connectionInfo.jsBundleURL).then(worker => {
        connectPortAndWorker(port, worker)
    })

// This is the bridge between content scripts (that want to connect to Sourcegraph extensions) and the background
// script (that spawns JS bundles or connects to WebSocket endpoints).:
chrome.runtime.onConnect.addListener(port => {
    // When a content script wants to create a connection to a Sourcegraph extension, it first connects to the
    // background script on a random port and sends a message containing the platform information for that
    // Sourcegraph extension (e.g. a JS bundle at localhost:1234/index.js).
    onFirstMessage(port, (connectionInfo: ExtensionConnectionInfo) => {
        // The background script receives the message and attempts to spawn the
        // extension:
        spawnAndConnect({ connectionInfo, port }).then(
            // If spawning succeeds, the background script sends {} (so the content script knows it succeeded) and
            // the port communicates using the internal Sourcegraph extension RPC API after that.
            () => {
                // Success is represented by the absence of an error
                port.postMessage({})
            },
            // If spawning fails, the background script sends { error } (so the content script knows it failed) and
            // the port is immediately disconnected. There is always a 1-1 correspondence between ports and content
            // scripts, so this won't disrupt any other connections.
            error => {
                port.postMessage({ error })
                port.disconnect()
            }
        )
    })
})

// Add "Enable Sourcegraph on this domain" context menu item
DPT.addContextMenu()
