import { Subscription } from 'rxjs'

import { createExtensionHostClientConnection } from '../api/client/connection'
import type { InitData } from '../api/extension/extensionHost'
import { syncPromiseSubscription } from '../api/util'
import type { PlatformContext } from '../platform/context'

import type { Controller } from './controller'

/**
 * Creates the controller, which handles all communication between the client application and extensions.
 *
 * There should only be a single controller for the entire client application. The controller's model represents
 * all of the client application state that the client needs to know.
 *
 * The implementation (`createExtensionHostClientConnection`) is lazy loaded to avoid adding bytes when
 * the extension system is disabled
 */
export function createController(
    context: Pick<
        PlatformContext,
        | 'updateSettings'
        | 'settings'
        | 'getGraphQLClient'
        | 'requestGraphQL'
        | 'getStaticExtensions'
        | 'telemetryService'
        | 'telemetryRecorder'
        | 'clientApplication'
        | 'sourcegraphURL'
        | 'createExtensionHost'
    >
): Controller {
    const subscriptions = new Subscription()
    const initData: Omit<InitData, 'initialSettings'> = {
        sourcegraphURL: context.sourcegraphURL,
        clientApplication: context.clientApplication,
    }
    const extensionHostClientPromise = createExtensionHostClientConnection(
        context.createExtensionHost(),
        initData,
        context
    )

    subscriptions.add(() => extensionHostClientPromise.then(({ subscription }) => subscription.unsubscribe()))

    // TODO: Debug helpers, logging

    return {
        executeCommand: parameters =>
            extensionHostClientPromise.then(({ exposedToClient }) => exposedToClient.executeCommand(parameters)),
        registerCommand: entryToRegister =>
            syncPromiseSubscription(
                extensionHostClientPromise.then(({ exposedToClient }) =>
                    exposedToClient.registerCommand(entryToRegister)
                )
            ),
        extHostAPI: extensionHostClientPromise.then(({ api }) => api),
        unsubscribe: () => subscriptions.unsubscribe(),
    }
}
