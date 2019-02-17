import * as comlink from 'comlink'
import { Subscription, Unsubscribable } from 'rxjs'
import * as sourcegraph from 'sourcegraph'
import { ClientAPI } from '../client/api/api'
import { createProxy } from '../common/proxy'
import { Connection, createConnection, Logger, MessageTransports } from '../protocol/jsonrpc2/connection'
import { ExtensionHostAPI } from './api/api'
import { ExtCommands } from './api/commands'
import { ExtConfiguration } from './api/configuration'
import { ExtContext } from './api/context'
import { createDecorationType } from './api/decorations'
import { ExtDocuments } from './api/documents'
import { ExtExtensions } from './api/extensions'
import { ExtLanguageFeatures } from './api/languageFeatures'
import { ExtRoots } from './api/roots'
import { ExtSearch } from './api/search'
import { ExtViews } from './api/views'
import { ExtWindows } from './api/windows'
import { Location } from './types/location'
import { Position } from './types/position'
import { Range } from './types/range'
import { Selection } from './types/selection'
import { URI } from './types/uri'

const consoleLogger: Logger = console

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
 * @param transports The message reader and writer to use for communication with the client.
 * @return An unsubscribable to terminate the extension host.
 */
export function startExtensionHost(
    transports: MessageTransports
): Unsubscribable & { __testAPI: Promise<typeof sourcegraph> } {
    const connection = createConnection(transports, consoleLogger)
    connection.listen()

    const subscription = new Subscription()
    subscription.add(connection)

    // Wait for "initialize" message from client application before proceeding to create the
    // extension host.
    let initialized = false
    const __testAPI = new Promise<typeof sourcegraph>(resolve => {
        connection.onRequest('initialize', ([initData]: [InitData]) => {
            if (initialized) {
                throw new Error('extension host is already initialized')
            }
            initialized = true
            const { unsubscribe, __testAPI } = initializeExtensionHost(connection, initData)
            subscription.add(unsubscribe)
            resolve(__testAPI)
        })
    })

    return { unsubscribe: () => subscription.unsubscribe(), __testAPI }
}

/**
 * Initializes the extension host using the {@link InitData} from the client application. It is
 * called by {@link startExtensionHost} after the {@link InitData} is received.
 *
 * The extension API is made globally available to all requires/imports of the "sourcegraph" module
 * by other scripts running in the same JavaScript context.
 *
 * @param connection The connection used to communicate with the client.
 * @param initData The information to initialize this extension host.
 * @return An unsubscribable to terminate the extension host.
 */
function initializeExtensionHost(
    connection: Connection,
    initData: InitData
): Unsubscribable & { __testAPI: typeof sourcegraph } {
    const subscriptions = new Subscription()

    const { api, subscription: apiSubscription } = createExtensionAPI(initData, connection)
    subscriptions.add(apiSubscription)

    // Make `import 'sourcegraph'` or `require('sourcegraph')` return the extension API.
    ;(global as any).require = (modulePath: string): any => {
        if (modulePath === 'sourcegraph') {
            return api
        }
        // All other requires/imports in the extension's code should not reach here because their JS
        // bundler should have resolved them locally.
        throw new Error(`require: module not found: ${modulePath}`)
    }
    subscriptions.add(() => {
        ;(global as any).require = () => {
            // Prevent callers from attempting to access the extension API after it was
            // unsubscribed.
            throw new Error(`require: Sourcegraph extension API was unsubscribed`)
        }
    })

    return { unsubscribe: () => subscriptions.unsubscribe(), __testAPI: api }
}

