import { Subscription } from 'rxjs'
import * as sourcegraph from 'sourcegraph'
import { InitializedNotification, InitializeParams, InitializeRequest, InitializeResult } from '../protocol'
import { createMessageConnection, Logger, MessageConnection, MessageTransports } from '../protocol/jsonrpc2/connection'
import { createWebWorkerMessageTransports } from '../protocol/jsonrpc2/transports/webWorker'
import { createRegisterProviderFunctions } from './api/provider'
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
 * Creates the Sourcegraph extension host and the extension API handle (which extensions access with `import
 * sourcegraph from 'sourcegraph'`).
 *
 * @param transports The message reader and writer to use for communication with the client. Defaults to
 *                   communicating using self.postMessage and MessageEvents with the parent (assuming that it is
 *                   called in a Web Worker).
 * @return A promise that resolves when the extension host is ready (and extensions may be activated in it).
 */
export async function createExtensionHost(
    transports: MessageTransports = createWebWorkerMessageTransports()
): Promise<typeof sourcegraph> {
    const connection = createMessageConnection(transports, consoleLogger)
    return new Promise<typeof sourcegraph>(resolve => {
        let initializationParams!: InitializeParams
        connection.onRequest(InitializeRequest.type, params => {
            initializationParams = params
            return {} as InitializeResult
        })
        connection.onNotification(InitializedNotification.type, () =>
            resolve(createExtensionHandle(connection, initializationParams))
        )
        connection.listen()
    })
}

function createExtensionHandle(
    rawConnection: MessageConnection,
    initializeParams: InitializeParams
): typeof sourcegraph {
    const subscription = new Subscription()
    subscription.add(rawConnection)

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

        ...createRegisterProviderFunctions(rawConnection),

        internal: {
            sync: () => rawConnection.sendRequest('ping'),
            experimentalCapabilities: initializeParams.capabilities.experimental,
            rawConnection,
        },
    }
}
