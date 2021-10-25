// We want to polyfill first.
import '../../shared/polyfills'

import { Endpoint } from 'comlink'
import { without } from 'lodash'
import { combineLatest, Observable, Subject, Subscription, timer } from 'rxjs'
import {
    bufferCount,
    filter,
    groupBy,
    switchMap,
    map,
    mergeMap,
    concatMap,
    mapTo,
    catchError,
    distinctUntilChanged,
} from 'rxjs/operators'
import addDomainPermissionToggle from 'webext-domain-permission-toggle'
import { patternToRegex } from 'webext-patterns'

import { createExtensionHostWorker } from '@sourcegraph/shared/src/api/extension/worker'
import { GraphQLResult, requestGraphQLCommon } from '@sourcegraph/shared/src/graphql/graphql'
import { EndpointPair } from '@sourcegraph/shared/src/platform/context'
import { isDefined } from '@sourcegraph/shared/src/util/types'

import { getHeaders } from '../../shared/backend/headers'
import { fetchSite } from '../../shared/backend/server'
import { initializeOmniboxInterface } from '../../shared/cli'
import { browserPortToMessagePort, findMessagePorts } from '../../shared/platform/ports'
import { SourcegraphUrlService } from '../../shared/platform/sourcegraphUrlService'
import { createBlobURLForBundle } from '../../shared/platform/worker'
import { initSentry } from '../../shared/sentry'
import { BrowserActionIconState, setBrowserActionIconState } from '../browser-action-icon'
import { assertEnvironment } from '../environmentAssertion'
import { fromBrowserEvent } from '../web-extension-api/fromBrowserEvent'
import { observeStorageKey } from '../web-extension-api/storage'
import { BackgroundPageApi, BackgroundPageApiHandlers } from '../web-extension-api/types'

// Interval to check if the Sourcegraph URL is valid
// This polling allows to detect if Sourcegraph instance is invalid or needs authentication.
const INTERVAL_FOR_SOURCEGRPAH_URL_CHECK = 5 /* minutes */ * 60 * 1000

assertEnvironment('BACKGROUND')

initSentry('background')

let customServerOrigins: string[] = []

/**
 * For each tab, we store a flag if we know that we are on a private
 * repository that has not been added to Cloud (+ the extension
 * points to Cloud). The content script notifies the background page if it has
 * experienced a private code on Cloud error by `notifyPrivateCloudError` message.
 */
const tabPrivateCloudErrorCache = (() => {
    const cache = new Map<number, boolean>()
    const subject = new Subject<ReadonlyMap<number, boolean>>()
    return {
        observable: subject.asObservable(),
        /**
         * Update the background page's cache of which tabs have experienced a
         * private code on Cloud error.
         */
        setTabHasPrivateCloudError(tabId: number, hasPrivateCloudError: boolean): void {
            if (!hasPrivateCloudError) {
                // An absent value is equivalent to being false; so we can delete it.
                cache.delete(tabId)
            }
            cache.set(tabId, hasPrivateCloudError)

            // Emit the updated repository cache when it changes, so that consumers can
            // observe the value.
            subject.next(cache)
        },
        getTabHasPrivateCloudError(tabId: number): boolean {
            return !!cache.get(tabId)
        },
    }
})()

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

const requestGraphQL = <T, V = object>({
    request,
    variables,
    sourcegraphURL,
}: {
    request: string
    variables: V
    sourcegraphURL: string
}): Observable<GraphQLResult<T>> =>
    requestGraphQLCommon<T, V>({
        request,
        variables,
        baseUrl: sourcegraphURL,
        headers: getHeaders(),
        credentials: 'include',
    })

