import * as Comlink from 'comlink'
import React, { useMemo } from 'react'
import { render } from 'react-dom'
import { of } from 'rxjs'

import { wrapRemoteObservable } from '@sourcegraph/shared/src/api/client/api/common'
import { proxySubscribable } from '@sourcegraph/shared/src/api/extension/api/common'
import { AnchorLink, setLinkComponent, useObservable } from '@sourcegraph/wildcard'

import { ExtensionCoreAPI, SearchSidebarAPI } from '../../contract'
import { createEndpointsForWebToNode } from '../comlink/webviewEndpoint'

// TODO: load extension host

const vsCodeApi = window.acquireVsCodeApi()

const searchSidebarAPI: SearchSidebarAPI = {
    ping: () => {
        console.log('ping called')
        return proxySubscribable(of('pong'))
    },
    addTextDocumentIfNotExists: () => {
        console.log('addTextDocumentIfNotExists called')
    },
}

const { proxy, expose } = createEndpointsForWebToNode(vsCodeApi)

Comlink.expose(searchSidebarAPI, expose)

export const extensionCoreAPI: Comlink.Remote<ExtensionCoreAPI> = Comlink.wrap(proxy)

setLinkComponent(AnchorLink)

const Main: React.FC = () => {
    console.log('rendering webview')

    const state = useObservable(useMemo(() => wrapRemoteObservable(extensionCoreAPI.observeState()), []))

    return (
        <div>
            <h1>state: {state?.status}</h1>
        </div>
    )
}

render(<Main />, document.querySelector('#root'))
