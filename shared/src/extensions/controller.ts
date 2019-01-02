import { combineLatest, from, Observable, Subject, Subscription, Unsubscribable } from 'rxjs'
import { distinctUntilChanged, map, share, switchMap, tap } from 'rxjs/operators'
import { createExtensionHostClient } from '../api/client/client'
import { Services } from '../api/client/services'
import { ExecuteCommandParams } from '../api/client/services/command'
import { ContributionRegistry } from '../api/client/services/contribution'
import { ExtensionsService } from '../api/client/services/extensionsService'
import { MessageType } from '../api/client/services/notifications'
import { InitData } from '../api/extension/extensionHost'
import { Contributions } from '../api/protocol'
import { createConnection } from '../api/protocol/jsonrpc2/connection'
import { BrowserConsoleTracer } from '../api/protocol/jsonrpc2/trace'
import { registerBuiltinClientCommands } from '../commands/commands'
import { Notification } from '../notifications/notification'
import { PlatformContext } from '../platform/context'
import { asError, isErrorLike } from '../util/errors'
import { ExtensionManifest } from './extensionManifest'

export interface Controller extends Unsubscribable {
    /**
     * Global notification messages that should be displayed to the user, from the following sources:
     *
     * - window/showMessage notifications from extensions
     * - Errors thrown or returned in command invocation
     */
    readonly notifications: Observable<Notification>

    services: Services

    /**
     * Executes the command (registered in the CommandRegistry) specified in params. If an error is thrown, the
     * error is returned *and* emitted on the {@link Controller#notifications} observable.
     *
     * All callers should execute commands using this method instead of calling
     * {@link sourcegraph:CommandRegistry#executeCommand} directly (to ensure errors are emitted as notifications).
     *
     * @param suppressNotificationOnError By default, if command execution throws (or rejects with) an error, the
     * error will be shown in the global notification UI component. Pass suppressNotificationOnError as true to
     * skip this. The error is always returned to the caller.
     */
    executeCommand(params: ExecuteCommandParams, suppressNotificationOnError?: boolean): Promise<any>

    /**
     * Frees all resources associated with this client.
     */
    unsubscribe(): void
}

/**
 * React props or state containing the client. There should be only a single client for the whole
 * application.
 */
export interface ExtensionsControllerProps {
    /**
     * The client, which is used to communicate with and manage extensions.
     */
    extensionsController: Controller
}

/**
 * Creates the controller, which handles all communication between the client application and extensions.
 *
 * There should only be a single controller for the entire client application. The controller's model represents
 * all of the client application state that the client needs to know.
 */
export function createController(context: PlatformContext): Controller {
    const subscriptions = new Subscription()

    const services = new Services(context)
    const extensionHostConnection = combineLatest(
        context.createExtensionHost().pipe(
            switchMap(async messageTransports => {
                const connection = createConnection(messageTransports)
                connection.listen()

                const initData: InitData = {
                    sourcegraphURL: context.sourcegraphURL,
                    clientApplication: context.clientApplication,
                }
                await connection.sendRequest('initialize', [initData])
                return connection
            }),
            share()
        ),
        context.traceExtensionHostCommunication,
        context.unpackedExtensionURL
    ).pipe(
        tap(([connection, trace, unpackedExtensionURL]) => {
            connection.trace(trace ? new BrowserConsoleTracer('') : null)
            if (unpackedExtensionURL) {
                services.extensions.unpackedExtensionURL.next(unpackedExtensionURL)
            }
        }),
        map(([connection]) => connection),
        distinctUntilChanged()
    )
    const client = createExtensionHostClient(services, extensionHostConnection)
    subscriptions.add(client)

    const notifications = new Subject<Notification>()

    subscriptions.add(registerBuiltinClientCommands(services, context))
    subscriptions.add(registerExtensionContributions(services.contribution, services.extensions))

    // Show messages (that don't need user input) as global notifications.
    subscriptions.add(
        services.notifications.showMessages.subscribe(({ message, type }) => notifications.next({ message, type }))
    )
    subscriptions.add(
        services.notifications.progresses.subscribe(({ title, progress }) => {
            notifications.next({ message: title, progress, type: MessageType.Log })
        })
    )

    function messageFromExtension(message: string): string {
        return `From extension:\n\n${message}`
    }
    subscriptions.add(
        services.notifications.showMessageRequests.subscribe(({ message, actions, resolve }) => {
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
    )
    subscriptions.add(
        services.notifications.showInputs.subscribe(({ message, defaultValue, resolve }) =>
            resolve(prompt(messageFromExtension(message), defaultValue))
        )
    )

    // Print window/logMessage log messages to the browser devtools console.
    subscriptions.add(
        services.notifications.logMessages.subscribe(({ message }) => {
            log('info', 'EXT', message)
        })
    )

    // Debug helpers.
    const DEBUG = true
    if (DEBUG) {
        // Debug helper: log model changes.
        const LOG_MODEL = false
        if (LOG_MODEL) {
            subscriptions.add(services.model.model.subscribe(model => log('info', 'model', model)))
        }

        // Debug helpers: e.g., just run `sxservices` in devtools to get a reference to the services.
        ;(window as any).sxservices = services
        // This value is synchronously available because observable has an underlying BehaviorSubject source.
        subscriptions.add(services.model.model.subscribe(v => ((window as any).sxmodel = v)))
    }

    return {
        notifications,
        services,
        executeCommand: (params, suppressNotificationOnError) =>
            services.commands.executeCommand(params).catch(err => {
                if (!suppressNotificationOnError) {
                    notifications.next({
                        message: asError(err).message,
                        type: MessageType.Error,
                        source: params.command,
                    })
                }
                return Promise.reject(err)
            }),
        unsubscribe: () => subscriptions.unsubscribe(),
    }
}

function registerExtensionContributions(
    contributionRegistry: ContributionRegistry,
    { activeExtensions }: Pick<ExtensionsService, 'activeExtensions'>
): Unsubscribable {
    const contributions = from(activeExtensions).pipe(
        map(extensions =>
            extensions
                .map(({ manifest }) => manifest)
                .filter((manifest): manifest is ExtensionManifest => manifest !== null && !isErrorLike(manifest))
                .map(({ contributes }) => contributes)
                .filter((contributions): contributions is Contributions => !!contributions)
        )
    )
    return contributionRegistry.registerContributions({
        contributions,
    })
}

/** Prints a nicely formatted console log or error message. */
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
        '%c EXT %s %c',
        `font-weight:bold;background-color:${backgroundColor};color:${color}`,
        subject,
        'font-weight:normal;background-color:unset',
        message,
        other || ''
    )
}
