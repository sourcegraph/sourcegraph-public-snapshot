import 'cross-fetch/polyfill'

import { of, ReplaySubject } from 'rxjs'
import vscode from 'vscode'

import { proxySubscribable } from '@sourcegraph/shared/src/api/extension/api/common'
import polyfillEventSource from '@sourcegraph/shared/src/polyfills/vendor/eventSource'
import { fetchStreamSuggestions } from '@sourcegraph/shared/src/search/suggestions'

import { observeAuthenticatedUser } from './backend/authenticatedUser'
import { logEvent } from './backend/eventLogger'
import { initializeInstanceVersionNumber } from './backend/instanceVersion'
import { requestGraphQLFromVSCode } from './backend/requestGraphQl'
import { initializeSearchContexts } from './backend/searchContexts'
import { initializeSourcegraphSettings } from './backend/sourcegraphSettings'
import { createStreamSearch } from './backend/streamSearch'
import { initializeCodeSharingCommands } from './commands/initialize'
import { ExtensionCoreAPI } from './contract'
import { openSourcegraphUriCommand } from './file-system/commands'
import { initializeSourcegraphFileSystem } from './file-system/initialize'
import { SourcegraphUri } from './file-system/SourcegraphUri'
import { Event } from './graphql-operations'
import { accessTokenSetting } from './settings/accessTokenSetting'
import { endpointRequestHeadersSetting, endpointSetting, setEndpoint } from './settings/endpointSetting'
import { invalidateContextOnSettingsChange } from './settings/invalidation'
import { LocalStorageService, SELECTED_SEARCH_CONTEXT_SPEC_KEY } from './settings/LocalStorageService'
import { watchUninstall } from './settings/uninstall'
import { createVSCEStateMachine, VSCEQueryState } from './state'
import { focusSearchPanel, openSourcegraphLinks, registerWebviews, copySourcegraphLinks } from './webview/commands'
import { processOldToken, scretTokenKey, SourcegraphAuthProvider } from './webview/platform/AuthProvider'
/**
 * See CONTRIBUTING docs for the Architecture Diagram
 */
