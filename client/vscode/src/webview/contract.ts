import { FlatExtensionHostAPI } from '@sourcegraph/shared/src/api/contract'
import { ProxySubscribable } from '@sourcegraph/shared/src/api/extension/api/common'
import { ViewerData, ViewerId } from '@sourcegraph/shared/src/api/viewerTypes'
import { GraphQLResult } from '@sourcegraph/shared/src/graphql/graphql'
import { SearchPatternType } from '@sourcegraph/shared/src/graphql/schema'
import { QueryState } from '@sourcegraph/shared/src/search/helpers'
import { Filter } from '@sourcegraph/shared/src/search/stream'
import { Settings, SettingsCascadeOrError } from '@sourcegraph/shared/src/settings/settings'

import { SearchSidebarMediator } from './search-sidebar/mediator'

/**
 * Sourcegraph VS Code methods exposed to Webviews
 *
 * TODO: Kind of a "hub" for all webview communication.
 *
 * Note: this API object lives in the VS Code extension host runtime.
 */
export interface SourcegraphVSCodeExtensionAPI
    extends Pick<
        SearchSidebarMediator,
        | 'observeActiveWebviewQueryState'
        | 'observeActiveWebviewDynamicFilters'
        | 'setActiveWebviewQueryState'
        | 'submitActiveWebviewSearch'
    > {
    ping: () => ProxySubscribable<'pong'>

    // Shared methods
    requestGraphQL: (request: string, variables: any) => Promise<GraphQLResult<any>>
    getSettings: () => ProxySubscribable<SettingsCascadeOrError<Settings>>

    // For search webview
    panelInitialized: (panelId: string) => void
    /** TODO explain, we deliberately do not react to URL changes in webviews. */
    getInstanceHostname: () => string
    /** TODO document. sourcegraph://${host}/${uri} */
    openFile: (sourcegraphUri: string) => void

    // For search sidebar
    openSearchPanel: () => void

    // For extension host sidebar
    // mainThreadAPI methods
}

/**
 * Search webview methods exposed to the Sourcegraph VS Code extension.
 */
export interface SourcegraphVSCodeSearchWebviewAPI {
    observeQueryState: () => ProxySubscribable<QueryStateWithInputProps>
    observeDynamicFilters: () => ProxySubscribable<Filter[] | null>
    setQueryState: (queryState: QueryState) => void
    submitSearch: (queryState?: QueryState) => void
}

export interface QueryStateWithInputProps {
    queryState: QueryState
    caseSensitive: boolean
    patternType: SearchPatternType
}

export interface SourcegraphVSCodeSearchSidebarAPI {}

/**
 * A subset of the Sourcegraph extension host API that is used by the VS Code extension.
 * TODO just extend + pick
 */
export interface SourcegraphVSCodeExtensionHostAPI
    extends Pick<FlatExtensionHostAPI, 'getDefinition' | 'getHover' | 'getReferences' | 'addTextDocumentIfNotExists'> {
    // get hover
    // definition
    // getDefinition: (
    //     parameters: TextDocumentPositionParameters
    //     // instance URL?
    // ) => ProxySubscribable<MaybeLoadingResult<HoverMerged | null>>
    addViewerIfNotExists(viewer: ViewerData): Promise<ViewerId>
    // add
    // TODO addWorkspaceRoot if necessary?
    // references
    // get editor decorations
}
