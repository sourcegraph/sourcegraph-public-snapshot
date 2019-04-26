// We want to polyfill first.
import '../../config/polyfill'

import { Endpoint } from '@sourcegraph/comlink'
import { without } from 'lodash'
import { noop, Observable } from 'rxjs'
import { bufferCount, filter, groupBy, map, mergeMap } from 'rxjs/operators'
import * as domainPermissionToggle from 'webext-domain-permission-toggle'
import { createExtensionHostWorker } from '../../../../../shared/src/api/extension/worker'
import { IGraphQLResponseRoot } from '../../../../../shared/src/graphql/schema'
import { defaultStorageItems, storage } from '../../browser/storage'
import { BackgroundMessageHandlers, StorageItems } from '../../browser/types'
import { initializeOmniboxInterface } from '../../libs/cli'
import { initSentry } from '../../libs/sentry'
import { createBlobURLForBundle } from '../../platform/worker'
import { GraphQLRequestArgs, requestGraphQL } from '../../shared/backend/graphql'
import { resolveClientConfiguration } from '../../shared/backend/server'
import { fromBrowserEvent } from '../../shared/util/browser'
import { DEFAULT_SOURCEGRAPH_URL, getPlatformName, setSourcegraphUrl } from '../../shared/util/context'
import { assertEnv } from '../envAssertion'

assertEnv('BACKGROUND')

initSentry('background')

let customServerOrigins: string[] = []

const contentScripts = browser.runtime.getManifest().content_scripts

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

const configureOmnibox = (serverUrl: string): void => {
    browser.omnibox.setDefaultSuggestion({
        description: `Search code on ${serverUrl}`,
    })
}

initializeOmniboxInterface()

