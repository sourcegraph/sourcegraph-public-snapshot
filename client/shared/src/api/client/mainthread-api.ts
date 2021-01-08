import { Remote, proxy } from 'comlink'
import { updateSettings } from './services/settings'
import { Subscription, from } from 'rxjs'
import { PlatformContext } from '../../platform/context'
import { isSettingsValid } from '../../settings/settings'
import { switchMap } from 'rxjs/operators'
import { FlatExtensionHostAPI, MainThreadAPI } from '../contract'
import { ProxySubscription } from './api/common'
import * as sourcegraph from 'sourcegraph'

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
 * For state that needs to live in the main thread. Resides in the controller
 * for synchronous access by client applications.
 */
export interface MainThreadAPIDependencies {
    registerCommand: (entryToRegister: CommandEntry) => sourcegraph.Unsubscribable
    executeCommand: (parameters: ExecuteCommandParameters) => Promise<any>
}

export const initMainThreadAPI = (
    extensionHost: Remote<FlatExtensionHostAPI>,
    platformContext: Pick<
        PlatformContext,
        'updateSettings' | 'settings' | 'requestGraphQL' | 'showMessage' | 'showInputBox'
    >,
    mainThreadAPIDependences: MainThreadAPIDependencies
): { api: MainThreadAPI; subscription: Subscription } => {
    const subscription = new Subscription()
    // Settings
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
        executeCommand: (command, args) => mainThreadAPIDependences.executeCommand({ command, args }),
        registerCommand: (command, run) => {
            const subscription = new Subscription()
            subscription.add(mainThreadAPIDependences.registerCommand({ command, run }))
            subscription.add(new ProxySubscription(run))
            return proxy(subscription)
        },
        // User interaction methods
        showMessage: message =>
            platformContext.showMessage ? platformContext.showMessage(message) : defaultShowMessage(message),
        showInputBox: options =>
            platformContext.showInputBox ? platformContext.showInputBox(options) : defaultShowInputBox(options),
    }

    return { api, subscription }
}

function defaultShowMessage(message: string): Promise<void> {
    return new Promise<void>(resolve => {
        alert(messageFromExtension(message))
        resolve()
    })
}

function defaultShowInputBox(options?: sourcegraph.InputBoxOptions): Promise<string | undefined> {
    return new Promise<string | undefined>(resolve => {
        const response = prompt(messageFromExtension(options?.prompt ?? ''), options?.value)
        resolve(response ?? undefined)
    })
}
