import * as comlink from 'comlink'
import { from, Subject, Subscription } from 'rxjs'
import { first } from 'rxjs/operators'
import { ContextValues, Progress, ProgressOptions, Unsubscribable } from 'sourcegraph'
import { PlatformContext, ClosableEndpointPair } from '../../platform/context'
import { ExtensionHostAPIFactory } from '../extension/api/api'
import { InitData } from '../extension/extensionHost'
import { ClientAPI } from './api/api'
import { createClientContent } from './api/content'
import { ClientContext } from './api/context'
import { ClientLanguageFeatures } from './api/languageFeatures'
import { ClientViews } from './api/views'
import { ClientWindows } from './api/windows'
import { Services } from './services'
import {
    MessageActionItem,
    ShowInputParams,
    ShowMessageRequestParams,
    ShowNotificationParams,
} from './services/notifications'
import { registerComlinkTransferHandlers } from '../util'
import { initMainThreadAPI } from './mainthread-api'
import { isSettingsValid } from '../../settings/settings'
import { FlatExtHostAPI } from '../contract'

export interface ExtensionHostClientConnection {
    /**
     * Closes the connection to and terminates the extension host.
     */
    unsubscribe(): void
}

/**
 * An activated extension.
 */
export interface ActivatedExtension {
    /**
     * The extension's extension ID (which uniquely identifies it among all activated extensions).
     */
    id: string

    /**
     * Deactivate the extension (by calling its "deactivate" function, if any).
     */
    deactivate(): void | Promise<void>
}

/**
 * @param endpoints The Worker object to communicate with
 */
export async function createExtensionHostClientConnection(
    endpointsPromise: Promise<ClosableEndpointPair>,
    services: Services,
    initData: Omit<InitData, 'initialSettings'>,
    platformContext: Pick<PlatformContext, 'settings' | 'updateSettings'>
): Promise<{ subscription: Unsubscribable; api: comlink.Remote<FlatExtHostAPI> }> {
    const subscription = new Subscription()

    // MAIN THREAD

    registerComlinkTransferHandlers()

    const endpoints = await endpointsPromise
    subscription.add(endpoints.subscription)

    /** Proxy to the exposed extension host API */
    const initializeExtensionHost = comlink.wrap<ExtensionHostAPIFactory>(endpoints.proxy)

    const initialSettings = await from(platformContext.settings).pipe(first()).toPromise()
    const proxy = await initializeExtensionHost({
        ...initData,
        // TODO what to do in error case?
        initialSettings: isSettingsValid(initialSettings) ? initialSettings : { final: {}, subjects: [] },
    })

    const clientContext = new ClientContext((updates: ContextValues) => services.context.updateContext(updates))
    subscription.add(clientContext)

    const clientWindows = new ClientWindows(
        (parameters: ShowNotificationParams) => services.notifications.showMessages.next({ ...parameters }),
        (parameters: ShowMessageRequestParams) =>
            new Promise<MessageActionItem | null>(resolve => {
                services.notifications.showMessageRequests.next({ ...parameters, resolve })
            }),
        (parameters: ShowInputParams) =>
            new Promise<string | null>(resolve => {
                services.notifications.showInputs.next({ ...parameters, resolve })
            }),
        ({ title }: ProgressOptions) => {
            const reporter = new Subject<Progress>()
            services.notifications.progresses.next({ title, progress: reporter.asObservable() })
            return reporter
        }
    )

    const clientViews = new ClientViews(
        services.panelViews,
        services.textDocumentLocations,
        services.viewer,
        services.view
    )

    const clientLanguageFeatures = new ClientLanguageFeatures(
        services.textDocumentDefinition,
        services.textDocumentReferences,
        services.textDocumentLocations,
        services.completionItems
    )

    const clientContent = createClientContent(services.linkPreviews)

    const { api: newAPI, subscription: apiSubscriptions } = initMainThreadAPI(proxy, platformContext, services)

    subscription.add(apiSubscriptions)

    const clientAPI: ClientAPI = {
        ping: () => 'pong',
        context: clientContext,
        languageFeatures: clientLanguageFeatures,
        windows: clientWindows,
        views: clientViews,
        content: clientContent,
        ...newAPI,
    }

    comlink.expose(clientAPI, endpoints.expose)

    return { subscription, api: proxy }
}
