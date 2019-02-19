// We want to polyfill first.
// prettier-ignore
import '../../config/polyfill'

import { without } from 'lodash'
import { fromEventPattern, noop, Observable } from 'rxjs'
import { bufferCount, filter, groupBy, mergeMap, reduce } from 'rxjs/operators'
import DPT from 'webext-domain-permission-toggle'
import { EndpointPair } from '../../../../../shared/src/platform/context'
import * as browserAction from '../../browser/browserAction'
import * as omnibox from '../../browser/omnibox'
import * as permissions from '../../browser/permissions'
import * as runtime from '../../browser/runtime'
import storage, { defaultStorageItems } from '../../browser/storage'
import * as tabs from '../../browser/tabs'
import { featureFlagDefaults, FeatureFlags } from '../../browser/types'
import initializeCli from '../../libs/cli'
import { initSentry } from '../../libs/sentry'
import { createBlobURLForBundle, createExtensionHostWorker } from '../../platform/worker'
import { requestGraphQL } from '../../shared/backend/graphql'
import { resolveClientConfiguration } from '../../shared/backend/server'
import { DEFAULT_SOURCEGRAPH_URL, setSourcegraphUrl } from '../../shared/util/context'
import { assertEnv } from '../envAssertion'

assertEnv('BACKGROUND')

initSentry('background')

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
        storage.getSync(() => {
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

storage.setSyncMigration(items => {
    const newItems = { ...defaultStorageItems, ...items }

    let featureFlags: FeatureFlags = {
        ...featureFlagDefaults,
        ...(newItems.featureFlags || {}),
    }

    const keysToRemove: string[] = []

    // Ensure all feature flags are in storage.
    for (const key of Object.keys(featureFlagDefaults)) {
        if (typeof featureFlags[key] === 'undefined') {
            keysToRemove.push(key)
            featureFlags = {
                ...featureFlagDefaults,
                ...items.featureFlags,
                [key]: featureFlagDefaults[key],
            }
        }
    }

    newItems.featureFlags = featureFlags

    newItems.featureFlags.renderMermaidGraphsEnabled = true

    // TODO: Remove this block after a few releases
    const clientSettings = JSON.parse(items.clientSettings || '{}')
    if (clientSettings['codecov.endpoints'] || typeof clientSettings['codecov.showCoverage'] !== 'undefined') {
        if (typeof clientSettings.extensions === 'undefined') {
            clientSettings.extensions = clientSettings.extensions || {}
        }
        clientSettings.extensions['souercegraph/codecov'] = true
        newItems.clientSettings = JSON.stringify(clientSettings, null, 4)
    }

    return { newItems, keysToRemove }
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

                // TODO: remove "!"" added after typescript upgrade
                cb!(identity)
            })
            return true

        case 'setEnterpriseUrl':
            // TODO: remove "!"" added after typescript upgrade
            requestPermissionsForEnterpriseUrls([message.payload], cb!)
            return true

        case 'setSourcegraphUrl':
            requestPermissionsForSourcegraphUrl(message.payload)
            return true

        case 'removeEnterpriseUrl':
            // TODO: remove "!"" added after typescript upgrade
            removeEnterpriseUrl(message.payload, cb!)
            return true

        case 'insertCSS':
            const details = message.payload as { file: string; origin: string }
            storage.getSyncItem('sourcegraphURL', ({ sourcegraphURL }) =>
                tabs.insertCSS(0, {
                    ...details,
                    whitelist: details.origin ? [details.origin] : [],
                    blacklist: [sourcegraphURL],
                })
            )
            return

        case 'setBadgeText':
            browserAction.setBadgeText({ text: message.payload })
            return
        case 'openOptionsPage':
            runtime.openOptionsPage()
            return true
        case 'createBlobURL':
            createBlobURLForBundle(message.payload! as string)
                .then(url => {
                    if (cb) {
                        cb(url)
                    }
                })
                .catch(err => {
                    throw new Error(`Unable to create blob url for bundle ${message.payload} error: ${err}`)
                })
            return true
        case 'requestGraphQL':
            requestGraphQL(message.payload)
                .toPromise()
                .then(result => cb && cb({ result }))
                .catch(err => cb && cb({ err }))
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
    browserAction.setPopup({ popup: 'options.html?popup=true' }).catch(err => console.error(err))
}

browserAction.onClicked(noop)
setDefaultBrowserAction()

// Add "Enable Sourcegraph on this domain" context menu item
DPT.addContextMenu()

const ENDPOINT_KIND_REGEX = /^(proxy|expose)-/

/**
 * A stream of EndpointPair created from Port objects emitted by chrome.runtime.onConnect.
 *
 * Each EndpointPair represents a connection with an instance of the content script.
 */
const endpointPairs: Observable<{ proxy: chrome.runtime.Port; expose: chrome.runtime.Port }> = fromEventPattern(
    handler => chrome.runtime.onConnect.addListener(handler),
    handler => chrome.runtime.onConnect.removeListener(handler)
).pipe(
    groupBy(
        (port: any): string => (port.name || 'other').replace(ENDPOINT_KIND_REGEX, ''),
        (port: any) => {
            const endpointKind = (port.name || '').match(ENDPOINT_KIND_REGEX)
            if (endpointKind) {
                return {
                    [endpointKind[1]]: port,
                }
            }
            return {}
        },
        group => group.pipe(bufferCount(2))
    ),
    filter(group => group.key !== 'other'),
    mergeMap(group => group.pipe(reduce((acc, cur) => ({ ...acc, ...cur }), {})))
) as any

// Create one extension host worker per endpoint pair
endpointPairs.subscribe(({ proxy, expose }) => {
    const worker = createExtensionHostWorker()
    const clientAPIChannel = new MessageChannel()
    const extensionHostAPIChannel = new MessageChannel()
    const workerEndpoints: EndpointPair = {
        proxy: clientAPIChannel.port2,
        expose: extensionHostAPIChannel.port2,
    }
    worker.postMessage({ endpoints: workerEndpoints, wrapEndpoints: true }, Object.values(workerEndpoints))
    // Connect proxy client endpoint
    extensionHostAPIChannel.port1.start()
    proxy.onMessage.addListener(message => {
        extensionHostAPIChannel.port1.postMessage(message)
    })
    extensionHostAPIChannel.port1.addEventListener('message', ({ data }) => {
        proxy.postMessage({ data })
    })
    // Connect expose client endpoint
    clientAPIChannel.port1.start()
    expose.onMessage.addListener(message => {
        clientAPIChannel.port1.postMessage(message)
    })
    clientAPIChannel.port1.addEventListener('message', ({ data }) => {
        expose.postMessage({ data })
    })
    // Kill worker when either port disconnects
    proxy.onDisconnect.addListener(() => worker.terminate())
    expose.onDisconnect.addListener(() => worker.terminate())
})
