import * as comlink from 'comlink'
import { from, Subscription } from 'rxjs'
import { first } from 'rxjs/operators'
import { ContextValues, Unsubscribable } from 'sourcegraph'
import { PlatformContext, ClosableEndpointPair } from '../../platform/context'
import { ExtensionHostAPIFactory } from '../extension/api/api'
import { InitData } from '../extension/extensionHost'
import { ClientAPI } from './api/api'
import { ClientCodeEditor } from './api/codeEditor'
import { createClientContent } from './api/content'
import { ClientContext } from './api/context'
import { ClientExtensions } from './api/extensions'
import { ClientLanguageFeatures } from './api/languageFeatures'
import { ClientViews } from './api/views'
import { Services } from './services'
import { registerComlinkTransferHandlers } from '../util'
import { initMainThreadAPI, MainThreadAPIDependencies } from './mainthread-api'
import { isSettingsValid } from '../../settings/settings'
import { FlatExtensionHostAPI, MainThreadAPI } from '../contract'

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
    platformContext: Pick<PlatformContext, 'settings' | 'updateSettings' | 'requestGraphQL' | 'telemetryService'>,
    mainThreadAPIDependences: MainThreadAPIDependencies
): Promise<{ subscription: Unsubscribable; api: comlink.Remote<FlatExtensionHostAPI>; mainThreadAPI: MainThreadAPI }> {
    const subscription = new Subscription()

    // MAIN THREAD

    registerComlinkTransferHandlers()

    const { endpoints, subscription: endpointsSubscription } = await endpointsPromise
    subscription.add(endpointsSubscription)

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

    const clientViews = new ClientViews(services.panelViews, services.textDocumentLocations, services.view, proxy)

    const clientCodeEditor = new ClientCodeEditor(services.textDocumentDecoration)
    subscription.add(clientCodeEditor)

    const clientLanguageFeatures = new ClientLanguageFeatures(
        services.textDocumentReferences,
        services.textDocumentLocations,
        services.completionItems
    )
    subscription.add(new ClientExtensions(proxy.extensions, services.extensions, platformContext))

    const clientContent = createClientContent(services.linkPreviews)

    const { api: newAPI, subscription: apiSubscriptions } = initMainThreadAPI(
        proxy,
        platformContext,
        mainThreadAPIDependences
    )

    subscription.add(apiSubscriptions)

    const clientAPI: ClientAPI = {
        ping: () => 'pong',
        context: clientContext,
        languageFeatures: clientLanguageFeatures,
        codeEditor: clientCodeEditor,
        views: clientViews,
        content: clientContent,
        ...newAPI,
    }

    comlink.expose(clientAPI, endpoints.expose)

    // TODO(tj): return MainThreadAPI and add to Controller interface
    // to allow app to interact with APIs whose state lives in the main thread
    return { subscription, api: proxy, mainThreadAPI: newAPI }
}
