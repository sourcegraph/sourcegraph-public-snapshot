import { Subscription } from 'rxjs'
import {
    InitializedNotification,
    InitializeParams,
    InitializeRequest,
    InitializeResult,
    RegistrationParams,
    RegistrationRequest,
} from '../../protocol'
import { Connection, createConnection, Logger, MessageTransports } from '../../protocol/jsonrpc2/connection'
import { createWebWorkerMessageTransports } from '../../protocol/jsonrpc2/transports/webWorker'
import { Commands, SourcegraphExtensionAPI } from './api'
import { createExtCommands } from './features/commands'

class ExtensionHandle<C> implements SourcegraphExtensionAPI<C> {
    public readonly commands: Commands

    private subscription = new Subscription()

    constructor(public readonly rawConnection: Connection, public readonly initializeParams: InitializeParams) {
        this.subscription.add(this.rawConnection)

        this.commands = createExtCommands(this.rawConnection)
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
    const connection = createConnection(transports, consoleLogger)
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
