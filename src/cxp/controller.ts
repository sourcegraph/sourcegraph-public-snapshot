import { ClientOptions } from 'cxp/module/client/client'
import {
    CloseAction,
    ErrorAction,
    ErrorHandler as CXPErrorHandler,
    InitializationFailedHandler,
} from 'cxp/module/client/errorHandler'
import { ClientKey, Controller } from 'cxp/module/environment/controller'
import { Environment } from 'cxp/module/environment/environment'
import { Extension } from 'cxp/module/environment/extension'
import { MessageTransports } from 'cxp/module/jsonrpc2/connection'
import { Message, ResponseError } from 'cxp/module/jsonrpc2/messages'
import { BrowserConsoleTracer } from 'cxp/module/jsonrpc2/trace'
import { InitializeError } from 'cxp/module/protocol'
import { catchError, map, mergeMap, switchMap, take } from 'rxjs/operators'
import { Context } from '../context'
import { Settings } from '../copypasta'
import { asError, isErrorLike } from '../errors'
import { ConfiguredExtension, isExtensionEnabled } from '../extensions/extension'
import { ConfigurationSubject } from '../settings'
import { getSavedClientTrace } from './client'

/**
 * Adds the manifest to CXP extensions in the CXP environment, so we can consult it in the createMessageTransports
 * callback (to know how to communicate with or run the extension).
 */
interface ExtensionWithManifest extends Extension, Pick<ConfiguredExtension, 'manifest'> {}

/**
 * React props or state containing the CXP controller. There should be only a single CXP controller for the whole
 * application.
 */
export interface CXPControllerProps<C extends object = Settings> {
    /**
     * The CXP controller, which is used to communicate with the extensions and manages extensions based on the CXP
     * environment.
     */
    cxpController: Controller<ExtensionWithManifest, C>
}

interface CXPInitializationFailedHandler {
    initializationFailed: InitializationFailedHandler
}

/** The CXP client initializion-failed and error handler. */
class ErrorHandler implements CXPInitializationFailedHandler, CXPErrorHandler {
    /** The number of connection times to record. */
    private static MAX_CONNECTION_TIMESTAMPS = 4

    /** The timestamps of the last connection initiation times, with the 0th element being the oldest. */
    private connectionTimestamps: number[] = [Date.now()]

    public constructor(private extensionID: string) {}

    public error(err: Error, message: Message, count: number): ErrorAction {
        log(
            'error',
            `${this.extensionID}${count > 1 ? ` (count: ${count})` : ''}`,
            err,
            message ? { message } : undefined
        )

        if (err.message && err.message.includes('got unsubscribed')) {
            return ErrorAction.ShutDown
        }

        // Language servers differ in when they decide to return an error vs. just return an empty result. This
        // constant here is a guess that should be adjusted.
        if (count && count <= 5) {
            return ErrorAction.Continue
        }
        return ErrorAction.ShutDown
    }

    private computeDelayBeforeRetry(): number {
        const lastRestart: number | undefined = this.connectionTimestamps[this.connectionTimestamps.length - 1]
        const now = Date.now()

        // Bound the size of the array.
        if (this.connectionTimestamps.length === ErrorHandler.MAX_CONNECTION_TIMESTAMPS) {
            this.connectionTimestamps.shift()
        }
        this.connectionTimestamps.push(now)

        const diff = now - (lastRestart || 0)
        if (diff <= 10 * 1000) {
            // If the connection was created less than 10 seconds ago, wait longer to restart to avoid excessive
            // attempts.
            return 2500
        }
        // Otherwise restart after a shorter period.
        return 500
    }

    public initializationFailed(err: ResponseError<InitializeError> | Error | any): boolean | Promise<boolean> {
        log('error', this.extensionID, err)

        const EINVALIDREQUEST = -32600 // JSON-RPC 2.0 error code

        if (
            isResponseError(err) &&
            ((err.message.includes('dial tcp') && err.message.includes('connect: connection refused')) ||
                (err.code === EINVALIDREQUEST && err.message.includes('client proxy handler is already initialized')))
        ) {
            return false
        }

        const retry = isResponseError(err) && !!err.data && err.data.retry && this.connectionTimestamps.length === 0
        return delayed(retry, this.computeDelayBeforeRetry())
    }

    public closed(): CloseAction | Promise<CloseAction> {
        if (this.connectionTimestamps.length === ErrorHandler.MAX_CONNECTION_TIMESTAMPS) {
            const diff = this.connectionTimestamps[this.connectionTimestamps.length - 1] - this.connectionTimestamps[0]
            if (diff <= 60 * 1000) {
                // Stop restarting the server if it has restarted n times in the last minute.
                return CloseAction.DoNotReconnect
            }
        }

        return delayed(CloseAction.Reconnect, this.computeDelayBeforeRetry())
    }
}

/**
 * Filter the environment to omit extensions that should not be activated (based on their manifest's
 * activationEvents).
 *
 * @template C settings type
 */
