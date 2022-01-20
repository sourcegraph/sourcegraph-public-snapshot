import * as Comlink from 'comlink'
import React from 'react'
import { render } from 'react-dom'
import { of } from 'rxjs'

import { proxySubscribable } from '@sourcegraph/shared/src/api/extension/api/common'
import { AnchorLink, setLinkComponent } from '@sourcegraph/shared/src/components/Link'

import { ExtensionCoreAPI, SearchPanelAPI } from '../../contract'
import { createEndpointsForWebToNode } from '../comlink/webviewEndpoint'

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

extensionCoreAPI.panelInitialized(document.documentElement.dataset.panelId!).catch(() => {
    // noop (TODO?)
})

// TODO create platform context.

setLinkComponent(AnchorLink)

const Main: React.FC = () => {
    console.log('rendering webview')
    return (
        <div>
            <h1>WIP</h1>
        </div>
    )
}

render(<Main />, document.querySelector('#root'))