export async function activate(context: vscode.ExtensionContext): Promise<void> {
    const secretStorage = context.secrets
    // Move token from user setting to secret storage
    await processOldToken(secretStorage)
    // Register SourcegraphAuthProvider
    context.subscriptions.push(
        vscode.authentication.registerAuthenticationProvider(
            endpointSetting(),
            scretTokenKey,
            new SourcegraphAuthProvider(secretStorage)
        )
    )
    const session = await vscode.authentication.getSession(endpointSetting(), [], { createIfNone: false })
    const authenticatedUser = observeAuthenticatedUser(secretStorage)
    const initialInstanceURL = endpointSetting()
    const initialAccessToken = await secretStorage.get(scretTokenKey)
    const localStorageService = new LocalStorageService(context.globalState)
    const stateMachine = createVSCEStateMachine({ localStorageService })
    invalidateContextOnSettingsChange({ context, stateMachine })
    initializeSearchContexts({ localStorageService, stateMachine, context })
    const sourcegraphSettings = initializeSourcegraphSettings({ context })
    const editorTheme = vscode.ColorThemeKind[vscode.window.activeColorTheme.kind]
    const eventSourceType = initializeInstanceVersionNumber(localStorageService, initialAccessToken, initialInstanceURL)
    // Sets global `EventSource` for Node, which is required for streaming search.
    // Add custom headers to `EventSource` Authorization header when provided
    const customHeaders = endpointRequestHeadersSetting()
    polyfillEventSource(initialAccessToken ? { Authorization: `token ${initialAccessToken}`, ...customHeaders } : {})
    // For search panel webview to signal that it is ready for messages.
    // Replay subject with large buffer size just in case panels are opened in quick succession.
    const initializedPanelIDs = new ReplaySubject<string>(7)
    // Used to observe search box query state from sidebar
    const sidebarQueryStates = new ReplaySubject<VSCEQueryState>(1)
    // Use for file tree panel
    const { fs } = initializeSourcegraphFileSystem({ context, initialInstanceURL })
    // Use api endpoint for stream search
    const streamSearch = createStreamSearch({
        context,
        stateMachine,
        sourcegraphURL: `${initialInstanceURL}/.api`,
        session,
    })
    async function login(newtoken: string, newuri: string): Promise<void> {
        try {
            const newEndpoint = new URL(newuri)
            const newTokenKey = newEndpoint.hostname
            await secretStorage.store(newTokenKey, newtoken)
            await setEndpoint(newEndpoint.href)
            // stateMachine.emit({ type: 'sourcegraph_url_change' })
        } catch (error) {
            console.error(error)
        }
    }
    async function logout(): Promise<void> {
        await secretStorage.delete(scretTokenKey)
        await setEndpoint(undefined)
        extensionCoreAPI.reloadWindow()
    }
    const extensionCoreAPI: ExtensionCoreAPI = {
        panelInitialized: panelId => initializedPanelIDs.next(panelId),
        observeState: () => proxySubscribable(stateMachine.observeState()),
        observePanelQueryState: () => proxySubscribable(sidebarQueryStates.asObservable()),
        emit: event => stateMachine.emit(event),
        requestGraphQL: requestGraphQLFromVSCode,
        observeSourcegraphSettings: () => proxySubscribable(sourcegraphSettings.settings),
        // Debt: converting Promises into Observables for ease of use with
        // `useObservable` hook. Add `usePromise`s hook to fix.
        getAuthenticatedUser: () => proxySubscribable(authenticatedUser),
        getInstanceURL: () => proxySubscribable(of(initialInstanceURL)),
        openSourcegraphFile: (uri: string) => openSourcegraphUriCommand(fs, SourcegraphUri.parse(uri)),
        openLink: uri => openSourcegraphLinks(uri),
        copyLink: uri => copySourcegraphLinks(uri),
        getAccessToken: accessTokenSetting(context.secrets),
        removeAccessToken: () => logout(),
        setEndpointUri: (accessToken, uri) => login(accessToken, uri),
        reloadWindow: () => vscode.commands.executeCommand('workbench.action.reloadWindow'),
        focusSearchPanel,
        streamSearch,
        fetchStreamSuggestions: (query, sourcegraphURL) =>
            // Use api endpoint for stream search
            proxySubscribable(fetchStreamSuggestions(query, `${sourcegraphURL}/.api`)),
        setSelectedSearchContextSpec: spec => {
            stateMachine.emit({ type: 'set_selected_search_context_spec', spec })
            return localStorageService.setValue(SELECTED_SEARCH_CONTEXT_SPEC_KEY, spec)
        },
        setSidebarQueryState: sidebarQueryState => sidebarQueryStates.next(sidebarQueryState),
        getLocalStorageItem: key => localStorageService.getValue(key),
        setLocalStorageItem: (key: string, value: string) => localStorageService.setValue(key, value),
        logEvents: (variables: Event) => logEvent(variables),
        getEventSource: eventSourceType,
        getEditorTheme: editorTheme,
    }
    // Also initializes code intel.
    registerWebviews({
        context,
        extensionCoreAPI,
        initializedPanelIDs,
        sourcegraphSettings,
        fs,
        instanceURL: initialInstanceURL,
    })
    initializeCodeSharingCommands(context, eventSourceType, localStorageService)
    // Watch for uninstall to log uninstall event
    watchUninstall(eventSourceType, localStorageService)

    // Add Sourcegraph to workspace recommendations (disabled for now as it was reported to violate
    // VS Code's UX guidelines for notifications: https://code.visualstudio.com/api/ux-guidelines/notifications)
    // recommendSourcegraph(localStorageService).catch(() => {})
}
