import * as comlink from 'comlink'
import { from, merge, Subject, Subscription, of } from 'rxjs'
import { concatMap } from 'rxjs/operators'
import { ContextValues, Progress, ProgressOptions, Unsubscribable } from 'sourcegraph'
import { EndpointPair, PlatformContext } from '../../platform/context'
import { ExtensionHostAPIFactory } from '../extension/api/api'
import { InitData } from '../extension/extensionHost'
import { ClientAPI } from './api/api'
import { ClientCodeEditor } from './api/codeEditor'
import { createClientContent } from './api/content'
import { ClientContext } from './api/context'
import { ClientExtensions } from './api/extensions'
import { ClientLanguageFeatures } from './api/languageFeatures'
import { ClientSearch } from './api/search'
import { ClientViews } from './api/views'
import { ClientWindows } from './api/windows'
import { Services } from './services'
import {
    MessageActionItem,
    ShowInputParams,
    ShowMessageRequestParams,
    ShowNotificationParams,
} from './services/notifications'
import { TextModelUpdate } from './services/modelService'
import { ViewerUpdate } from './services/viewerService'
import { registerComlinkTransferHandlers } from '../util'
import { initMainThreadAPI } from './mainthread-api'

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
    endpoints: EndpointPair,
    services: Services,
    initData: InitData,
    platformContext: Pick<PlatformContext, 'settings' | 'updateSettings'>
): Promise<Unsubscribable> {
    const subscription = new Subscription()

    // MAIN THREAD

    registerComlinkTransferHandlers()

    /** Proxy to the exposed extension host API */
    const initializeExtensionHost = comlink.wrap<ExtensionHostAPIFactory>(endpoints.proxy)
    const proxy = await initializeExtensionHost(initData)

    const clientContext = new ClientContext((updates: ContextValues) => services.context.updateContext(updates))
    subscription.add(clientContext)

    // Sync models and viewers to the extension host
    subscription.add(
        merge(
            of([...services.model.models.entries()].map(([, model]): TextModelUpdate => ({ type: 'added', ...model }))),
            from(services.model.modelUpdates)
        )
            .pipe(concatMap(modelUpdates => proxy.documents.$acceptDocumentData(modelUpdates)))
            .subscribe()
    )
    subscription.add(
        merge(
            of(
                [...services.viewer.viewers.entries()].map(
                    ([viewerId, viewerData]): ViewerUpdate => ({
                        type: 'added',
                        viewerId,
                        viewerData,
                    })
                )
            ),
            from(services.viewer.viewerUpdates)
        )
            .pipe(concatMap(viewerUpdates => proxy.windows.$acceptWindowData(viewerUpdates)))
            .subscribe()
    )

    const clientWindows = new ClientWindows(
        (params: ShowNotificationParams) => services.notifications.showMessages.next({ ...params }),
        (params: ShowMessageRequestParams) =>
            new Promise<MessageActionItem | null>(resolve => {
                services.notifications.showMessageRequests.next({ ...params, resolve })
            }),
        (params: ShowInputParams) =>
            new Promise<string | null>(resolve => {
                services.notifications.showInputs.next({ ...params, resolve })
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

    const clientCodeEditor = new ClientCodeEditor(services.textDocumentDecoration)
    subscription.add(clientCodeEditor)

    const clientLanguageFeatures = new ClientLanguageFeatures(
        services.textDocumentHover,
        services.textDocumentDefinition,
        services.textDocumentReferences,
        services.textDocumentLocations,
        services.completionItems
    )
    const clientSearch = new ClientSearch(services.queryTransformer)
    subscription.add(new ClientExtensions(proxy.extensions, services.extensions))

    const clientContent = createClientContent(services.linkPreviews)

    const { api: newAPI, subscription: apiSubscriptions } = initMainThreadAPI(proxy, platformContext, services)

    subscription.add(apiSubscriptions)

    const clientAPI: ClientAPI = {
        ping: () => 'pong',
        context: clientContext,
        search: clientSearch,
        languageFeatures: clientLanguageFeatures,
        windows: clientWindows,
        codeEditor: clientCodeEditor,
        views: clientViews,
        content: clientContent,
        ...newAPI,
    }
    comlink.expose(clientAPI, endpoints.expose)

    return subscription
}
