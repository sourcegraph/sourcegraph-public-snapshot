import { FlatExtensionHostAPI } from '@sourcegraph/shared/src/api/contract'
import { ProxySubscribable } from '@sourcegraph/shared/src/api/extension/api/common'

import { VSCEState } from './state'

export interface ExtensionCoreAPI {
    // TODO remove once other methods are implemented
    ping: () => ProxySubscribable<'pong'>

    /** For search panel webview to signal that it is ready for messages. */
    panelInitialized: (panelId: string) => void

    // TODO check if this is still necessary in follow-up PR
    // requestGraphQL (to be used by PlatformContext)
    // observeSourcegraphSettings

    observeState: () => ProxySubscribable<VSCEState>
}

export interface SearchPanelAPI {
    // TODO remove once other methods are implemented
    ping: () => ProxySubscribable<'pong'>
}

export interface SearchSidebarAPI extends Pick<FlatExtensionHostAPI, 'addTextDocumentIfNotExists'> {
    // TODO remove once other methods are implemented
    ping: () => ProxySubscribable<'pong'>
    // TODO: ExtensionHostAPI methods
}
