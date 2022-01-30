import { GraphQLResult } from '@sourcegraph/http-client'
import { FlatExtensionHostAPI } from '@sourcegraph/shared/src/api/contract'
import { ProxySubscribable } from '@sourcegraph/shared/src/api/extension/api/common'
import { AuthenticatedUser } from '@sourcegraph/shared/src/auth'
import { SearchMatch, StreamSearchOptions } from '@sourcegraph/shared/src/search/stream'
import { SettingsCascadeOrError } from '@sourcegraph/shared/src/settings/settings'

import { VSCEState, VSCEStateMachine } from './state'

export interface ExtensionCoreAPI {
    /** For search panel webview to signal that it is ready for messages. */
    panelInitialized: (panelId: string) => void

    requestGraphQL: (request: string, variables: any, overrideAccessToken?: string) => Promise<GraphQLResult<any>>
    observeSourcegraphSettings: () => ProxySubscribable<SettingsCascadeOrError>
    getAuthenticatedUser: () => ProxySubscribable<AuthenticatedUser | null>
    getInstanceURL: () => ProxySubscribable<string>
    setAccessToken: (accessToken: string) => void

    observeState: () => ProxySubscribable<VSCEState>
    emit: VSCEStateMachine['emit']

    openLink: (uri: string) => void
    copyLink: (uri: string) => void
    reloadWindow: () => void

    /**
     * Cancels previous search when called.
     */
    streamSearch: (query: string, options: StreamSearchOptions) => void
    fetchStreamSuggestions: (query: string, sourcegraphURL: string) => ProxySubscribable<SearchMatch[]>
    setSelectedSearchContextSpec: (spec: string) => void
}

// Data flows one way for now (one sidebar <-> one panel UX),
// but these APIs are in place in case we implement a one sidebar <-> many panels UX
export interface SearchPanelAPI {
    // TODO remove once other methods are implemented
    ping: () => ProxySubscribable<'pong'>
}

export interface SearchSidebarAPI extends Pick<FlatExtensionHostAPI, 'addTextDocumentIfNotExists'> {
    // TODO remove once other methods are implemented
    ping: () => ProxySubscribable<'pong'>
    // TODO: ExtensionHostAPI methods
}
