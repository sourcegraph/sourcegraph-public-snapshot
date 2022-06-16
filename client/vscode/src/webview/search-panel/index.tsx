import '../platform/polyfills'

import React, { useMemo } from 'react'

import { ShortcutProvider } from '@slimsag/react-shortcuts'
import { VSCodeProgressRing } from '@vscode/webview-ui-toolkit/react'
import * as Comlink from 'comlink'
import { render } from 'react-dom'
import { MemoryRouter } from 'react-router'

import { wrapRemoteObservable } from '@sourcegraph/shared/src/api/client/api/common'
import {
    AnchorLink,
    setLinkComponent,
    useObservable,
    WildcardThemeContext,
    // This is the root Tooltip usage
    // eslint-disable-next-line no-restricted-imports
    DeprecatedTooltip,
} from '@sourcegraph/wildcard'

import { ExtensionCoreAPI } from '../../contract'
import { createEndpointsForWebToNode } from '../comlink/webviewEndpoint'
import { createPlatformContext, WebviewPageContext, WebviewPageProps } from '../platform/context'
import { adaptMonacoThemeToEditorTheme } from '../theming/monacoTheme'
import { adaptSourcegraphThemeToEditorTheme } from '../theming/sourcegraphTheme'

import { searchPanelAPI } from './api'
import { SearchHomeView } from './SearchHomeView'
import { SearchResultsView } from './SearchResultsView'

import './index.module.scss'

const vsCodeApi = window.acquireVsCodeApi()

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

const Main: React.FC<React.PropsWithChildren<unknown>> = () => {
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
        return <VSCodeProgressRing />
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

    // Render SearchHomeView until the user submits a search.
    if (state.context.submittedSearchQueryState === null) {
        return (
            <WebviewPageContext.Provider value={webviewPageProps}>
                <SearchHomeView {...webviewPageProps} context={state.context} />
            </WebviewPageContext.Provider>
        )
    }

    return (
        <WebviewPageContext.Provider value={webviewPageProps}>
            <SearchResultsView
                {...webviewPageProps}
                context={{
                    ...state.context,
                    submittedSearchQueryState: state.context.submittedSearchQueryState,
                }}
            />
        </WebviewPageContext.Provider>
    )
}

render(
    <ShortcutProvider>
        <WildcardThemeContext.Provider value={{ isBranded: true }}>
            {/* Required for shared components that depend on `location`. */}
            <MemoryRouter>
                <Main />
            </MemoryRouter>
            <DeprecatedTooltip key={1} className="sourcegraph-tooltip" />
        </WildcardThemeContext.Provider>
    </ShortcutProvider>,
    document.querySelector('#root')
)
