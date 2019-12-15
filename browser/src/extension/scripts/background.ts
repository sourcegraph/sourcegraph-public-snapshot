// We want to polyfill first.
import '../polyfills'

import { Endpoint } from '@sourcegraph/comlink'
import { without } from 'lodash'
import { noop, Observable } from 'rxjs'
import { bufferCount, filter, groupBy, map, mergeMap, switchMap, take } from 'rxjs/operators'
import addDomainPermissionToggle from 'webext-domain-permission-toggle'
import { createExtensionHostWorker } from '../../../../shared/src/api/extension/worker'
import { GraphQLResult, requestGraphQL as requestGraphQLCommon } from '../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../shared/src/graphql/schema'
import { BackgroundMessageHandlers } from '../../browser/types'
import { initializeOmniboxInterface } from '../../libs/cli'
import { initSentry } from '../../libs/sentry'
import { createBlobURLForBundle } from '../../platform/worker'
import { getHeaders } from '../../shared/backend/headers'
import { fromBrowserEvent } from '../../shared/util/browser'
import { observeSourcegraphURL } from '../../shared/util/context'
import { assertEnv } from '../envAssertion'
import { observeStorageKey, storage } from '../../browser/storage'
import { isDefined } from '../../../../shared/src/util/types'

const IS_EXTENSION = true

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

const requestGraphQL = <T extends GQL.IQuery | GQL.IMutation>({
    request,
    variables,
}: {
    request: string
    variables: {}
}): Observable<GraphQLResult<T>> =>
    observeSourcegraphURL(IS_EXTENSION).pipe(
        take(1),
        switchMap(sourcegraphURL =>
            requestGraphQLCommon<T>({
                request,
                variables,
                baseUrl: sourcegraphURL,
                headers: getHeaders(),
                credentials: 'include',
            })
        )
    )

initializeOmniboxInterface(requestGraphQL)

async function main(): Promise<void> {
    // Mirror the managed sourcegraphURL to sync storage
    observeStorageKey('managed', 'sourcegraphURL')
        .pipe(filter(isDefined))
        // eslint-disable-next-line @typescript-eslint/no-misused-promises
        .subscribe(async sourcegraphURL => {
            await storage.sync.set({ sourcegraphURL })
        })
    // Configure the omnibox when the sourcegraphURL changes.
    observeSourcegraphURL(IS_EXTENSION).subscribe(sourcegraphURL => {
        configureOmnibox(sourcegraphURL)
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
        browser.permissions.onAdded.addListener(permissions => {
            if (!permissions.origins) {
                return
            }
            const origins = without(permissions.origins, ...jsContentScriptOrigins)
            customServerOrigins.push(...origins)
        })
    }
    if (browser.permissions.onRemoved) {
        browser.permissions.onRemoved.addListener(permissions => {
            if (!permissions.origins) {
                return
            }
            customServerOrigins = without(customServerOrigins, ...permissions.origins)
            const urlsToRemove: string[] = []
            for (const url of permissions.origins) {
                urlsToRemove.push(url.replace('/*', ''))
            }
        })
    }

    // Inject content script whenever a new tab was opened with a URL that we have permissions for
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
        async openOptionsPage(): Promise<void> {
            await browser.runtime.openOptionsPage()
        },

        async createBlobURL(bundleUrl: string): Promise<string> {
            return createBlobURLForBundle(bundleUrl)
        },

        async requestGraphQL<T extends GQL.IQuery | GQL.IMutation>({
            request,
            variables,
        }: {
            request: string
            variables: {}
        }): Promise<GraphQLResult<T>> {
            return requestGraphQL<T>({ request, variables }).toPromise()
        },
    }

    // Handle calls from other scripts
    browser.runtime.onMessage.addListener(async message => {
        const method = message.type as keyof typeof handlers
        if (!handlers[method]) {
            throw new Error(`Invalid RPC call for "${method}"`)
        }
        return handlers[method](message.payload)
    })

    await browser.runtime.setUninstallURL('https://about.sourcegraph.com/uninstall/')

    browser.browserAction.onClicked.addListener(noop)
    browser.browserAction.setBadgeText({ text: '' })
    browser.browserAction.setPopup({ popup: 'options.html?popup=true' })

    // Add "Enable Sourcegraph on this domain" context menu item
    addDomainPermissionToggle()

    const ENDPOINT_KIND_REGEX = /^(proxy|expose)-/

    const portKind = (port: browser.runtime.Port): string | undefined => {
        const match = port.name.match(ENDPOINT_KIND_REGEX)
        return match?.[1]
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
    const endpointPairs: Observable<Record<'proxy' | 'expose', browser.runtime.Port>> = fromBrowserEvent(
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
    endpointPairs.subscribe(
        ({ proxy, expose }) => {
            console.log('Extension host client connected')
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
        },
        err => {
            console.error('Error handling extension host client connection', err)
        }
    )

    console.log('Sourcegraph background page initialized')
}

// Browsers log this unhandled Promise automatically (and with a better stack trace through console.error)
// eslint-disable-next-line @typescript-eslint/no-floating-promises
main()
