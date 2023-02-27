import { Remote, proxy } from 'comlink'
import { Unsubscribable, Subscription, from, Observable, Subject, of } from 'rxjs'
import { publishReplay, refCount, switchMap } from 'rxjs/operators'

import { asError, logger } from '@sourcegraph/common'

import { InputBoxOptions } from '../../codeintel/legacy-extensions/api'
import { registerBuiltinClientCommands } from '../../commands/commands'
import { PlatformContext } from '../../platform/context'
import { isSettingsValid } from '../../settings/settings'
import { FlatExtensionHostAPI, MainThreadAPI } from '../contract'
import { proxySubscribable } from '../extension/api/common'
import { NotificationType, PlainNotification } from '../extension/extensionHostApi'

import { ProxySubscription } from './api/common'
import { updateSettings } from './services/settings'

/** A registered command in the command registry. */
export interface CommandEntry {
    /** The command ID (conventionally, e.g., "myextension.mycommand"). */
    command: string

    /** The function called to run the command and return an async value. */
    run: (...args: any[]) => Promise<any>
}

export interface ExecuteCommandParameters {
    /**
     * The identifier of the actual command handler.
     */
    command: string

    /**
     * Arguments that the command should be invoked with.
     */
    args?: any[]
}

function messageFromExtension(message: string): string {
    return `From extension:\n\n${message}`
}

/**
 * For state that needs to live in the main thread.
 * Returned to Controller for access by client applications.
 */
export interface ExposedToClient {
    registerCommand: (entryToRegister: CommandEntry) => Unsubscribable
    executeCommand: (parameters: ExecuteCommandParameters, suppressNotificationOnError?: boolean) => Promise<any>

    /**
     * Observable of error notifications as a result of client applications executing commands.
     */
    commandErrors: Observable<PlainNotification>
}

export const initMainThreadAPI = (
    extensionHost: Remote<FlatExtensionHostAPI>,
    platformContext: Pick<
        PlatformContext,
        | 'updateSettings'
        | 'settings'
        | 'getGraphQLClient'
        | 'requestGraphQL'
        | 'showMessage'
        | 'showInputBox'
        | 'getStaticExtensions'
        | 'telemetryService'
        | 'clientApplication'
    >
): { api: MainThreadAPI; exposedToClient: ExposedToClient; subscription: Subscription } => {
    const subscription = new Subscription()

    subscription.add(
        from(platformContext.settings)
            .pipe(
                switchMap(settings => {
                    if (isSettingsValid(settings)) {
                        return extensionHost.syncSettingsData(settings)
                    }
                    return []
                })
            )
            .subscribe()
    )

    // Commands
    const commands = new Map<string, CommandEntry>()
    const registerCommand = ({ command, run }: CommandEntry): Unsubscribable => {
        if (commands.has(command)) {
            throw new Error(`command is already registered: ${JSON.stringify(command)}`)
        }

        commands.set(command, { command, run })
        return {
            unsubscribe: () => commands.delete(command),
        }
    }
    const executeCommand = ({ args, command }: ExecuteCommandParameters): Promise<any> => {
        const commandEntry = commands.get(command)
        if (!commandEntry) {
            throw new Error(`command not found: ${JSON.stringify(command)}`)
        }
        return Promise.resolve(commandEntry.run(...(args || [])))
    }

    subscription.add(registerBuiltinClientCommands(platformContext, extensionHost, registerCommand))

    const commandErrors = new Subject<PlainNotification>()
    const exposedToClient: ExposedToClient = {
        registerCommand,
        executeCommand: (parameters, suppressNotificationOnError) =>
            executeCommand(parameters).catch(error => {
                if (!suppressNotificationOnError) {
                    commandErrors.next({
                        message: asError(error).message,
                        type: NotificationType.Error,
                        source: parameters.command,
                    })
                }
                return Promise.reject(error)
            }),
        commandErrors,
    }

    const api: MainThreadAPI = {
        applySettingsEdit: edit => updateSettings(platformContext, edit),
        requestGraphQL: (request, variables) =>
            platformContext
                .requestGraphQL({
                    request,
                    variables,
                    mightContainPrivateInfo: true,
                })
                .toPromise(),
        // Commands
        executeCommand: (command, args) => executeCommand({ command, args }),
        registerCommand: (command, run) => {
            const subscription = new Subscription()
            subscription.add(registerCommand({ command, run }))
            subscription.add(new ProxySubscription(run))
            return proxy(subscription)
        },
        // User interaction methods
        showMessage: message =>
            platformContext.showMessage ? platformContext.showMessage(message) : defaultShowMessage(message),
        showInputBox: options =>
            platformContext.showInputBox ? platformContext.showInputBox(options) : defaultShowInputBox(options),

        getEnabledExtensions: () => {
            if (platformContext.getStaticExtensions) {
                return proxySubscribable(
                    platformContext
                        .getStaticExtensions()
                        .pipe(
                            switchMap(staticExtensions =>
                                staticExtensions ? of(staticExtensions).pipe(publishReplay(1), refCount()) : of([])
                            )
                        )
                )
            }

            return proxySubscribable(of([]))
        },
        logEvent: (eventName, eventProperties) => platformContext.telemetryService?.log(eventName, eventProperties),
        logExtensionMessage: (...data) => logger.log(...data),
    }

    return { api, exposedToClient, subscription }
}

function defaultShowMessage(message: string): Promise<void> {
    return new Promise<void>(resolve => {
        alert(messageFromExtension(message))
        resolve()
    })
}

function defaultShowInputBox(options?: InputBoxOptions): Promise<string | undefined> {
    return new Promise<string | undefined>(resolve => {
        const response = prompt(messageFromExtension(options?.prompt ?? ''), options?.value)
        resolve(response ?? undefined)
    })
}
