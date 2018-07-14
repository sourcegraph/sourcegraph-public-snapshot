import { ClientOptions } from 'cxp/lib/client/client'
import { Controller } from 'cxp/lib/environment/controller'
import { MessageTransports } from 'cxp/lib/jsonrpc2/connection'
import { createWebSocketMessageTransports } from 'cxp/lib/jsonrpc2/transports/browserWebSocket'
import { createWebWorkerMessageTransports } from 'cxp/lib/jsonrpc2/transports/webWorker'
import { catchError, mergeMap } from 'rxjs/operators'
import { toGQLKeyPath, updateUserExtensionSettings } from '../registry/backend'
import { isErrorLike } from '../util/errors'
import { CXPExtensionWithManifest } from './CXPEnvironment'
import { importScriptsBlobURL } from './webWorker'

/**
 * Creates the CXP controller, which handles all CXP communication between the React app and CXP extension.
 *
 * There should only be a single controller for the entire application. The controller's environment represents all
 * of the application state that the controller needs to know.
 *
 * It receives state updates via calls to the setEnvironment method from React components. It provides results to
 * React components via its registries and the showMessages, etc., observables.
 */
export function createController(): Controller<CXPExtensionWithManifest> {
    const controller = new Controller<CXPExtensionWithManifest>(
        {
            initializationFailedHandler: err => {
                console.error('Initialization failed:', err)
                return false
            },
        },
        createMessageTransports
    )

    controller.showMessages.subscribe(({ message }) => alert(message))
    controller.showMessageRequests.subscribe(({ message, actions, resolve }) => {
        if (!actions || actions.length === 0) {
            alert(message)
            resolve(null)
            return
        }
        const value = prompt(
            `${message}\n\nValid responses: ${actions.map(({ title }) => JSON.stringify(title)).join(', ')}`,
            actions[0].title
        )
        resolve(actions.find(a => a.title === value) || null)
    })
    controller.configurationUpdates
        .pipe(
            mergeMap(params =>
                updateUserExtensionSettings({
                    extensionID: params.extension,
                    edit: { keyPath: toGQLKeyPath(params.path), value: params.value },
                }).pipe(
                    // TODO(sqs): Apply updated settings for this extension in the React component hierarchy, e.g. by
                    // somehow calling this.props.onExtensionsChange and merging in the update.
                    catchError(err => {
                        console.error(err)
                        return []
                    })
                )
            )
        )
        .subscribe(undefined, err => console.error(err))

    // Print window/logMessage log messages to the browser devtools console.
    controller.logMessages.subscribe(({ message, extension }) => {
        console.log(
            '%c CXP %s %c %s',
            'font-weight:bold;background-color:#eee',
            extension,
            'font-weight:normal;background-color:unset',
            message
        )
    })

    // Debug helpers.
    const DEBUG = true
    if (DEBUG) {
        // Debug helper: log environment changes.
        controller.environment.environment.subscribe(environment =>
            console.log(
                '%c CXP env %c %o',
                'font-weight:bold;background-color:#999;color:white',
                'background-color:unset;color:unset;font-weight:unset',
                environment
            )
        )

        // Debug helpers: e.g., just run `cxp` in devtools to get a reference to this controller. (If multiple
        // controllers are created, this points to the last one created.)
        Object.defineProperty(window, 'cxp', {
            get: () => controller,
        })
        Object.defineProperty(window, 'cxpenv', {
            get: () => controller.environment.environment.value,
        })
    }

    return controller
}

function createMessageTransports(
    extension: CXPExtensionWithManifest,
    clientOptions: ClientOptions
): Promise<MessageTransports> {
    if (!extension.manifest) {
        throw new Error(`unable to connect to extension ${JSON.stringify(extension.id)}: no manifest found`)
    }
    if (isErrorLike(extension.manifest)) {
        throw new Error(
            `unable to connect to extension ${JSON.stringify(extension.id)}: invalid manifest: ${
                extension.manifest.message
            }`
        )
    }
    if (extension.manifest.platform.type === 'bundle') {
        const APPLICATION_JSON_MIME_TYPE = 'application/json'
        if (
            typeof extension.manifest.platform.contentType === 'string' &&
            extension.manifest.platform.contentType !== APPLICATION_JSON_MIME_TYPE
        ) {
            // Until these are supported, prevent people from
            throw new Error(
                `unable to run extension ${JSON.stringify(extension.id)} bundle: content type ${JSON.stringify(
                    extension.manifest.platform.contentType
                )} is not supported (use ${JSON.stringify(APPLICATION_JSON_MIME_TYPE)})`
            )
        }
        const worker = new Worker(importScriptsBlobURL(extension.id, extension.manifest.platform.url))
        return Promise.resolve(createWebWorkerMessageTransports(worker))
    }

    // Include ?mode=&repo= in the url to make it easier to find the correct WebSocket connection in (e.g.) the
    // Chrome network inspector. It does not affect any behaviour.
    const url = `${window.location.protocol === 'https:' ? 'wss:' : 'ws:'}//${window.location.host}/.api/lsp?mode=${
        extension.id
    }&rootUri=${clientOptions.root}`
    return createWebSocketMessageTransports(new WebSocket(url))
}