function environmentFilter<C extends Settings>(
    nextEnvironment: Environment<ExtensionWithManifest, C>
): Environment<ExtensionWithManifest, C> {
    return {
        ...nextEnvironment,
        extensions:
            nextEnvironment.extensions &&
            nextEnvironment.extensions.filter(x => {
                try {
                    const component = nextEnvironment.component
                    if (!isExtensionEnabled(nextEnvironment.configuration, x.id)) {
                        return false
                    } else if (!x.manifest) {
                        console.warn(
                            `Extension ${
                                x.id
                            } was not found on your primary Sourcegraph instance so it will not be activated.`
                        )
                        return false
                    } else if (isErrorLike(x.manifest)) {
                        console.warn(asError(x.manifest))
                        return false
                    } else if (!x.manifest.activationEvents) {
                        console.warn(`Extension ${x.id} has no activation events, so it will never be activated.`)
                        return false
                    }
                    return x.manifest.activationEvents.some(
                        e => e === '*' || (!!component && e === `onLanguage:${component.document.languageId}`)
                    )
                } catch (err) {
                    console.error(err)
                }
                return false
            }),
    }
}

/**
 * Creates the CXP controller, which handles all CXP communication between the React app and CXP extension.
 *
 * There should only be a single controller for the entire application. The controller's environment represents all
 * of the application state that the controller needs to know.
 *
 * It receives state updates via calls to the setEnvironment method from React components. It provides results to
 * React components via its registries and the showMessages, etc., observables.
 */
export function createController<S extends ConfigurationSubject, C extends object = Settings>(
    context: Context<S, C>,
    createMessageTransports: (extension: ExtensionWithManifest, options: ClientOptions) => Promise<MessageTransports>
): Controller<ExtensionWithManifest, C> {
    const controller = new Controller<ExtensionWithManifest, C>({
        clientOptions: (key: ClientKey, options: ClientOptions, extension: ExtensionWithManifest) => {
            const errorHandler = new ErrorHandler(extension.id)
            return {
                createMessageTransports: () => createMessageTransports(extension, options),
                initializationFailedHandler: err => errorHandler.initializationFailed(err),
                errorHandler,
                trace: getSavedClientTrace(key),
                tracer: new BrowserConsoleTracer(extension.id),
            }
        },
        environmentFilter,
    })

    function messageFromExtension(extension: string, message: string): string {
        return `From extension ${extension}:\n\n${message}`
    }
    controller.showMessages.subscribe(({ extension, message }) => alert(messageFromExtension(extension, message)))
    controller.showMessageRequests.subscribe(({ extension, message, actions, resolve }) => {
        if (!actions || actions.length === 0) {
            alert(messageFromExtension(extension, message))
            resolve(null)
            return
        }
        const value = prompt(
            messageFromExtension(
                extension,
                `${message}\n\nValid responses: ${actions.map(({ title }) => JSON.stringify(title)).join(', ')}`
            ),
            actions[0].title
        )
        resolve(actions.find(a => a.title === value) || null)
    })
    controller.showInputs.subscribe(({ extension, message, defaultValue, resolve }) =>
        resolve(prompt(messageFromExtension(extension, message), defaultValue))
    )
    controller.configurationUpdates
        .pipe(
            mergeMap(params =>
                // Find the highest-precedence subject and edit its configuration.
                //
                // TODO(sqs): Allow extensions to specify which subject's configuration to update
                // (instead of always updating the highest-precedence subject's configuration).
                context.configurationCascade.pipe(
                    take(1),
                    map(x => x.subjects[x.subjects.length - 1]),
                    switchMap(subject =>
                        context
                            .updateExtensionSettings(subject.subject.id, {
                                extensionID: params.extension,
                                edit: params,
                            })
                            .pipe(
                                catchError(err => {
                                    console.error(err)
                                    return []
                                })
                            )
                    )
                )
            )
        )
        .subscribe(undefined, err => console.error(err))

    // Print window/logMessage log messages to the browser devtools console.
    controller.logMessages.subscribe(({ message, extension }) => {
        log('info', extension, message)
    })

    // Debug helpers.
    const DEBUG = true
    if (DEBUG) {
        // Debug helper: log environment changes.
        const LOG_ENVIRONMENT = false
        if (LOG_ENVIRONMENT) {
            controller.environment.environment.subscribe(environment => log('info', 'env', environment))
        }

        // Debug helpers: e.g., just run `cxp` in devtools to get a reference to this controller. (If multiple
        // controllers are created, this points to the last one created.)
        if ('cxp' in window) {
            delete (window as any).cxp
        }
        Object.defineProperty(window, 'cxp', {
            get: () => controller,
        })
        if ('cxpenv' in window) {
            delete (window as any).cxpenv
        }
        Object.defineProperty(window, 'cxpenv', {
            get: () => controller.environment.environment.value,
        })
    }

    return controller
}

function isResponseError(err: any): err is ResponseError<InitializeError> {
    return 'code' in err && 'message' in err
}

function delayed<T>(value: T, msec: number): Promise<T> {
    return new Promise(resolve => {
        setTimeout(() => resolve(value), msec)
    })
}

function log(level: 'info' | 'error', subject: string, message: any, other?: { [name: string]: any }): void {
    let f: typeof console.log
    let color: string
    let backgroundColor: string
    if (level === 'info') {
        f = console.log
        color = '#000'
        backgroundColor = '#eee'
    } else {
        f = console.error
        color = 'white'
        backgroundColor = 'red'
    }
    f(
        '%c CXP %s %c',
        `font-weight:bold;background-color:${backgroundColor};color:${color}`,
        subject,
        'font-weight:normal;background-color:unset',
        message,
        other || ''
    )
}
