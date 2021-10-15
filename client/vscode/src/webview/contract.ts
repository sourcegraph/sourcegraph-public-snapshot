import { ProxySubscribable } from '@sourcegraph/shared/src/api/extension/api/common'
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

    // For search sidebar

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

export interface SourcegraphVSCodeExtensionHostSidebarAPI {
    // get hover
    // definition
    // references
    // get editor decorations
}
