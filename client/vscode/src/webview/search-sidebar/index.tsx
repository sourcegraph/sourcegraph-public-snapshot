import '../platform/polyfills'

import * as Comlink from 'comlink'
import React from 'react'
import { render } from 'react-dom'

import { AnchorLink, setLinkComponent } from '@sourcegraph/shared/src/components/Link'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'

import { SourcegraphVSCodeExtensionAPI, SourcegraphVSCodeSearchSidebarAPI } from '../contract'
import { createPlatformContext, WebviewPageProps } from '../platform/context'
import { createEndpointsForWebToNode } from '../platform/webviewEndpoint'
import { adaptToEditorTheme } from '../theme'

import { SearchSidebar } from './SearchSidebar'

const vsCodeApi = window.acquireVsCodeApi()

const { proxy, expose } = createEndpointsForWebToNode(vsCodeApi)

const webviewAPI: SourcegraphVSCodeSearchSidebarAPI = {}

Comlink.expose(webviewAPI, expose)

const sourcegraphVSCodeExtensionAPI: Comlink.Remote<SourcegraphVSCodeExtensionAPI> = Comlink.wrap(proxy)

const platformContext = createPlatformContext(sourcegraphVSCodeExtensionAPI)

setLinkComponent(AnchorLink)

// Get theme
const themes = adaptToEditorTheme()

const Main: React.FC = () => {
    const theme = useObservable(themes) || 'theme-dark'

    const commonPageProps: WebviewPageProps = {
        sourcegraphVSCodeExtensionAPI,
        platformContext,
        theme,
    }

    return <SearchSidebar {...commonPageProps} />
}

render(<Main />, document.querySelector('#root'))
