import { ShortcutProvider } from '@slimsag/react-shortcuts'
import * as Comlink from 'comlink'
import React, { useMemo } from 'react'
import { render } from 'react-dom'
import { of } from 'rxjs'

import { wrapRemoteObservable } from '@sourcegraph/shared/src/api/client/api/common'
import { proxySubscribable } from '@sourcegraph/shared/src/api/extension/api/common'
import {
    AnchorLink,
    LoadingSpinner,
    setLinkComponent,
    useObservable,
    WildcardThemeContext,
} from '@sourcegraph/wildcard'

import { ExtensionCoreAPI, SearchPanelAPI } from '../../contract'
import { createEndpointsForWebToNode } from '../comlink/webviewEndpoint'
import { createPlatformContext, WebviewPageProps } from '../platform/context'
import { adaptMonacoThemeToEditorTheme } from '../theming/monacoTheme'
import { adaptSourcegraphThemeToEditorTheme } from '../theming/sourcegraphTheme'

import { SearchHomeView } from './SearchHomeView'
import { SearchResultsView } from './SearchResultsView'

import './index.module.scss'

const vsCodeApi = window.acquireVsCodeApi()

const searchPanelAPI: SearchPanelAPI = {
    ping: () => {
        console.log('ping called')
        return proxySubscribable(of('pong'))
    },
}

const { proxy, expose } = createEndpointsForWebToNode(vsCodeApi)

Comlink.expose(searchPanelAPI, expose)

export const extensionCoreAPI: Comlink.Remote<ExtensionCoreAPI> = Comlink.wrap(proxy)

const themes = adaptSourcegraphThemeToEditorTheme()
adaptMonacoThemeToEditorTheme()

extensionCoreAPI.panelInitialized(document.documentElement.dataset.panelId!).catch(() => {
    // noop (TODO?)
})

const platformContext = createPlatformContext(extensionCoreAPI)

setLinkComponent(AnchorLink)

const Main: React.FC = () => {
    const state = useObservable(useMemo(() => wrapRemoteObservable(extensionCoreAPI.observeState()), []))

    const authenticatedUser = useObservable(
        useMemo(() => wrapRemoteObservable(extensionCoreAPI.getAuthenticatedUser()), [])
    )

    const instanceURL = useObservable(useMemo(() => wrapRemoteObservable(extensionCoreAPI.getInstanceURL()), []))

    const theme = useObservable(themes)

    const settingsCascade = useObservable(
        useMemo(() => wrapRemoteObservable(extensionCoreAPI.observeSourcegraphSettings()), [])
    )
    // Do not block rendering on settings unless we observe UI jitter

    // TODO: If init is taking too long, show a message.
    // Also check if anything has errored out.

    // If any of the remote values have yet to load.
    const initialized =
        state !== undefined &&
        authenticatedUser !== undefined &&
        instanceURL !== undefined &&
        theme !== undefined &&
        settingsCascade !== undefined

    if (!initialized) {
        return <LoadingSpinner />
    }

    const webviewPageProps: WebviewPageProps = {
        extensionCoreAPI,
        platformContext,
        theme,
        authenticatedUser,
        settingsCascade,
        instanceURL,
    }

    if (state?.status === 'context-invalidated') {
        // TODO context-invalidated state
        return null
    }

    // Idle and remote browsing state should be same as search results.
    if (state?.status === 'search-home') {
        return <SearchHomeView {...webviewPageProps} context={state.context} />
    }
    if (state?.status === 'search-results') {
        return <SearchResultsView {...webviewPageProps} context={state.context} />
    }
    // If state is remote browsing but the search panel is still visible in a different column,
    // we should still show results. Determine state by whether submittedSearchQuery is not null.
    if (state.context.submittedSearchQueryState !== null) {
        return (
            <SearchResultsView
                {...webviewPageProps}
                context={{
                    ...state.context,
                    submittedSearchQueryState: state.context.submittedSearchQueryState,
                }}
            />
        )
    }

    return null
}

render(
    <ShortcutProvider>
        <WildcardThemeContext.Provider value={{ isBranded: true }}>
            <Main />
        </WildcardThemeContext.Provider>
    </ShortcutProvider>,
    document.querySelector('#root')
)