async function main(): Promise<void> {
    const subscriptions = new Subscription()

    // Open installation page after the extension was installed
    browser.runtime.onInstalled.addListener(event => {
        if (event.reason === 'install') {
            browser.tabs.create({ url: browser.extension.getURL('after_install.html') }).catch(error => {
                console.error('Error opening after-install page:', error)
            })
        }
    })

    // Mirror the managed sourcegraphURL to sync storage
    subscriptions.add(
        observeStorageKey('managed', 'sourcegraphURL')
            .pipe(
                filter(isDefined),
                concatMap(sourcegraphURL => SourcegraphUrlService.setSelfHostedSourcegraphURL(sourcegraphURL))
            )
            .subscribe()
    )

    if (browser.omnibox) {
        initializeOmniboxInterface()

        // Configure the omnibox when the sourcegraphURL changes.
        subscriptions.add(
            SourcegraphUrlService.observeSelfHostedOrCloud().subscribe(sourcegraphURL => {
                configureOmnibox(sourcegraphURL)
            })
        )
    }

    // Update the browserAction icon based on the state of the extension
    subscriptions.add(
        observeBrowserActionState().subscribe(state => {
            setBrowserActionIconState(state)
        })
    )

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

    browser.tabs.onUpdated.addListener(async (tabId, changeInfo, tab) => {
        if (changeInfo.status === 'loading') {
            // A new URL is loading in the tab, so clear the cached private cloud error flag.
            tabPrivateCloudErrorCache.setTabHasPrivateCloudError(tabId, false)
            return
        }

        if (
            changeInfo.status === 'complete' &&
            customServerOrigins.some(
                origin => origin === '<all_urls>' || (!!tab.url && urlMatchesPattern(tab.url, origin))
            )
        ) {
            // Inject content script whenever a new tab was opened with a URL for which we have permission
            await browser.tabs.executeScript(tabId, { file: 'js/inject.bundle.js', runAt: 'document_end' })
        }
    })

    const handlers: BackgroundPageApiHandlers = {
        async openOptionsPage(): Promise<void> {
            await browser.runtime.openOptionsPage()
        },

        async createBlobURL(bundleUrl: string): Promise<string> {
            return createBlobURLForBundle(bundleUrl)
        },

        async requestGraphQL<T, V = object>({
            request,
            variables,
            sourcegraphURL,
        }: {
            request: string
            variables: V
            sourcegraphURL: string
        }): Promise<GraphQLResult<T>> {
            return requestGraphQL<T, V>({ request, variables, sourcegraphURL }).toPromise()
        },

        async notifyPrivateCloudError(
            hasPrivateCloudError: boolean,
            sender: browser.runtime.MessageSender
        ): Promise<void> {
            const tabId = sender.tab?.id
            if (tabId !== undefined) {
                tabPrivateCloudErrorCache.setTabHasPrivateCloudError(tabId, hasPrivateCloudError)
            }
            return Promise.resolve()
        },

        async checkPrivateCloudError(tabId: number): Promise<boolean> {
            return Promise.resolve(!!tabPrivateCloudErrorCache.getTabHasPrivateCloudError(tabId))
        },
    }

    // Handle calls from other scripts
    browser.runtime.onMessage.addListener(async (message: { type: keyof BackgroundPageApi; payload: any }, sender) => {
        const method = message.type

        if (!handlers[method]) {
            throw new Error(`Invalid RPC call for "${method}"`)
        }

        // https://stackoverflow.com/questions/55572797/why-does-typescript-expect-never-as-function-argument-when-retrieving-the-func
        return (handlers[method] as (
            payload: any,
            sender?: browser.runtime.MessageSender
        ) => ReturnType<BackgroundPageApi[typeof method]>)(message.payload, sender)
    })

    await browser.runtime.setUninstallURL('https://about.sourcegraph.com/uninstall/')

    // The `popup=true` param is used by the options page to determine if it's
    // loaded in the popup or in th standalone options page.
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
    const browserPortPairs: Observable<Record<keyof EndpointPair, browser.runtime.PortWithSender>> = fromBrowserEvent(
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

    // Extension Host Connection
    // When an Port pair is emitted, create an extension host worker.
    // Messages from the ports are forwarded to the endpoints returned by {@link createExtensionHostWorker}, and vice-versa.
    // The lifetime of the extension host worker is tied to that of the content script instance:
    // when a port disconnects, the worker is terminated. This means there should always be exactly one
    // extension host worker per active instance of the content script.
    subscriptions.add(
        browserPortPairs.subscribe({
            next: browserPortPair => {
                subscriptions.add(handleBrowserPortPair(browserPortPair))
            },
            error: error => {
                console.error('Error handling extension host client connection', error)
            },
        })
    )

    console.log('Sourcegraph background page initialized')
}

const workerBundleURL = browser.runtime.getURL('js/extensionHostWorker.bundle.js')

/**
 * Handle an incoming browser port pair coming from a content script.
 */
function handleBrowserPortPair(
    browserPortPair: Record<keyof EndpointPair, browser.runtime.PortWithSender>
): Subscription {
    /** Subscriptions for this browser port pair */
    const subscriptions = new Subscription()

    console.log('Extension host client connected')
    const { worker, clientEndpoints } = createExtensionHostWorker(workerBundleURL)
    subscriptions.add(() => worker.terminate())

    /** Forwards all messages between two endpoints (in one direction) */
    const forwardEndpoint = (from: Endpoint, to: Endpoint): void => {
        const messageListener = (event: Event): void => {
            const { data } = event as MessageEvent
            to.postMessage(data, [...findMessagePorts(data)])
        }
        from.addEventListener('message', messageListener)
        subscriptions.add(() => from.removeEventListener('message', messageListener))

        // False positive https://github.com/eslint/eslint/issues/12822
        from.start?.()
    }

    const linkPortAndEndpoint = (role: keyof EndpointPair): void => {
        const browserPort = browserPortPair[role]
        const endpoint = clientEndpoints[role]
        const tabId = browserPort.sender.tab?.id
        if (!tabId) {
            throw new Error('Expected Port to come from tab')
        }
        const link = browserPortToMessagePort(browserPort, `comlink-${role}-`, name =>
            browser.tabs.connect(tabId, { name })
        )
        subscriptions.add(link.subscription)

        forwardEndpoint(link.messagePort, endpoint)
        forwardEndpoint(endpoint, link.messagePort)

        // Clean up when the port disconnects
        const disconnectListener = subscriptions.unsubscribe.bind(subscriptions)
        browserPort.onDisconnect.addListener(disconnectListener)
        subscriptions.add(() => browserPort.onDisconnect.removeListener(disconnectListener))
    }

    // Connect proxy client endpoint
    linkPortAndEndpoint('proxy')
    // Connect expose client endpoint
    linkPortAndEndpoint('expose')

    return subscriptions
}

// Browsers log this unhandled Promise automatically (and with a better stack trace through console.error)
// eslint-disable-next-line @typescript-eslint/no-floating-promises
main()

function validateSite(sourcegraphURL: string): Observable<boolean> {
    return fetchSite(options => requestGraphQL({ ...options, sourcegraphURL })).pipe(
        mapTo(true),
        catchError(() => [false])
    )
}

/**
 * Return an observable of the currently active tab id, which changes every time
 * the user opens or switches to a different tab.
 */
function observeCurrentTabId(): Observable<number> {
    return fromBrowserEvent(browser.tabs.onActivated).pipe(map(([event]) => event.tabId))
}

/**
 * Returns an observable that indicates whether the current tab has experienced
 * a private code on Cloud error.
 */
function observeCurrentTabPrivateCloudError(): Observable<boolean> {
    return combineLatest([observeCurrentTabId(), tabPrivateCloudErrorCache.observable]).pipe(
        map(([tabId, privateCloudErrorCache]) => !!privateCloudErrorCache.get(tabId)),
        distinctUntilChanged()
    )
}

function observeSourcegraphUrlValidation(): Observable<boolean> {
    // TODO: check if we need to check both URLs
    return SourcegraphUrlService.getSelfHostedSourcegraphURL().pipe(
        filter(url => !!url),
        switchMap(url => timer(0, INTERVAL_FOR_SOURCEGRPAH_URL_CHECK).pipe(() => validateSite(url as string)))
    )
}

function observeBrowserActionState(): Observable<BrowserActionIconState> {
    return combineLatest([
        observeStorageKey('sync', 'disableExtension'),
        observeSourcegraphUrlValidation(),
        observeCurrentTabPrivateCloudError(),
    ]).pipe(
        map(([isDisabled, isSourcegraphUrlValid, hasPrivateCloudError]) => {
            if (isDisabled) {
                return 'inactive'
            }

            if (!isSourcegraphUrlValid || hasPrivateCloudError) {
                return 'active-with-alert'
            }

            return 'active'
        }),
        distinctUntilChanged()
    )
}

function urlMatchesPattern(url: string, originPermissionPattern: string): boolean {
    // Workaround for bug in `webext-patterns`. Remove workaround when fixed upstream.
    // https://github.com/fregante/webext-patterns/issues/2
    if (originPermissionPattern.includes('://*.')) {
        const bareDomainPattern = originPermissionPattern.replace('://*.', '://')
        if (patternToRegex(bareDomainPattern).test(url)) {
            return true
        }
    }

    return patternToRegex(originPermissionPattern).test(url)
}
