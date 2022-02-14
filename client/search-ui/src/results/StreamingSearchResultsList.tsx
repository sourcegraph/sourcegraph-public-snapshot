import classNames from 'classnames'
import * as H from 'history'
import AlphaSBoxIcon from 'mdi-react/AlphaSBoxIcon'
import FileDocumentIcon from 'mdi-react/FileDocumentIcon'
import FileIcon from 'mdi-react/FileIcon'
import SourceCommitIcon from 'mdi-react/SourceCommitIcon'
import SourceRepositoryIcon from 'mdi-react/SourceRepositoryIcon'
import React, { useCallback } from 'react'
import { Observable } from 'rxjs'

import { SettingsCascadeProps } from '@sourcegraph/client-api'
import { SearchContextProps } from '@sourcegraph/search'
import { SearchResult } from '@sourcegraph/search-ui'
import { AuthenticatedUser } from '@sourcegraph/shared/src/auth'
import { FetchFileParameters } from '@sourcegraph/shared/src/components/CodeExcerpt'
import { FileMatch } from '@sourcegraph/shared/src/components/FileMatch'
import { displayRepoName } from '@sourcegraph/shared/src/components/RepoFileLink'
import { VirtualList } from '@sourcegraph/shared/src/components/VirtualList'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import {
    AggregateStreamingSearchResults,
    ContentMatch,
    SymbolMatch,
    PathMatch,
    SearchMatch,
    getMatchUrl,
} from '@sourcegraph/shared/src/search/stream'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'

import { NoResultsPage } from './NoResultsPage'
import { StreamingSearchResultFooter } from './StreamingSearchResultsFooter'
import styles from './StreamingSearchResultsList.module.scss'
import { useItemsToShow } from './use-items-to-show'

export interface StreamingSearchResultsListProps
    extends ThemeProps,
        SettingsCascadeProps,
        TelemetryProps,
        Pick<SearchContextProps, 'searchContextsEnabled'>,
        PlatformContextProps<'requestGraphQL'> {
    isSourcegraphDotCom: boolean
    results?: AggregateStreamingSearchResults
    location: H.Location
    allExpanded: boolean
    fetchHighlightedFileLineRanges: (parameters: FetchFileParameters, force?: boolean) => Observable<string[][]>
    authenticatedUser: AuthenticatedUser | null
    showSearchContext: boolean
    /** Available to web app through JS Context */
    assetsRoot?: string
    /** Render prop for `<SearchUserNeedsCodeHost>`  */
    renderSearchUserNeedsCodeHost?: (user: AuthenticatedUser) => JSX.Element
}

export const StreamingSearchResultsList: React.FunctionComponent<StreamingSearchResultsListProps> = ({
    results,
    location,
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
}) => {
    const resultsNumber = results?.results.length || 0
    const { itemsToShow, handleBottomHit } = useItemsToShow(location.search, resultsNumber)

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
                        <FileMatch
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
                        />
                    )
                case 'commit':
                    return (
                        <SearchResult
                            icon={SourceCommitIcon}
                            result={result}
                            repoName={result.repository}
                            platformContext={platformContext}
                            onSelect={() => logSearchResultClicked(index, 'commit')}
                        />
                    )
                case 'repo':
                    return (
                        <SearchResult
                            icon={SourceRepositoryIcon}
                            result={result}
                            repoName={result.repository}
                            platformContext={platformContext}
                            onSelect={() => logSearchResultClicked(index, 'repo')}
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
