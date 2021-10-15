import * as Comlink from 'comlink'
import React from 'react'
import { render } from 'react-dom'

import { AnchorLink, setLinkComponent } from '@sourcegraph/shared/src/components/Link'

import { SourcegraphVSCodeExtensionAPI, SourcegraphVSCodeSearchSidebarAPI } from '../contract'
import { createPlatformContext } from '../platform/context'
import { createEndpoints } from '../platform/webviewEndpoint'
import { adaptToEditorTheme } from '../theme'

import { SearchSidebar } from './SearchSidebar'

const vsCodeApi = window.acquireVsCodeApi()

const { proxy, expose } = createEndpoints(vsCodeApi)

const webviewAPI: SourcegraphVSCodeSearchSidebarAPI = {}

Comlink.expose(webviewAPI, expose)

const sourcegraphVSCodeExtensionAPI: Comlink.Remote<SourcegraphVSCodeExtensionAPI> = Comlink.wrap(proxy)

const platformContext = createPlatformContext(sourcegraphVSCodeExtensionAPI)

setLinkComponent(AnchorLink)

// eslint-disable-next-line rxjs/no-ignored-observable
adaptToEditorTheme()

const Main: React.FC = () => (
    <SearchSidebar platformContext={platformContext} sourcegraphVSCodeExtensionAPI={sourcegraphVSCodeExtensionAPI} />
)

render(<Main />, document.querySelector('#root'))