function createExtensionAPI(
    initData: InitData,
    connection: Connection
): { api: typeof sourcegraph; subscription: Subscription } {
    const subscriptions = new Subscription()

    // EXTENSION HOST WORKER

    /** Proxy to main thread */
    const proxy = comlink.proxy<ClientAPI>(self)

    // For debugging/tests.
    const sync = () => connection.sendRequest<void>('ping')
    connection.onRequest('ping', () => 'pong')

    const context = new ExtContext(createProxy(connection, 'context'))

    const documents = new ExtDocuments(sync)

    const extensions = new ExtExtensions()
    subscriptions.add(extensions)

    const roots = new ExtRoots()

    const windows = new ExtWindows(createProxy(connection, 'windows'), createProxy(connection, 'codeEditor'), documents)

    const views = new ExtViews(createProxy(connection, 'views'))
    subscriptions.add(views)

    const configuration = new ExtConfiguration<any>(createProxy(connection, 'configuration'))

    const languageFeatures = new ExtLanguageFeatures(proxy.languageFeatures, documents)

    const search = new ExtSearch(proxy.search)

    const commands = new ExtCommands(createProxy(connection, 'commands'))
    subscriptions.add(commands)

    // Expose the extension host API to the client (main thread)
    const extHostAPI: ExtensionHostAPI = {
        commands,
        configuration,
        documents,
        extensions,
        languageFeatures,
        roots,
        windows,
    }
    comlink.expose(extHostAPI, self)

    // Expose the extension API to extensions
    const api: typeof sourcegraph = {
        URI,
        Position,
        Range,
        Selection,
        Location,
        MarkupKind: {
            // The const enum MarkupKind values can't be used because then the `sourcegraph` module import at the
            // top of the file is emitted in the generated code. That is problematic because it hasn't been defined
            // yet (in workerMain.ts). It seems that using const enums should *not* emit an import in the generated
            // code; this is a known issue: https://github.com/Microsoft/TypeScript/issues/16671
            // https://github.com/palantir/tslint/issues/1798 https://github.com/Microsoft/TypeScript/issues/18644.
            PlainText: 'plaintext' as sourcegraph.MarkupKind.PlainText,
            Markdown: 'markdown' as sourcegraph.MarkupKind.Markdown,
        },

        app: {
            activeWindowChanges: windows.activeWindowChanges,
            get activeWindow(): sourcegraph.Window | undefined {
                return windows.getActive()
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
            get roots(): ReadonlyArray<sourcegraph.WorkspaceRoot> {
                return roots.getAll()
            },
            onDidChangeRoots: roots.changes,
            rootChanges: roots.changes,
        },

        configuration: {
            get: () => configuration.get(),
            subscribe: (next: () => void) => configuration.subscribe(next),
        },

        languages: {
            registerHoverProvider: (selector: sourcegraph.DocumentSelector, provider: sourcegraph.HoverProvider) =>
                languageFeatures.registerHoverProvider(selector, provider),

            registerDefinitionProvider: (
                selector: sourcegraph.DocumentSelector,
                provider: sourcegraph.DefinitionProvider
            ) => languageFeatures.registerDefinitionProvider(selector, provider),

            registerTypeDefinitionProvider: (
                selector: sourcegraph.DocumentSelector,
                provider: sourcegraph.TypeDefinitionProvider
            ) => languageFeatures.registerTypeDefinitionProvider(selector, provider),

            registerImplementationProvider: (
                selector: sourcegraph.DocumentSelector,
                provider: sourcegraph.ImplementationProvider
            ) => languageFeatures.registerImplementationProvider(selector, provider),

            registerReferenceProvider: (
                selector: sourcegraph.DocumentSelector,
                provider: sourcegraph.ReferenceProvider
            ) => languageFeatures.registerReferenceProvider(selector, provider),

            registerLocationProvider: (
                id: string,
                selector: sourcegraph.DocumentSelector,
                provider: sourcegraph.LocationProvider
            ) => languageFeatures.registerLocationProvider(id, selector, provider),
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

        internal: {
            sync,
            updateContext: (updates: sourcegraph.ContextValues) => context.updateContext(updates),
            sourcegraphURL: new URI(initData.sourcegraphURL),
            clientApplication: initData.clientApplication,
        },
    }
    return { api, subscription: subscriptions }
}