async function main(): Promise<void> {
    const { sourcegraphURL } = await browser.storage.sync.get()
    // If no sourcegraphURL is set ensure we default back to https://sourcegraph.com.
    if (!sourcegraphURL) {
        await browser.storage.sync.set({ sourcegraphURL: DEFAULT_SOURCEGRAPH_URL })
        setSourcegraphUrl(DEFAULT_SOURCEGRAPH_URL)
    }

    async function syncClientConfiguration(): Promise<void> {
        const config = await resolveClientConfiguration().toPromise()
        // ClientConfiguration is the new storage option.
        // Request permissions for the urls.
        await browser.storage.sync.set({
            clientConfiguration: {
                parentSourcegraph: {
                    url: config.parentSourcegraph.url,
                },
                contentScriptUrls: config.contentScriptUrls,
            },
        })
    }

    configureOmnibox(sourcegraphURL)

    // Sync managed enterprise URLs
    // TODO why sync vs merging values?
    // Managed storage is currently only supported for Google Chrome (GSuite Admin)
    // We don't have a managed storage manifest for Firefox, so storage.managed.get() throws on Firefox.
    if (getPlatformName() === 'chrome-extension') {
        const items = await storage.managed.get()
        if (items.enterpriseUrls && items.enterpriseUrls.length > 1) {
            setDefaultBrowserAction()
            const urls = items.enterpriseUrls.map(item => item.replace(/\/$/, ''))
            await handleManagedPermissionRequest(urls)
        }
    }

    browser.storage.onChanged.addListener(async (changes: browser.storage.ChangeDict<StorageItems>, areaName) => {
        if (areaName === 'managed') {
            if (changes.enterpriseUrls && changes.enterpriseUrls.newValue) {
                await handleManagedPermissionRequest(changes.enterpriseUrls.newValue)
            }
            return
        }

        if (changes.sourcegraphURL && changes.sourcegraphURL.newValue) {
            setSourcegraphUrl(changes.sourcegraphURL.newValue)
            await syncClientConfiguration()
            configureOmnibox(changes.sourcegraphURL.newValue)
        }
    })

    const permissions = await browser.permissions.getAll()
    if (!permissions.origins) {
        customServerOrigins = []
        return
    }
    customServerOrigins = without(permissions.origins, ...jsContentScriptOrigins)

    // Not supported in Firefox
    // https://developer.mozilla.org/en-US/docs/Mozilla/Add-ons/WebExtensions/API/permissions/onAdded#Browser_compatibility
    if (browser.permissions.onAdded) {
        browser.permissions.onAdded.addListener(async permissions => {
            if (!permissions.origins) {
                return
            }
            const items = await storage.sync.get()
            const enterpriseUrls = items.enterpriseUrls || []
            for (const url of permissions.origins) {
                enterpriseUrls.push(url.replace('/*', ''))
            }
            await browser.storage.sync.set({ enterpriseUrls })

            const origins = without(permissions.origins, ...jsContentScriptOrigins)
            customServerOrigins.push(...origins)
        })
    }
    if (browser.permissions.onRemoved) {
        browser.permissions.onRemoved.addListener(async permissions => {
            if (!permissions.origins) {
                return
            }
            customServerOrigins = without(customServerOrigins, ...permissions.origins)
            const items = await storage.sync.get()
            const enterpriseUrls = items.enterpriseUrls || []
            const urlsToRemove: string[] = []
            for (const url of permissions.origins) {
                urlsToRemove.push(url.replace('/*', ''))
            }
            await storage.sync.set({
                enterpriseUrls: without(enterpriseUrls, ...urlsToRemove),
            })
        })
    }

    // Inject content script whenever a new tab was opened with a URL that we have premissions for
    browser.tabs.onUpdated.addListener(async (tabId, changeInfo, tab) => {
        if (
            changeInfo.status === 'complete' &&
            customServerOrigins.some(
                origin => origin === '<all_urls>' || (!!tab.url && tab.url.startsWith(origin.replace('/*', '')))
            )
        ) {
            await browser.tabs.executeScript(tabId, { file: 'js/inject.bundle.js', runAt: 'document_end' })
        }
    })

    const handlers: BackgroundMessageHandlers = {
        async setIdentity({ identity }: { identity: string }): Promise<void> {
            await browser.storage.local.set({ identity })
        },

        async getIdentity(): Promise<string | undefined> {
            const { identity } = await (browser.storage.local as browser.storage.StorageArea<StorageItems>).get(
                'identity'
            )
            return identity
        },

        async setEnterpriseUrl(url: string): Promise<void> {
            await requestPermissionsForEnterpriseUrls([url])
        },

        async setSourcegraphUrl(url: string): Promise<void> {
            await requestPermissionsForSourcegraphUrl(url)
        },

        async removeEnterpriseUrl(url: string): Promise<void> {
            await removeEnterpriseUrl(url)
        },

        async insertCSS(details: { file: string; origin: string }): Promise<void> {
            await browser.tabs.insertCSS(0, { ...details })
        },
        setBadgeText(text: string): void {
            browser.browserAction.setBadgeText({ text })
        },

        async openOptionsPage(): Promise<void> {
            await browser.runtime.openOptionsPage()
        },

        async createBlobURL(bundleUrl: string): Promise<string> {
            return await createBlobURLForBundle(bundleUrl)
        },

        async requestGraphQL(params: GraphQLRequestArgs): Promise<IGraphQLResponseRoot> {
            return await requestGraphQL(params).toPromise()
        },
    }

    // Handle calls from other scripts
    browser.runtime.onMessage.addListener(async message => {
        const handler = message.type as keyof typeof handlers
        return await handlers[handler](message.payload)
    })

    async function requestPermissionsForEnterpriseUrls(urls: string[]): Promise<void> {
        const items = await storage.sync.get()
        const enterpriseUrls = items.enterpriseUrls || []
        // Add requested URLs, without duplicating
        await browser.storage.sync.set({
            enterpriseUrls: [...new Set([...enterpriseUrls, ...urls])],
        })
    }

    async function requestPermissionsForSourcegraphUrl(url: string): Promise<void> {
        const granted = await browser.permissions.request({ origins: [url + '/*'] })
        if (granted) {
            await storage.sync.set({ sourcegraphURL: url })
        }
    }

    async function removeEnterpriseUrl(url: string): Promise<void> {
        try {
            await browser.permissions.remove({ origins: [url + '/*'] })
            // tslint:disable-next-line:no-unnecessary-type-assertion False positive
            const { enterpriseUrls } = await storage.sync.get('enterpriseUrls')
            await storage.sync.set({ enterpriseUrls: without(enterpriseUrls, url) })
        } catch (err) {
            console.error('Could not remove permission', err)
        }
    }

    await browser.runtime.setUninstallURL('https://about.sourcegraph.com/uninstall/')

    browser.runtime.onInstalled.addListener(async () => {
        setDefaultBrowserAction()
        const items = await storage.sync.get()
        // Enterprise deployments of Sourcegraph are passed a configuration file.
        const managedItems = await storage.managed.get()
        await storage.sync.set({
            ...defaultStorageItems,
            ...items,
            ...managedItems,
        })
        if (managedItems && managedItems.enterpriseUrls && managedItems.enterpriseUrls.length) {
            await handleManagedPermissionRequest(managedItems.enterpriseUrls)
        } else {
            setDefaultBrowserAction()
        }
    })

    async function handleManagedPermissionRequest(managedUrls: string[]): Promise<void> {
        setDefaultBrowserAction()
        if (managedUrls.length === 0) {
            return
        }
        const perms = await browser.permissions.getAll()
        const origins = perms.origins || []
        if (managedUrls.every(val => origins.indexOf(`${val}/*`) >= 0)) {
            setDefaultBrowserAction()
            return
        }
        browser.browserAction.onClicked.addListener(async () => {
            await browser.runtime.openOptionsPage()
        })
    }

    function setDefaultBrowserAction(): void {
        browser.browserAction.setBadgeText({ text: '' })
        browser.browserAction.setPopup({ popup: 'options.html?popup=true' })
    }

    browser.browserAction.onClicked.addListener(noop)
    setDefaultBrowserAction()

    // Add "Enable Sourcegraph on this domain" context menu item
    domainPermissionToggle.addContextMenu()

    const ENDPOINT_KIND_REGEX = /^(proxy|expose)-/

    const portKind = (port: browser.runtime.Port): string | null => {
        const match = port.name.match(ENDPOINT_KIND_REGEX)
        return match && match[1]
    }

    /**
     * A stream of EndpointPair created from Port objects emitted by browser.runtime.onConnect.
     *
     * On initialization, the content script creates a pair of browser.runtime.Port objects
     * using browser.runtime.connect(). The two ports are named 'proxy-{uuid}' and 'expose-{uuid}',
     * and wrapped using {@link endpointFromPort} to behave like comlink endpoints on the content script side.
     *
     * This listens to events on browser.runtime.onConnect, pairs emitted ports using their naming pattern,
     * and emits pairs. Each pair of ports represents a connection with an instance of the content script.
     */
    const endpointPairs: Observable<{ proxy: browser.runtime.Port; expose: browser.runtime.Port }> = fromBrowserEvent(
        browser.runtime.onConnect
    ).pipe(
        map(([port]) => port),
        groupBy(
            port => (port.name || 'other').replace(ENDPOINT_KIND_REGEX, ''),
            port => port,
            group => group.pipe(bufferCount(2))
        ),
        filter(group => group.key !== 'other'),
        mergeMap(group =>
            group.pipe(
                bufferCount(2),
                map(ports => {
                    const proxyPort = ports.find(port => portKind(port) === 'proxy')
                    if (!proxyPort) {
                        throw new Error('No proxy port')
                    }
                    const exposePort = ports.find(port => portKind(port) === 'expose')
                    if (!exposePort) {
                        throw new Error('No expose port')
                    }
                    return {
                        proxy: proxyPort,
                        expose: exposePort,
                    }
                })
            )
        )
    )

    /**
     * Extension Host Connection
     *
     * When an Port pair is emitted, create an extension host worker.
     *
     * Messages from the ports are forwarded to the endpoints returned by {@link createExtensionHostWorker}, and vice-versa.
     *
     * The lifetime of the extension host worker is tied to that of the content script instance:
     * when a port disconnects, the worker is terminated. This means there should always be exactly one
     * extension host worker per active instance of the content script.
     *
     */
    endpointPairs.subscribe(({ proxy, expose }) => {
        // It's necessary to wrap endpoints because browser.runtime.Port objects do not support transfering MessagePorts.
        // See https://github.com/GoogleChromeLabs/comlink/blob/master/messagechanneladapter.md
        const { worker, clientEndpoints } = createExtensionHostWorker({ wrapEndpoints: true })
        const connectPortAndEndpoint = (
            port: browser.runtime.Port,
            endpoint: Endpoint & Pick<MessagePort, 'start'>
        ): void => {
            endpoint.start()
            port.onMessage.addListener(message => {
                endpoint.postMessage(message)
            })
            endpoint.addEventListener('message', ({ data }) => {
                port.postMessage(data)
            })
        }
        // Connect proxy client endpoint
        connectPortAndEndpoint(proxy, clientEndpoints.proxy)
        // Connect expose client endpoint
        connectPortAndEndpoint(expose, clientEndpoints.expose)
        // Kill worker when either port disconnects
        proxy.onDisconnect.addListener(() => worker.terminate())
        expose.onDisconnect.addListener(() => worker.terminate())
    })

    console.log('Sourcegraph background page initialized')
}

// Browsers log this unhandled Promise automatically (and with a better stack trace through console.error)
// tslint:disable-next-line: no-floating-promises
main()
