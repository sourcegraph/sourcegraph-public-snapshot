import { ClientOptions } from 'cxp/module/client/client'
import { ClientKey, Controller as BaseController } from 'cxp/module/environment/controller'
import { Environment } from 'cxp/module/environment/environment'
import { MessageTransports } from 'cxp/module/jsonrpc2/connection'
import { BrowserConsoleTracer } from 'cxp/module/jsonrpc2/trace'
import { Contributions, ExecuteCommandParams, MessageType } from 'cxp/module/protocol'
import { Subject, Unsubscribable } from 'rxjs'
import { filter, map, mergeMap } from 'rxjs/operators'
import { Notification } from '../app/notifications/notification'
import { Context } from '../context'
import { asError, isErrorLike } from '../errors'
import { ConfiguredExtension, isExtensionEnabled } from '../extensions/extension'
import { CXPExtensionManifest } from '../schema/extension.schema'
import { ConfigurationCascade, ConfigurationSubject, Settings } from '../settings'
import { getSavedClientTrace } from './client'
import { registerBuiltinClientCommands, updateConfiguration } from './clientCommands'
import { ErrorHandler } from './errorHandler'
import { log } from './log'

/**
 * Extends the base CXP {@link BaseController} class to add functionality that is useful to this package's
 * consumers.
 */
export class Controller<S extends ConfigurationSubject, C extends Settings> extends BaseController<
    ConfiguredExtension,
    ConfigurationCascade<S, C>
> {
    /**
     * Global notification messages that should be displayed to the user, from the following sources:
     *
     * - CXP window/showMessage notifications from extensions
     * - Errors thrown or returned in command invocation
     */
    public readonly notifications = new Subject<Notification>()

    /**
     * Executes the command (registered in the CommandRegistry) specified in params. If an error is thrown, the error
     * is returned *and* emitted on the {@link Controller#notifications} observable.
     *
     * All callers should execute commands using this method instead of calling
     * {@link cxp:CommandRegistry#executeCommand} directly (to ensure errors are emitted as notifications).
     */
    public executeCommand(params: ExecuteCommandParams): Promise<any> {
        return this.registries.commands.executeCommand(params).catch(err => {
            this.notifications.next({ message: err, type: MessageType.Error, source: params.command })
            return Promise.reject(err)
        })
    }
}

/**
 * React props or state containing the CXP controller. There should be only a single CXP controller for the whole
 * application.
 */
export interface CXPControllerProps<S extends ConfigurationSubject, C extends Settings> {
    /**
     * The CXP controller, which is used to communicate with the extensions and manages extensions based on the CXP
     * environment.
     */
    cxpController: Controller<S, C>
}

/**
 * Filter the environment to omit extensions that should not be activated (based on their manifest's
 * activationEvents).
 *
 * @template CC configuration cascade type
 */
function environmentFilter<S extends ConfigurationSubject, CC extends ConfigurationCascade<S>>(
    nextEnvironment: Environment<ConfiguredExtension, CC>
): Environment<ConfiguredExtension, CC> {
    return {
        ...nextEnvironment,
        extensions:
            nextEnvironment.extensions &&
            nextEnvironment.extensions.filter(x => {
                try {
                    const component = nextEnvironment.component
                    if (!isExtensionEnabled(nextEnvironment.configuration.merged, x.id)) {
                        return false
                    } else if (!x.manifest) {
                        console.warn(
                            `Extension ${x.id} was not found. Remove it from settings to suppress this warning.`
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
export function createController<S extends ConfigurationSubject, C extends Settings>(
    context: Context<S, C>,
    createMessageTransports: (extension: ConfiguredExtension, options: ClientOptions) => Promise<MessageTransports>
): Controller<S, C> {
    const controller = new Controller<S, C>({
        clientOptions: (key: ClientKey, options: ClientOptions, extension: ConfiguredExtension) => {
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

    registerBuiltinClientCommands(context, controller)
    registerExtensionContributions(controller)

    // Show messages (that don't need user input) as global notifications.
    controller.showMessages.subscribe(({ message, type, extension }) =>
        controller.notifications.next({ message, type, source: extension })
    )

    function messageFromExtension(extension: string, message: string): string {
        return `From extension ${extension}:\n\n${message}`
    }
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
        .pipe(mergeMap(params => updateConfiguration(context, params)))
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
            get: () => {
                // This value is synchronously available because observable has an underlying
                // BehaviorSubject source.
                let value: Environment | undefined
                controller.environment.environment.subscribe(v => (value = v)).unsubscribe()
                if (value === undefined) {
                    throw new Error('environment was not synchronously available')
                }
                return value!
            },
        })
    }

    return controller
}

/**
 * Registers the builtin client commands that are required by CXP. See
 * {@link module:cxp/module/protocol/contribution.ActionContribution#command} for documentation.
 */
function registerExtensionContributions<S extends ConfigurationSubject, C extends Settings>(
    controller: Controller<S, C>
): Unsubscribable {
    const contributions = controller.environment.environment.pipe(
        map(({ extensions }) => extensions),
        filter((extensions): extensions is ConfiguredExtension[] => !!extensions),
        map(extensions =>
            extensions
                .map(({ manifest }) => manifest)
                .filter((manifest): manifest is CXPExtensionManifest => manifest !== null && !isErrorLike(manifest))
                .map(({ contributes }) => contributes)
                .filter((contributions): contributions is Contributions => !!contributions)
        )
    )
    return controller.registries.contribution.registerContributions({
        contributions,
    })
}
