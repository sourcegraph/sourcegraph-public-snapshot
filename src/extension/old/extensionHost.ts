import { Subscription } from 'rxjs'
import {
    ConfigurationCascade,
    InitializedNotification,
    InitializeParams,
    InitializeRequest,
    InitializeResult,
    RegistrationParams,
    RegistrationRequest,
} from '../../protocol'
import {
    createMessageConnection,
    Logger,
    MessageConnection,
    MessageTransports,
} from '../../protocol/jsonrpc2/connection'
import { createWebWorkerMessageTransports } from '../../protocol/jsonrpc2/transports/webWorker'
import { Commands, Configuration, ExtensionContext, Observable, SourcegraphExtensionAPI, Window, Windows } from './api'
import { createExtCommands } from './features/commands'
import { createExtConfiguration } from './features/configuration'
import { createExtContext } from './features/context'
import { ExtWindows } from './features/windows'

class ExtensionHandle<C> implements SourcegraphExtensionAPI<C> {
    public readonly configuration: Configuration<C> & Observable<C>
    public get windows(): Windows & Observable<Window[]> {
        return this._windows
    }
    public readonly commands: Commands
    public readonly context: ExtensionContext

    private _windows: ExtWindows
    private subscription = new Subscription()

    constructor(public readonly rawConnection: MessageConnection, public readonly initializeParams: InitializeParams) {
        this.subscription.add(this.rawConnection)

        this.configuration = createExtConfiguration<C>(
            this.rawConnection,
            initializeParams.configurationCascade as ConfigurationCascade<C>
        )
        this._windows = new ExtWindows(this.rawConnection)
        this.commands = createExtCommands(this.rawConnection)
        this.context = createExtContext(this.rawConnection)
    }

    public get activeWindow(): Window | null {
        return this._windows.activeWindow
    }

    public close(): void {
        this.subscription.unsubscribe()
    }
}

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
 * Activates a Sourcegraph extension by calling its `run` entrypoint function with the Sourcegraph extension API
 * handle as the first argument.
 *
 * @template C the extension's settings
 * @param run The extension's `run` entrypoint function.
 * @param transports The message reader and writer to use for communication with the client. Defaults to
 *                   communicating using self.postMessage and MessageEvents with the parent (assuming that it is
 *                   called in a Web Worker).
 * @return A promise that resolves when the extension's `run` function has been called.
 */
export function activateExtension<C>(
    run: (sourcegraph: SourcegraphExtensionAPI<C>) => void | Promise<void>,
    transports: MessageTransports = createWebWorkerMessageTransports()
): Promise<void> {
    const connection = createMessageConnection(transports, consoleLogger)
    return new Promise<void>(resolve => {
        let initializationParams!: InitializeParams
        connection.onRequest(InitializeRequest.type, params => {
            initializationParams = params
            return {} as InitializeResult
        })
        connection.onNotification(InitializedNotification.type, () => {
            const handle = new ExtensionHandle<C>(connection, initializationParams)

            // Register some capabilities that used to be implicit, for backcompat.
            connection
                .sendRequest(RegistrationRequest.type, {
                    registrations: [
                        {
                            id: '__backcompat_1',
                            method: 'textDocument/didOpen',
                            registerOptions: { documentSelector: ['*'] },
                        },
                        {
                            id: '__backcompat_2',
                            method: 'textDocument/didClose',
                            registerOptions: { documentSelector: ['*'] },
                        },
                        {
                            id: '__backcompat_3',
                            method: 'workspace/didChangeConfiguration',
                        },
                    ],
                } as RegistrationParams)
                .then(() => {
                    run(handle)
                    resolve()
                })
                .catch(err => console.error(err))
        })

        connection.listen()
    })
}
