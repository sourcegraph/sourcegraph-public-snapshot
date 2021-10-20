import * as Comlink from 'comlink'
import React from 'react'
import { render } from 'react-dom'
import { BehaviorSubject } from 'rxjs'

import { useObservable } from '@sourcegraph/shared/src/util/useObservable'

import { SourcegraphVSCodeExtensionAPI, SourcegraphVSCodeWebviewAPI } from './contract'
import { createPlatformContext, VSCodePlatformContext } from './platform/context'
import { vsCodeWebviewEndpoint } from './platform/webviewEndpoint'
import { SearchPage } from './search'

declare global {
    // eslint-disable-next-line no-var
    var acquireVsCodeApi: () => VsCodeApi
}

export interface VsCodeApi {
    postMessage: (message: any) => void
}

// TODO: can probably use zustand for all state + serialization?
const routeSubject = new BehaviorSubject<'search' | undefined>(undefined)

const webviewAPI: SourcegraphVSCodeWebviewAPI = {
    setRoute: route => {
        routeSubject.next(route)
    },
}

const vsCodeApi = window.acquireVsCodeApi()

Comlink.expose(webviewAPI, vsCodeWebviewEndpoint(vsCodeApi, 'webview'))

const sourcegraphVSCodeExtensionAPI: Comlink.Remote<SourcegraphVSCodeExtensionAPI> = Comlink.wrap(
    vsCodeWebviewEndpoint(vsCodeApi, 'extension')
)

const platformContext = createPlatformContext(sourcegraphVSCodeExtensionAPI)

export interface WebviewPageProps {
    sourcegraphVSCodeExtensionAPI: Comlink.Remote<SourcegraphVSCodeExtensionAPI>
    platformContext: VSCodePlatformContext
}

const Main: React.FC = () => {
    const route = useObservable(routeSubject)

    const commonPageProps: WebviewPageProps = {
        sourcegraphVSCodeExtensionAPI,
        platformContext,
    }

    if (route === undefined) {
        return <p>Loading...</p>
    }

    return <SearchPage {...commonPageProps} />
}
render(<Main />, document.querySelector('#root'))
