import * as comlink from 'comlink'
import { isMatch } from 'lodash'
import { Subscription, Unsubscribable } from 'rxjs'
import * as sourcegraph from 'sourcegraph'

import { Location, MarkupKind, Position, Range, Selection } from '@sourcegraph/extension-api-classes'

import { EndpointPair } from '../../platform/context'
import { SettingsCascade } from '../../settings/settings'
import { ClientAPI } from '../client/api/api'
import { registerComlinkTransferHandlers } from '../util'

import { activateExtensions } from './activation'
import { ExtensionHostAPI, ExtensionHostAPIFactory } from './api/api'
import { DocumentHighlightKind } from './api/documentHighlights'
import { createExtensionAPI } from './extensionApi'
import { createExtensionHostAPI, NotificationType } from './extensionHostApi'
import { createExtensionHostState, ExtensionHostState } from './extensionHostState'

/**
 * Required information when initializing an extension host.
 */
export interface InitData {
    /** @see {@link module:sourcegraph.internal.sourcegraphURL} */
    sourcegraphURL: string

    /** @see {@link module:sourcegraph.internal.clientApplication} */
    clientApplication: 'sourcegraph' | 'other'

    /** fetched initial settings object */
    initialSettings: Readonly<SettingsCascade<object>>
}
/**
 * Starts the extension host, which runs extensions. It is a Web Worker or other similar isolated
 * JavaScript execution context. There is exactly 1 extension host, and it has zero or more
 * extensions activated (and running).
 *
 * It expects to receive a message containing {@link InitData} from the client application as the
 * first message.
 *
 * @param endpoints The endpoints to the client.
 * @returns An unsubscribable to terminate the extension host.
 */
export function startExtensionHost(
    endpoints: EndpointPair
): Unsubscribable & { extensionAPI: Promise<typeof sourcegraph> } {
    const subscription = new Subscription()

    // Wait for "initialize" message from client application before proceeding to create the
    // extension host.
    let initialized = false
    const extensionAPI = new Promise<typeof sourcegraph>(resolve => {
        const factory: ExtensionHostAPIFactory = initData => {
            if (initialized) {
                throw new Error('extension host is already initialized')
            }
            initialized = true
            const { subscription: extensionHostSubscription, extensionAPI, extensionHostAPI } = initializeExtensionHost(
                endpoints,
                initData
            )
            subscription.add(extensionHostSubscription)
            resolve(extensionAPI)
            return extensionHostAPI
        }
        comlink.expose(factory, endpoints.expose)
    })

    return { unsubscribe: () => subscription.unsubscribe(), extensionAPI }
}

/**
 * Initializes the extension host using the {@link InitData} from the client application. It is
 * called by {@link startExtensionHost} after the {@link InitData} is received.
 *
 * The extension API is made globally available to all requires/imports of the "sourcegraph" module
 * by other scripts running in the same JavaScript context.
 *
 * @param endpoints The endpoints to the client.
 * @param initData The information to initialize this extension host.
 * @returns An unsubscribable to terminate the extension host.
 */
function initializeExtensionHost(
    endpoints: EndpointPair,
    initData: InitData
): { extensionHostAPI: ExtensionHostAPI; extensionAPI: typeof sourcegraph; subscription: Subscription } {
    const subscription = new Subscription()

    const { extensionAPI, extensionHostAPI, subscription: apiSubscription } = createExtensionAndExtensionHostAPIs(
        initData,
        endpoints
    )
    subscription.add(apiSubscription)

    // Make `import 'sourcegraph'` or `require('sourcegraph')` return the extension API.
    globalThis.require = ((modulePath: string): any => {
        if (modulePath === 'sourcegraph') {
            return extensionAPI
        }
        // All other requires/imports in the extension's code should not reach here because their JS
        // bundler should have resolved them locally.
        throw new Error(`require: module not found: ${modulePath}`)
    }) as any
    subscription.add(() => {
        globalThis.require = (() => {
            // Prevent callers from attempting to access the extension API after it was
            // unsubscribed.
            throw new Error('require: Sourcegraph extension API was unsubscribed')
        }) as any
    })

    return { subscription, extensionAPI, extensionHostAPI }
}

function createExtensionAndExtensionHostAPIs(
    initData: InitData,
    endpoints: Pick<EndpointPair, 'proxy'>
): { extensionHostAPI: ExtensionHostAPI; extensionAPI: typeof sourcegraph; subscription: Subscription } {
    const subscription = new Subscription()

    // EXTENSION HOST WORKER

    registerComlinkTransferHandlers()

    /** Proxy to main thread */
    const proxy = comlink.wrap<ClientAPI>(endpoints.proxy)

    // For debugging/tests.
    const sync = async (): Promise<void> => {
        await proxy.ping()
    }

    // Create extension host state
    const extensionHostState = createExtensionHostState(initData, proxy)
    // Create extension host API
    const extensionHostAPINew = createExtensionHostAPI(extensionHostState)
    // Create extension API
    const { configuration, workspace, commands, search, languages, graphQL, content, app } = createExtensionAPI(
        extensionHostState,
        proxy
    )
    // Activate extensions
    subscription.add(activateExtensions(extensionHostState, proxy))

    // Expose the extension host API to the client (main thread)
    const extensionHostAPI: ExtensionHostAPI = {
        [comlink.proxyMarker]: true,

        ping: () => 'pong',
        ...extensionHostAPINew,
    }

    // Expose the extension API to extensions
    // "redefines" everything instead of exposing internal Ext* classes directly so as to:
    // - Avoid exposing private methods to extensions
    // - Avoid exposing proxy.* to extensions, which gives access to the main thread
    const extensionAPI: typeof sourcegraph & {
        // Backcompat definitions that were removed from sourcegraph.d.ts but are still defined (as
        // noops with a log message), to avoid completely breaking extensions that use them.
        languages: {
            registerTypeDefinitionProvider: any
            registerImplementationProvider: any
        }
    } = {
        URI: URL,
        Position,
        Range,
        Selection,
        Location,
        MarkupKind,
        NotificationType,
        DocumentHighlightKind,
        app,

        workspace,
        configuration,

        languages,

        search,
        commands,
        graphQL,
        content,

        internal: {
            sync: () => sync(),
            updateContext: (updates: sourcegraph.ContextValues) => updateContext(updates, extensionHostState),
            sourcegraphURL: new URL(initData.sourcegraphURL),
            clientApplication: initData.clientApplication,
        },
    }

    return { extensionHostAPI, extensionAPI, subscription }
}

// Context (TODO(tj): move to extension/api/context)
// Same implementation is exposed to main and extensions
export function updateContext(update: { [k: string]: unknown }, state: ExtensionHostState): void {
    if (isMatch(state.context.value, update)) {
        return
    }
    const result: any = {}
    for (const [key, oldValue] of Object.entries(state.context.value)) {
        if (update[key] !== null) {
            result[key] = oldValue
        }
    }
    for (const [key, value] of Object.entries(update)) {
        if (value !== null) {
            result[key] = value
        }
    }
    state.context.next(result)
}
