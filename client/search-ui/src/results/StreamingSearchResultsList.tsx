import React, { useCallback } from 'react'

import classNames from 'classnames'
import AlphaSBoxIcon from 'mdi-react/AlphaSBoxIcon'
import FileDocumentIcon from 'mdi-react/FileDocumentIcon'
import FileIcon from 'mdi-react/FileIcon'
import { useLocation } from 'react-router'
import { Observable } from 'rxjs'

import { HoverMerged } from '@sourcegraph/client-api'
import { Hoverifier } from '@sourcegraph/codeintellify'
import { SearchContextProps } from '@sourcegraph/search'
import { CommitSearchResult, RepoSearchResult, FileSearchResult, FetchFileParameters } from '@sourcegraph/search-ui'
import { ActionItemAction } from '@sourcegraph/shared/src/actions/ActionItem'
import { AuthenticatedUser } from '@sourcegraph/shared/src/auth'
import { displayRepoName } from '@sourcegraph/shared/src/components/RepoLink'
import { VirtualList } from '@sourcegraph/shared/src/components/VirtualList'
import { Controller as ExtensionsController } from '@sourcegraph/shared/src/extensions/controller'
import { HoverContext } from '@sourcegraph/shared/src/hover/HoverOverlay.types'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import {
    AggregateStreamingSearchResults,
    ContentMatch,
    SymbolMatch,
    PathMatch,
    SearchMatch,
    getMatchUrl,
} from '@sourcegraph/shared/src/search/stream'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'

import { NoResultsPage } from './NoResultsPage'
import { StreamingSearchResultFooter } from './StreamingSearchResultsFooter'
import { useItemsToShow } from './use-items-to-show'

import styles from './StreamingSearchResultsList.module.scss'

export interface StreamingSearchResultsListProps
    extends ThemeProps,
        SettingsCascadeProps,
        TelemetryProps,
        Pick<SearchContextProps, 'searchContextsEnabled'>,
        PlatformContextProps<'requestGraphQL'> {
    isSourcegraphDotCom: boolean
    results?: AggregateStreamingSearchResults
    allExpanded: boolean
    fetchHighlightedFileLineRanges: (parameters: FetchFileParameters, force?: boolean) => Observable<string[][]>
    authenticatedUser: AuthenticatedUser | null
    showSearchContext: boolean
    /** Clicking on a match opens the link in a new tab. */
    openMatchesInNewTab?: boolean
    /** Available to web app through JS Context */
    assetsRoot?: string
    /** Render prop for `<SearchUserNeedsCodeHost>`  */
    renderSearchUserNeedsCodeHost?: (user: AuthenticatedUser) => JSX.Element

    extensionsController?: Pick<ExtensionsController, 'extHostAPI'>
    hoverifier?: Hoverifier<HoverContext, HoverMerged, ActionItemAction>
    /**
     * Latest run query. Resets scroll visibility state when changed.
     * For example, `location.search` on web.
     */
    executedQuery: string
    /**
     * Classname to be applied to the container of a search result.
     */
    resultClassName?: string
}

export const StreamingSearchResultsList: React.FunctionComponent<
    React.PropsWithChildren<StreamingSearchResultsListProps>
> = ({
    results,
    allExpanded,
    fetchHighlightedFileLineRanges,
    settingsCascade,
    telemetryService,
    isLightTheme,
    isSourcegraphDotCom,
    searchContextsEnabled,
    authenticatedUser,
    showSearchContext,
    assetsRoot,
    renderSearchUserNeedsCodeHost,
    platformContext,
    extensionsController,
    hoverifier,
    openMatchesInNewTab,
    executedQuery,
    resultClassName,
}) => {
    const resultsNumber = results?.results.length || 0
    const { itemsToShow, handleBottomHit } = useItemsToShow(executedQuery, resultsNumber)
    const location = useLocation()

    const logSearchResultClicked = useCallback(
        (index: number, type: string) => {
            telemetryService.log('SearchResultClicked')

            // This data ends up in Prometheus and is not part of the ping payload.
            telemetryService.log('search.ranking.result-clicked', { index, type })
        },
        [telemetryService]
    )

    const renderResult = useCallback(
        (result: SearchMatch, index: number): JSX.Element => {
            switch (result.type) {
                case 'content':
                case 'path':
                case 'symbol':
                    return (
                        <FileSearchResult
                            location={location}
                            telemetryService={telemetryService}
                            icon={getFileMatchIcon(result)}
                            result={result}
                            onSelect={() => logSearchResultClicked(index, 'fileMatch')}
                            expanded={false}
                            showAllMatches={false}
                            allExpanded={allExpanded}
                            fetchHighlightedFileLineRanges={fetchHighlightedFileLineRanges}
                            repoDisplayName={displayRepoName(result.repository)}
                            settingsCascade={settingsCascade}
                            extensionsController={extensionsController}
                            hoverifier={hoverifier}
                            openInNewTab={openMatchesInNewTab}
                            containerClassName={resultClassName}
                        />
                    )
                case 'commit':
                    return (
                        <CommitSearchResult
                            result={result}
                            platformContext={platformContext}
                            onSelect={() => logSearchResultClicked(index, 'commit')}
                            openInNewTab={openMatchesInNewTab}
                            containerClassName={resultClassName}
                        />
                    )
                case 'repo':
                    return (
                        <RepoSearchResult
                            result={result}
                            onSelect={() => logSearchResultClicked(index, 'repo')}
                            containerClassName={resultClassName}
                        />
                    )
            }
        },
        [
            location,
            telemetryService,
            logSearchResultClicked,
            allExpanded,
            fetchHighlightedFileLineRanges,
            settingsCascade,
            platformContext,
            extensionsController,
            hoverifier,
            openMatchesInNewTab,
            resultClassName,
        ]
    )

    return (
        <>
            <div className={classNames(styles.contentCentered, 'd-flex flex-column align-items-center')}>
                <div className="align-self-stretch">
                    {renderSearchUserNeedsCodeHost &&
                        isSourcegraphDotCom &&
                        searchContextsEnabled &&
                        authenticatedUser &&
                        results?.state === 'complete' &&
                        results?.results.length === 0 &&
                        renderSearchUserNeedsCodeHost(authenticatedUser)}
                </div>
            </div>
            <VirtualList<SearchMatch>
                className="mt-2"
                itemsToShow={itemsToShow}
                onShowMoreItems={handleBottomHit}
                items={results?.results || []}
                itemProps={undefined}
                itemKey={itemKey}
                renderItem={renderResult}
            />

            {itemsToShow >= resultsNumber && (
                <StreamingSearchResultFooter results={results}>
                    <>
                        {results?.state === 'complete' && resultsNumber === 0 && (
                            <NoResultsPage
                                searchContextsEnabled={searchContextsEnabled}
                                isSourcegraphDotCom={isSourcegraphDotCom}
                                isLightTheme={isLightTheme}
                                telemetryService={telemetryService}
                                showSearchContext={showSearchContext}
                                assetsRoot={assetsRoot}
                            />
                        )}
                    </>
                </StreamingSearchResultFooter>
            )}
        </>
    )
}

function itemKey(item: SearchMatch): string {
    if (item.type === 'content' || item.type === 'symbol') {
        return `file:${getMatchUrl(item)}`
    }
    return getMatchUrl(item)
}

function getFileMatchIcon(result: ContentMatch | SymbolMatch | PathMatch): React.ComponentType<{ className?: string }> {
    switch (result.type) {
        case 'content':
            return FileDocumentIcon
        case 'symbol':
            return AlphaSBoxIcon
        case 'path':
            return FileIcon
    }
}
