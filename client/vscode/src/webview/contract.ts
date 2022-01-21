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
    // Get Access Token
    hasAccessToken: () => boolean
    // If Access Token is valid
    hasValidAccessToken: () => boolean
    // Update Access Token - return true when updated successfully
    updateAccessToken: (token: string) => Promise<boolean>
    /** TODO document. sourcegraph://${host}/${uri} */
    openFile: (sourcegraphUri: string) => void
    // Open links in browser
    openLink: (uri: string) => void
    // Copy Link to Clipboard
    copyLink: (uri: string) => void
    // For search sidebar
    openSearchPanel: () => void
    // Check if on VS Code Desktop or VS Code Web
    onDesktop: () => boolean
    // Get Cors from Setting
    getCorsSetting: () => string
    // Update Cors Setting - return true when updated successfully
    updateCorsUri: (uri: string) => Promise<boolean>
    // Get item from VSCE local storage
    getLocalStorageItem: (key: string) => string[]
    // Set item in VSCE local storage
    setLocalStorageItem: (key: string, value: string[]) => Promise<boolean>
    // Get Last Selected Search Context from Local Storage
    getLastSelectedSearchContext: () => string
    // Update Last Selected Search Context in Local Storage
    updateLastSelectedSearchContext: (context: string) => Promise<boolean>
    // Get Last Selected Search Context from Local Storage
    getLocalRecentSearch: () => LocalRecentSeachProps[]
    // Update Last Selected Search Context in Local Storage
    setLocalRecentSearch: (searches: LocalRecentSeachProps[]) => Promise<boolean>
    // Display File Tree when repo is clicked
    displayFileTree: (setting: boolean) => void
    hasActivePanel: () => void
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
    executed?: boolean
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

export interface LocalRecentSeachProps {
    lastQuery: string
    lastSelectedSearchContextSpec: string
    lastCaseSensitive: boolean
    lastPatternType: string
    lastFullQuery: string
}

export interface LocalFileHistoryProps {
    repoName: string
    filePath: string
    sgUri: string
}
