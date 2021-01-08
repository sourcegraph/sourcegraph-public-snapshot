import { from, Subject, Subscription, Unsubscribable } from 'rxjs'
import { map, publishReplay, refCount } from 'rxjs/operators'
import { Services } from '../api/client/services'
import { ExecuteCommandParameters } from '../api/client/services/command'
import { ContributionRegistry, parseContributionExpressions } from '../api/client/services/contribution'
import { ExtensionsService } from '../api/client/services/extensionsService'
import { InitData } from '../api/extension/extensionHost'
import { registerBuiltinClientCommands } from '../commands/commands'
import { Notification } from '../notifications/notification'
import { PlatformContext } from '../platform/context'
import { asError, ErrorLike, isErrorLike } from '../util/errors'
import { isDefined } from '../util/types'
import { createExtensionHostClientConnection } from '../api/client/connection'
import { Remote } from 'comlink'
import { FlatExtensionHostAPI, NotificationType } from '../api/contract'
import { CommandEntry, MainThreadAPIDependencies } from '../api/client/mainthread-api'

export interface Controller extends Unsubscribable {
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
    executeCommand(parameters: ExecuteCommandParameters, suppressNotificationOnError?: boolean): Promise<any>

    registerCommand(entryToRegister: CommandEntry): Unsubscribable

    /**
     * Frees all resources associated with this client.
     */
    unsubscribe(): void

    extHostAPI: Promise<Remote<FlatExtensionHostAPI>>
}

/**
 * React props or state containing the client. There should be only a single client for the whole
 * application.
 */
export interface ExtensionsControllerProps<K extends keyof Controller = keyof Controller> {
    /**
     * The client, which is used to communicate with and manage extensions.
     */
    extensionsController: Pick<Controller, K>
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
    const initData: Omit<InitData, 'initialSettings'> = {
        sourcegraphURL: context.sourcegraphURL,
        clientApplication: context.clientApplication,
    }

    const commands = new Map<string, CommandEntry>()
    const mainThreadAPIDependencies: MainThreadAPIDependencies = {
        registerCommand: ({ command, run }) => {
            if (commands.has(command)) {
                throw new Error(`command is already registered: ${JSON.stringify(command)}`)
            }
            commands.set(command, { command, run })
            return {
                unsubscribe: () => commands.delete(command),
            }
        },
        executeCommand: ({ args, command }) => {
            const commandEntry = commands.get(command)
            if (!commandEntry) {
                throw new Error(`command not found: ${JSON.stringify(command)}`)
            }
            return Promise.resolve(commandEntry.run(...(args || [])))
        },
    }

    const extensionHostClientPromise = createExtensionHostClientConnection(
        context.createExtensionHost(),
        services,
        initData,
        context,
        // TOOD: replace services w/ this
        mainThreadAPIDependencies
    )

    subscriptions.add(() => extensionHostClientPromise.then(({ subscription }) => subscription.unsubscribe()))

    const notifications = new Subject<Notification>()

    subscriptions.add(registerBuiltinClientCommands(services, context, mainThreadAPIDependencies))
    subscriptions.add(registerExtensionContributions(services.contribution, services.extensions))

    // Debug helpers.
    const DEBUG = true
    if (DEBUG) {
        // Debug helper: log editor changes.
        const LOG_EDITORS = false
        if (LOG_EDITORS) {
            subscriptions.add(
                services.viewer.viewerUpdates.subscribe(() => log('info', 'editors', services.viewer.viewers))
            )
        }

        // Debug helpers: e.g., just run `sxservices` in devtools to get a reference to the services.
        ;(window as any).sxservices = services
    }

    return {
        services,
        executeCommand: (parameters, suppressNotificationOnError) =>
            mainThreadAPIDependencies.executeCommand(parameters).catch(error => {
                if (!suppressNotificationOnError) {
                    notifications.next({
                        message: asError(error).message,
                        type: NotificationType.Error,
                        source: parameters.command,
                    })
                }
                return Promise.reject(error)
            }),
        registerCommand: entryToRegister => mainThreadAPIDependencies.registerCommand(entryToRegister),
        extHostAPI: extensionHostClientPromise.then(({ api }) => api),
        unsubscribe: () => subscriptions.unsubscribe(),
    }
}

export function registerExtensionContributions(
    contributionRegistry: Pick<ContributionRegistry, 'registerContributions'>,
    { activeExtensions }: Pick<ExtensionsService, 'activeExtensions'>
): Unsubscribable {
    const contributions = from(activeExtensions).pipe(
        map(extensions =>
            extensions
                .map(({ manifest }) => manifest)
                .filter(
                    (manifest): manifest is Exclude<typeof manifest, ErrorLike | null> =>
                        manifest !== null && !isErrorLike(manifest)
                )
                .map(({ contributes }) => contributes)
                .filter(isDefined)
                .map(contributions => {
                    try {
                        // TODO(tj): looks like all contributions are parsed each time an extension is activated
                        return parseContributionExpressions(contributions)
                    } catch (error) {
                        // An error during evaluation causes all of the contributions in the same entry to be
                        // discarded.
                        console.warn('Discarding contributions: parsing expressions or templates failed.', {
                            contributions,
                            error,
                        })
                        return {}
                    }
                })
        ),
        // Perf optimization: only parse all the context expression once if there are multiple Subscribers.
        // This does not change the behaviour of the Observable, it always emits the current value on Subscription.
        publishReplay(1),
        refCount()
    )
    return contributionRegistry.registerContributions({ contributions })
}

/** Prints a nicely formatted console log or error message. */
function log(level: 'info' | 'error', subject: string, message: any, other?: { [name: string]: any }): void {
    let log: typeof console.log
    let color: string
    let backgroundColor: string
    if (level === 'info') {
        log = console.log.bind(console)
        color = '#000'
        backgroundColor = '#eee'
    } else {
        log = console.error.bind(console)
        color = 'white'
        backgroundColor = 'red'
    }
    log(
        '%c EXT %s %c',
        `font-weight:bold;background-color:${backgroundColor};color:${color}`,
        subject,
        'font-weight:normal;background-color:unset',
        message,
        other || ''
    )
}
