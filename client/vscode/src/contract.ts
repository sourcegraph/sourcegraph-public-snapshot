import { GraphQLResult } from '@sourcegraph/http-client'
import { FlatExtensionHostAPI } from '@sourcegraph/shared/src/api/contract'
import { ProxySubscribable } from '@sourcegraph/shared/src/api/extension/api/common'
import { SettingsCascadeOrError } from '@sourcegraph/shared/src/settings/settings'

import { VSCEState, VSCEStateMachine } from './state'

export interface ExtensionCoreAPI {
    /** For search panel webview to signal that it is ready for messages. */
    panelInitialized: (panelId: string) => void

    requestGraphQL: (request: string, variables: any) => Promise<GraphQLResult<any>>
    observeSourcegraphSettings: () => ProxySubscribable<SettingsCascadeOrError>

    observeState: () => ProxySubscribable<VSCEState>
    emit: VSCEStateMachine['emit']
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
