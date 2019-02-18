import * as comlink from '@sourcegraph/comlink'
import { isEqual } from 'lodash'
import { from, Subject, Subscription } from 'rxjs'
import { distinctUntilChanged, map } from 'rxjs/operators'
import { ContextValues, Progress, ProgressOptions, Unsubscribable } from 'sourcegraph'
import { EndpointPair } from '../../platform/context'
import { ExtensionHostAPIFactory } from '../extension/api/api'
import { InitData } from '../extension/extensionHost'
import { ClientAPI } from './api/api'
import { ClientCodeEditor } from './api/codeEditor'
import { ClientCommands } from './api/commands'
import { ClientConfiguration } from './api/configuration'
import { ClientContext } from './api/context'
import { ClientDocuments } from './api/documents'
import { ClientExtensions } from './api/extensions'
import { ClientLanguageFeatures } from './api/languageFeatures'
import { ClientRoots } from './api/roots'
import { ClientSearch } from './api/search'
import { ClientViews } from './api/views'
import { ClientWindows } from './api/windows'
import { applyContextUpdate } from './context/context'
import { Services } from './services'
import {
    MessageActionItem,
    ShowInputParams,
    ShowMessageParams,
    ShowMessageRequestParams,
} from './services/notifications'

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
 * @param endpoint The Worker object to communicate with
 */
export async function createExtensionHostClientConnection(
    endpoints: EndpointPair,
    services: Services,
    initData: InitData
): Promise<Unsubscribable> {
    const subscription = new Subscription()

    // MAIN THREAD

    /** Proxy to the exposed extension host API */
    const initializeExtensionHost = comlink.proxy<ExtensionHostAPIFactory>(endpoints.proxy)
    const proxy = await initializeExtensionHost(initData)

    const clientConfiguration = new ClientConfiguration<any>(proxy.configuration, services.settings)
    subscription.add(clientConfiguration)

    const clientContext = new ClientContext((updates: ContextValues) =>
        services.context.data.next(applyContextUpdate(services.context.data.value, updates))
    )
    subscription.add(clientContext)
    const clientWindows = new ClientWindows(
        proxy.windows,
        from(services.model.model).pipe(
            map(({ visibleViewComponents }) => visibleViewComponents),
            distinctUntilChanged()
        ),
        (params: ShowMessageParams) => services.notifications.showMessages.next({ ...params }),
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
    subscription.add(clientWindows)

    const clientViews = new ClientViews(services.views, services.textDocumentLocations)

    const clientCodeEditor = new ClientCodeEditor(services.textDocumentDecoration)
    subscription.add(clientCodeEditor)
    subscription.add(
        new ClientDocuments(
            proxy.documents,
            from(services.model.model).pipe(
                map(
                    ({ visibleViewComponents }) =>
                        visibleViewComponents && visibleViewComponents.map(({ item }) => item)
                ),
                distinctUntilChanged((a, b) => isEqual(a, b))
            )
        )
    )
    const clientLanguageFeatures = new ClientLanguageFeatures(
        services.textDocumentHover,
        services.textDocumentDefinition,
        services.textDocumentTypeDefinition,
        services.textDocumentImplementation,
        services.textDocumentReferences,
        services.textDocumentLocations
    )
    const clientSearch = new ClientSearch(services.queryTransformer)
    const clientCommands = new ClientCommands(services.commands)
    subscription.add(
        new ClientRoots(
            proxy.roots,
            from(services.model.model).pipe(
                map(({ roots }) => roots),
                distinctUntilChanged()
            )
        )
    )
    subscription.add(new ClientExtensions(proxy.extensions, services.extensions))

    const clientAPI: ClientAPI = {
        ping: () => 'pong',
        context: clientContext,
        search: clientSearch,
        configuration: clientConfiguration,
        languageFeatures: clientLanguageFeatures,
        commands: clientCommands,
        windows: clientWindows,
        codeEditor: clientCodeEditor,
        views: clientViews,
    }
    comlink.expose(clientAPI, endpoints.expose)

    return subscription
}
