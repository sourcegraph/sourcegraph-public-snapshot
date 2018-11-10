import { Subscription } from 'rxjs'
import * as sourcegraph from 'sourcegraph'
import { createProxy, handleRequests } from '../common/proxy'
import { Connection, createConnection, Logger, MessageTransports } from '../protocol/jsonrpc2/connection'
import { createWebWorkerMessageTransports } from '../protocol/jsonrpc2/transports/webWorker'
import { ExtCommands } from './api/commands'
import { ExtConfiguration } from './api/configuration'
import { ExtContext } from './api/context'
import { ExtDocuments } from './api/documents'
import { ExtLanguageFeatures } from './api/languageFeatures'
import { ExtSearch } from './api/search'
import { ExtViews } from './api/views'
import { ExtWindows } from './api/windows'
import { Location } from './types/location'
import { Position } from './types/position'
import { Range } from './types/range'
import { Selection } from './types/selection'
import { URI } from './types/uri'

const consoleLogger: Logger = {
    error(message: string): void {
        console.error(message)
    },
    warn(message: string): void {
        console.warn(message)
    },
    info(message: string): void {
        console.info(message)
    },
    log(message: string): void {
        console.log(message)
    },
}

/**
 * Required information when initializing an extension host.
 */
export interface InitData {
    /** The URL to the JavaScript source file (that exports an `activate` function) for the extension. */
    bundleURL: string

    /** @see {@link module:sourcegraph.internal.sourcegraphURL} */
    sourcegraphURL: string

    /** @see {@link module:sourcegraph.internal.clientApplication} */
    clientApplication: 'sourcegraph' | 'other'
}

/**
 * Creates the Sourcegraph extension host and the extension API handle (which extensions access with `import
 * sourcegraph from 'sourcegraph'`).
 *
 * @param initData The information to initialize this extension host.
 * @param transports The message reader and writer to use for communication with the client. Defaults to
 *                   communicating using self.postMessage and MessageEvents with the parent (assuming that it is
 *                   called in a Web Worker).
 * @return The extension API.
 */
export function createExtensionHost(
    initData: InitData,
    transports: MessageTransports = createWebWorkerMessageTransports()
): typeof sourcegraph {
    const connection = createConnection(transports, consoleLogger)
    connection.listen()
    return createExtensionHandle(initData, connection)
}

function createExtensionHandle(initData: InitData, connection: Connection): typeof sourcegraph {
    const subscription = new Subscription()
    subscription.add(connection)

    // For debugging/tests.
    const sync = () => connection.sendRequest<void>('ping')
    connection.onRequest('ping', () => 'pong')

    const proxy = (prefix: string) => createProxy((name, args) => connection.sendRequest(`${prefix}/${name}`, args))

    const context = new ExtContext(proxy('context'))
    handleRequests(connection, 'context', context)

    const documents = new ExtDocuments(sync)
    handleRequests(connection, 'documents', documents)

    const windows = new ExtWindows(proxy('windows'), proxy('codeEditor'), documents)
    handleRequests(connection, 'windows', windows)

    const views = new ExtViews(proxy('views'))
    handleRequests(connection, 'views', views)

    const configuration = new ExtConfiguration<any>(proxy('configuration'))
    handleRequests(connection, 'configuration', configuration)

    const languageFeatures = new ExtLanguageFeatures(proxy('languageFeatures'), documents)
    handleRequests(connection, 'languageFeatures', languageFeatures)

    const search = new ExtSearch(proxy('search'))
    handleRequests(connection, 'search', search)

    const commands = new ExtCommands(proxy('commands'))
    handleRequests(connection, 'commands', commands)

    return {
        URI,
        Position,
        Range,
        Selection,
        Location,
        MarkupKind: {
            PlainText: sourcegraph.MarkupKind.PlainText,
            Markdown: sourcegraph.MarkupKind.Markdown,
        },

        app: {
            get activeWindow(): sourcegraph.Window | undefined {
                return windows.getActive()
            },
            get windows(): sourcegraph.Window[] {
                return windows.getAll()
            },
            createPanelView: id => views.createPanelView(id),
        },

        workspace: {
            get textDocuments(): sourcegraph.TextDocument[] {
                return documents.getAll()
            },
            onDidOpenTextDocument: documents.onDidOpenTextDocument,
        },

        configuration: {
            get: () => configuration.get(),
            subscribe: next => configuration.subscribe(next),
        },

        languages: {
            registerHoverProvider: (selector, provider) => languageFeatures.registerHoverProvider(selector, provider),
            registerDefinitionProvider: (selector, provider) =>
                languageFeatures.registerDefinitionProvider(selector, provider),
            registerTypeDefinitionProvider: (selector, provider) =>
                languageFeatures.registerTypeDefinitionProvider(selector, provider),
            registerImplementationProvider: (selector, provider) =>
                languageFeatures.registerImplementationProvider(selector, provider),
            registerReferenceProvider: (selector, provider) =>
                languageFeatures.registerReferenceProvider(selector, provider),
        },

        search: {
            registerQueryTransformer: provider => search.registerQueryTransformer(provider),
        },

        commands: {
            registerCommand: (command, callback) => commands.registerCommand({ command, callback }),
            executeCommand: (command, ...args) => commands.executeCommand(command, args),
        },

        internal: {
            sync,
            updateContext: updates => context.updateContext(updates),
            sourcegraphURL: new URI(initData.sourcegraphURL),
            clientApplication: initData.clientApplication,
        },
    }
}
