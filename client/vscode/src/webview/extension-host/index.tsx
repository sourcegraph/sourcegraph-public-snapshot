import '../platform/polyfills'

import * as Comlink from 'comlink'
import React from 'react'
import { render } from 'react-dom'
import { from } from 'rxjs'
import { switchMap } from 'rxjs/operators'

import { wrapRemoteObservable } from '@sourcegraph/shared/src/api/client/api/common'
import { proxySubscribable } from '@sourcegraph/shared/src/api/extension/api/common'
import { AnchorLink, setLinkComponent } from '@sourcegraph/shared/src/components/Link'
import { createController as createExtensionsController } from '@sourcegraph/shared/src/extensions/controller'

import { SourcegraphVSCodeExtensionAPI, SourcegraphVSCodeExtensionHostAPI } from '../contract'
import { createPlatformContext } from '../platform/context'
import { createEndpointsForWebToNode } from '../platform/webviewEndpoint'

import { createExtensionHost } from './worker'

const vsCodeApi = window.acquireVsCodeApi()

const webviewAPI: SourcegraphVSCodeExtensionHostAPI = {
    getDefinition: parameters => {
        const definitions = from(extensionsController.extHostAPI).pipe(
            switchMap(extensionHostAPI => wrapRemoteObservable(extensionHostAPI.getDefinition(parameters)))
        )

        return proxySubscribable(definitions)
    },
    getReferences: (parameters, referenceContext) => {
        const references = from(extensionsController.extHostAPI).pipe(
            switchMap(extensionHostAPI =>
                wrapRemoteObservable(extensionHostAPI.getReferences(parameters, referenceContext))
            )
        )

        return proxySubscribable(references)
    },
    getHover: parameters => {
        const hovers = from(extensionsController.extHostAPI).pipe(
            switchMap(extensionHostAPI => wrapRemoteObservable(extensionHostAPI.getHover(parameters)))
        )

        return proxySubscribable(hovers)
    },

    addTextDocumentIfNotExists: textDocumentData =>
        extensionsController.extHostAPI.then(extensionHostAPI =>
            extensionHostAPI.addTextDocumentIfNotExists(textDocumentData)
        ),
    addViewerIfNotExists: viewer =>
        extensionsController.extHostAPI.then(extensionHostAPI => extensionHostAPI.addViewerIfNotExists(viewer)),
}

const { proxy, expose } = createEndpointsForWebToNode(vsCodeApi)

Comlink.expose(webviewAPI, expose)

const sourcegraphVSCodeExtensionAPI: Comlink.Remote<SourcegraphVSCodeExtensionAPI> = Comlink.wrap(proxy)

const platformContext = createPlatformContext(sourcegraphVSCodeExtensionAPI)

const sourcegraphURL = document.documentElement.dataset.instanceUrl!

console.log('in sg exthost', { sourcegraphURL })

// TODO get sourcegraphURL
const extensionsController = createExtensionsController({
    ...platformContext,
    sourcegraphURL,
    createExtensionHost: () => Promise.resolve(createExtensionHost()),
})

setLinkComponent(AnchorLink)

const Main: React.FC = () => <div />
render(<Main />, document.querySelector('#root'))
