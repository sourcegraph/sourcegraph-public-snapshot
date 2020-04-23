import * as comlink from '@sourcegraph/comlink'
import { Location, MarkupKind, Position, Range, Selection } from '@sourcegraph/extension-api-classes'
import { Subscription, Unsubscribable } from 'rxjs'
import * as sourcegraph from 'sourcegraph'
import { EndpointPair } from '../../platform/context'
import { ClientAPI } from '../client/api/api'
import { NotificationType } from '../client/services/notifications'
import { ExtensionHostAPI, ExtensionHostAPIFactory } from './api/api'
import { ExtCommands } from './api/commands'
import { ExtConfiguration } from './api/configuration'
import { ExtContent } from './api/content'
import { ExtContext } from './api/context'
import { createDecorationType } from './api/decorations'
import { ExtDocuments } from './api/documents'
import { ExtExtensions } from './api/extensions'
import { ExtLanguageFeatures } from './api/languageFeatures'
import { ExtRoots } from './api/roots'
import { ExtSearch } from './api/search'
import { ExtViews } from './api/views'
import { ExtWindows } from './api/windows'
import { registerComlinkTransferHandlers } from '../util'

/**
 * Required information when initializing an extension host.
 */
export interface InitData {
    /** @see {@link module:sourcegraph.internal.sourcegraphURL} */
    sourcegraphURL: string

    /** @see {@link module:sourcegraph.internal.clientApplication} */
    clientApplication: 'sourcegraph' | 'other'
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
            const { subscription: extHostSubscription, extensionAPI, extensionHostAPI } = initializeExtensionHost(
                endpoints,
                initData
            )
            subscription.add(extHostSubscription)
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

    const { extensionAPI, extensionHostAPI, subscription: apiSubscription } = createExtensionAPI(initData, endpoints)
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

function createExtensionAPI(
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
    const context = new ExtContext(proxy.context)
    const documents = new ExtDocuments(sync)

    const extensions = new ExtExtensions()
    subscription.add(extensions)

    const roots = new ExtRoots()
    const windows = new ExtWindows(proxy, documents)
    const views = new ExtViews(proxy.views)
    const configuration = new ExtConfiguration<any>(proxy.configuration)
    const languageFeatures = new ExtLanguageFeatures(proxy.languageFeatures, documents)
    const search = new ExtSearch(proxy.search)
    const commands = new ExtCommands(proxy.commands)
    const content = new ExtContent(proxy.content)

    // Expose the extension host API to the client (main thread)
    const extensionHostAPI: ExtensionHostAPI = {
        [comlink.proxyMarker]: true,

        ping: () => 'pong',
        configuration,
        documents,
        extensions,
        roots,
        windows,
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
        app: {
            activeWindowChanges: windows.activeWindowChanges,
            get activeWindow(): sourcegraph.Window | undefined {
                return windows.activeWindow
            },
            get windows(): sourcegraph.Window[] {
                return windows.getAll()
            },
            createPanelView: (id: string) => views.createPanelView(id),
            createDecorationType,
        },

        workspace: {
            get textDocuments(): sourcegraph.TextDocument[] {
                return documents.getAll()
            },
            onDidOpenTextDocument: documents.openedTextDocuments,
            openedTextDocuments: documents.openedTextDocuments,
            get roots(): readonly sourcegraph.WorkspaceRoot[] {
                return roots.getAll()
            },
            onDidChangeRoots: roots.changes,
            rootChanges: roots.changes,
        },

        configuration: Object.assign(configuration.changes.asObservable(), {
            get: () => configuration.get(),
        }),

        languages: {
            registerHoverProvider: (selector: sourcegraph.DocumentSelector, provider: sourcegraph.HoverProvider) =>
                languageFeatures.registerHoverProvider(selector, provider),

            registerDefinitionProvider: (
                selector: sourcegraph.DocumentSelector,
                provider: sourcegraph.DefinitionProvider
            ) => languageFeatures.registerDefinitionProvider(selector, provider),

            // These were removed, but keep them here so that calls from old extensions do not throw
            // an exception and completely break.
            registerTypeDefinitionProvider: () => {
                console.warn(
                    'sourcegraph.languages.registerTypeDefinitionProvider was removed. Use sourcegraph.languages.registerLocationProvider instead.'
                )
                return { unsubscribe: () => undefined }
            },
            registerImplementationProvider: () => {
                console.warn(
                    'sourcegraph.languages.registerImplementationProvider was removed. Use sourcegraph.languages.registerLocationProvider instead.'
                )
                return { unsubscribe: () => undefined }
            },

            registerReferenceProvider: (
                selector: sourcegraph.DocumentSelector,
                provider: sourcegraph.ReferenceProvider
            ) => languageFeatures.registerReferenceProvider(selector, provider),

            registerLocationProvider: (
                id: string,
                selector: sourcegraph.DocumentSelector,
                provider: sourcegraph.LocationProvider
            ) => languageFeatures.registerLocationProvider(id, selector, provider),

            registerCompletionItemProvider: (
                selector: sourcegraph.DocumentSelector,
                provider: sourcegraph.CompletionItemProvider
            ) => languageFeatures.registerCompletionItemProvider(selector, provider),
        },

        search: {
            registerQueryTransformer: (provider: sourcegraph.QueryTransformer) =>
                search.registerQueryTransformer(provider),
        },

        commands: {
            registerCommand: (command: string, callback: (...args: any[]) => any) =>
                commands.registerCommand({ command, callback }),

            executeCommand: (command: string, ...args: any[]) => commands.executeCommand(command, args),
        },

        content: {
            registerLinkPreviewProvider: (urlMatchPattern: string, provider: sourcegraph.LinkPreviewProvider) =>
                content.registerLinkPreviewProvider(urlMatchPattern, provider),
        },

        internal: {
            sync,
            updateContext: (updates: sourcegraph.ContextValues) => context.updateContext(updates),
            sourcegraphURL: new URL(initData.sourcegraphURL),
            clientApplication: initData.clientApplication,
        },
    }
    return { extensionHostAPI, extensionAPI, subscription }
}
