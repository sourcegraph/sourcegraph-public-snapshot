import * as Comlink from 'comlink'
import React, { useMemo } from 'react'
import { render } from 'react-dom'

import { wrapRemoteObservable } from '@sourcegraph/shared/src/api/client/api/common'
import {
    AnchorLink,
    LoadingSpinner,
    setLinkComponent,
    useObservable,
    WildcardThemeContext,
} from '@sourcegraph/wildcard'

import { ExtensionCoreAPI } from '../../contract'
import { createEndpointsForWebToNode } from '../comlink/webviewEndpoint'
import { createPlatformContext, WebviewPageProps } from '../platform/context'
import { adaptToEditorTheme } from '../theme'

import { createSearchSidebarAPI } from './api'
import { AuthSidebarView } from './AuthSidebarView'
import { ContextInvalidatedSidebarView } from './ContextInvalidatedSidebarView'

// TODO: load extension host

const vsCodeApi = window.acquireVsCodeApi()

const { proxy, expose } = createEndpointsForWebToNode(vsCodeApi)

const searchSidebarAPI = createSearchSidebarAPI()
Comlink.expose(searchSidebarAPI, expose)
export const extensionCoreAPI: Comlink.Remote<ExtensionCoreAPI> = Comlink.wrap(proxy)

const platformContext = createPlatformContext(extensionCoreAPI)

setLinkComponent(AnchorLink)

const themes = adaptToEditorTheme()

const Main: React.FC = () => {
    const state = useObservable(useMemo(() => wrapRemoteObservable(extensionCoreAPI.observeState()), []))

    const authenticatedUser = useObservable(
        useMemo(() => wrapRemoteObservable(extensionCoreAPI.getAuthenticatedUser()), [])
    )

    const instanceURL = useObservable(useMemo(() => wrapRemoteObservable(extensionCoreAPI.getInstanceURL()), []))

    const theme = useObservable(themes)

    // If any of the remote values have yet to load.
    const initialized =
        state !== undefined && authenticatedUser !== undefined && instanceURL !== undefined && theme !== undefined

    if (!initialized) {
        return <LoadingSpinner />
    }

    const webviewPageProps: WebviewPageProps = {
        extensionCoreAPI,
        platformContext,
        theme,
        instanceURL,
    }

    if (state.status === 'context-invalidated') {
        return <ContextInvalidatedSidebarView {...webviewPageProps} />
    }

    // TODO: should we hide the access token form permanently if an unauthenticated user
    // has performed a search before? Or just for this session?
    if (state.status === 'search-home' && !authenticatedUser) {
        return (
            <WildcardThemeContext.Provider value={{ isBranded: true }}>
                <AuthSidebarView {...webviewPageProps} />
            </WildcardThemeContext.Provider>
        )
    }

    if (state.status === 'remote-browsing') {
        // TODO files sidebar
    }

    if (state.status === 'idle') {
        // Search sidebar?
    }

    return (
        <div>
            <h1>state: {state?.status}</h1>
        </div>
    )
}

render(<Main />, document.querySelector('#root'))
