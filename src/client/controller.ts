import { from, Subject, Unsubscribable } from 'rxjs'
import { filter, map, mergeMap } from 'rxjs/operators'
import { Controller as BaseController, ExtensionConnectionKey } from 'sourcegraph/module/client/controller'
import { Environment } from 'sourcegraph/module/client/environment'
import { ExecuteCommandParams } from 'sourcegraph/module/client/providers/command'
import { Contributions, MessageType } from 'sourcegraph/module/protocol'
import { MessageTransports } from 'sourcegraph/module/protocol/jsonrpc2/connection'
import { BrowserConsoleTracer, Trace } from 'sourcegraph/module/protocol/jsonrpc2/trace'
import { ExtensionStatus } from '../app/ExtensionStatus'
import { Notification } from '../app/notifications/notification'
import { Context } from '../context'
import { asError, isErrorLike } from '../errors'
import { ConfiguredExtension, isExtensionEnabled } from '../extensions/extension'
import { ExtensionManifest } from '../schema/extension.schema'
import { ConfigurationCascade, ConfigurationSubject, Settings } from '../settings'
import { registerBuiltinClientCommands, updateConfiguration } from './clientCommands'
import { log } from './log'

/**
 * Extends the {@link BaseController} class to add functionality that is useful to this package's consumers.
 */
export class Controller<S extends ConfigurationSubject, C extends Settings> extends BaseController<
    ConfiguredExtension,
    ConfigurationCascade<S, C>
> {
    /**
     * Global notification messages that should be displayed to the user, from the following sources:
     *
     * - window/showMessage notifications from extensions
     * - Errors thrown or returned in command invocation
     */
    public readonly notifications = new Subject<Notification>()

    /**
     * Executes the command (registered in the CommandRegistry) specified in params. If an error is thrown, the
     * error is returned *and* emitted on the {@link Controller#notifications} observable.
     *
     * All callers should execute commands using this method instead of calling
     * {@link sourcegraph:CommandRegistry#executeCommand} directly (to ensure errors are
     * emitted as notifications).
     */
    public executeCommand(params: ExecuteCommandParams): Promise<any> {
        return this.registries.commands.executeCommand(params).catch(err => {
            this.notifications.next({ message: err, type: MessageType.Error, source: params.command })
            return Promise.reject(err)
        })
    }
}

/**
 * React props or state containing the controller. There should be only a single controller for the whole
 * application.
 */
export interface ControllerProps<S extends ConfigurationSubject, C extends Settings> {
    /**
     * The controller, which is used to communicate with the extensions and manages extensions based on the
     * environment.
     */
    extensionsController: Controller<S, C>
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
                    const visibleTextDocumentLanguages = nextEnvironment.visibleTextDocuments
                        ? nextEnvironment.visibleTextDocuments.map(({ languageId }) => languageId)
                        : []
                    return x.manifest.activationEvents.some(
                        e => e === '*' || visibleTextDocumentLanguages.some(l => e === `onLanguage:${l}`)
                    )
                } catch (err) {
                    console.error(err)
                }
                return false
            }),
    }
}

declare global {
    interface Window {
        sx: any
        sxenv: any
    }
}

/**
 * Creates the controller, which handles all communication between the React app and Sourcegraph extensions.
 *
 * There should only be a single controller for the entire application. The controller's environment represents all
 * of the application state that the controller needs to know.
 *
 * It receives state updates via calls to the setEnvironment method from React components. It provides results to
 * React components via its registries and the showMessages, etc., observables.
 */
export function createController<S extends ConfigurationSubject, C extends Settings>(
    context: Context<S, C>,
    createMessageTransports: (extension: ConfiguredExtension) => Promise<MessageTransports>
): Controller<S, C> {
    const controller = new Controller<S, C>({
        clientOptions: (_key: ExtensionConnectionKey, extension: ConfiguredExtension) => ({
            createMessageTransports: () => createMessageTransports(extension),
        }),
        environmentFilter,
    })

    // Apply trace settings.
    //
    // HACK(sqs): This is inefficient and doesn't unsubscribe itself.
    controller.clientEntries.subscribe(entries => {
        const traceEnabled = localStorage.getItem(ExtensionStatus.TRACE_STORAGE_KEY) !== null
        for (const e of entries) {
            e.connection
                .then(c => c.trace(traceEnabled ? Trace.Verbose : Trace.Off, new BrowserConsoleTracer(e.key.id)))
                .catch(err => console.error(err))
        }
    })

    registerBuiltinClientCommands(context, controller)
    registerExtensionContributions(controller)

    // Show messages (that don't need user input) as global notifications.
    controller.showMessages.subscribe(({ message, type }) => controller.notifications.next({ message, type }))

    function messageFromExtension(message: string): string {
        return `From extension:\n\n${message}`
    }
    controller.showMessageRequests.subscribe(({ message, actions, resolve }) => {
        if (!actions || actions.length === 0) {
            alert(messageFromExtension(message))
            resolve(null)
            return
        }
        const value = prompt(
            messageFromExtension(
                `${message}\n\nValid responses: ${actions.map(({ title }) => JSON.stringify(title)).join(', ')}`
            ),
            actions[0].title
        )
        resolve(actions.find(a => a.title === value) || null)
    })
    controller.showInputs.subscribe(({ message, defaultValue, resolve }) =>
        resolve(prompt(messageFromExtension(message), defaultValue))
    )
    controller.configurationUpdates
        .pipe(
            mergeMap(params => {
                const update = updateConfiguration(context, params)
                params.resolve(update)
                return from(update)
            })
        )
        .subscribe(undefined, err => console.error(err))

    // Print window/logMessage log messages to the browser devtools console.
    controller.logMessages.subscribe(({ message }) => {
        log('info', 'EXT', message)
    })

    // Debug helpers.
    const DEBUG = true
    if (DEBUG) {
        // Debug helper: log environment changes.
        const LOG_ENVIRONMENT = false
        if (LOG_ENVIRONMENT) {
            controller.environment.subscribe(environment => log('info', 'env', environment))
        }

        // Debug helpers: e.g., just run `sx` in devtools to get a reference to this controller. (If multiple
        // controllers are created, this points to the last one created.)
        window.sx = controller
        // This value is synchronously available because observable has an underlying
        // BehaviorSubject source.
        controller.environment.subscribe(v => (window.sxenv = v)).unsubscribe()
    }

    return controller
}

/**
 * Registers the builtin client commands that are required by Sourcegraph extensions. See
 * {@link module:sourcegraph.module/protocol/contribution.ActionContribution#command} for
 * documentation.
 */
function registerExtensionContributions<S extends ConfigurationSubject, C extends Settings>(
    controller: Controller<S, C>
): Unsubscribable {
    const contributions = controller.environment.pipe(
        map(({ extensions }) => extensions),
        filter((extensions): extensions is ConfiguredExtension[] => !!extensions),
        map(extensions =>
            extensions
                .map(({ manifest }) => manifest)
                .filter((manifest): manifest is ExtensionManifest => manifest !== null && !isErrorLike(manifest))
                .map(({ contributes }) => contributes)
                .filter((contributions): contributions is Contributions => !!contributions)
        )
    )
    return controller.registries.contribution.registerContributions({
        contributions,
    })
}
