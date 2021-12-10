import * as H from 'history'
import AlphaSBoxIcon from 'mdi-react/AlphaSBoxIcon'
import FileDocumentIcon from 'mdi-react/FileDocumentIcon'
import FileIcon from 'mdi-react/FileIcon'
import SourceCommitIcon from 'mdi-react/SourceCommitIcon'
import SourceRepositoryIcon from 'mdi-react/SourceRepositoryIcon'
import React, { useCallback } from 'react'
import { Observable } from 'rxjs'

import { SearchResult } from '@sourcegraph/branded/src/components/SearchResult'
import { NoResultsPage } from '@sourcegraph/branded/src/search/results/NoResultsPage'
import { StreamingSearchResultFooter } from '@sourcegraph/branded/src/search/results/StreamingSearchResultsFooter'
import { AuthenticatedUser } from '@sourcegraph/shared/src/auth'
import { FetchFileParameters } from '@sourcegraph/shared/src/components/CodeExcerpt'
import { FileMatch } from '@sourcegraph/shared/src/components/FileMatch'
import { displayRepoName } from '@sourcegraph/shared/src/components/RepoFileLink'
import { VirtualList } from '@sourcegraph/shared/src/components/VirtualList'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { SearchContextProps } from '@sourcegraph/shared/src/search'
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

import { useItemsToShow } from './use-items-to-show'

export interface StreamingSearchResultsListProps
    extends ThemeProps,
        SettingsCascadeProps,
        TelemetryProps,
        Pick<SearchContextProps, 'searchContextsEnabled' | 'selectedSearchContextSpec'>,
        PlatformContextProps<'requestGraphQL'> {
    isSourcegraphDotCom: boolean
    results?: AggregateStreamingSearchResults
    location?: H.Location
    allExpanded: boolean
    fetchHighlightedFileLineRanges: (parameters: FetchFileParameters, force?: boolean) => Observable<string[][]>
    authenticatedUser: AuthenticatedUser | null
    footerClassName?: string
    /** Available to web app through JS Context */
    assetsRoot?: string
    /** Scroll visibility state is reset when a new search is executed. */
    executedQuery: string
    /**
     * Used in contexts where we need to run a callback when a search result is selected.
     * Example: Running "open file" command in VS Code on search result click.
     */
    onSelect?: (result: SearchMatch) => void
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
    selectedSearchContextSpec,
    authenticatedUser,
    platformContext,
    footerClassName,
    assetsRoot,
    executedQuery,
    onSelect,
}) => {
    const resultsNumber = results?.results.length || 0
    const { itemsToShow, handleBottomHit } = useItemsToShow(executedQuery, resultsNumber)

    const logSearchResultClicked = useCallback(() => telemetryService.log('SearchResultClicked'), [telemetryService])

    const renderResult = useCallback(
        (result: SearchMatch): JSX.Element => {
            const onFileMatchClicked = (): void => {
                logSearchResultClicked()
                onSelect?.(result)
            }
            const onSearchResultClicked = (): void => {
                onSelect?.(result)
            }

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
                            onSelect={onFileMatchClicked}
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
                            telemetryService={telemetryService}
                            platformContext={platformContext}
                            onSelect={onSearchResultClicked}
                        />
                    )
                case 'repo':
                    return (
                        <SearchResult
                            icon={SourceRepositoryIcon}
                            result={result}
                            repoName={result.repository}
                            telemetryService={telemetryService}
                            platformContext={platformContext}
                            onSelect={onSearchResultClicked}
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
            onSelect,
        ]
    )
    return (
        <>
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
                <StreamingSearchResultFooter results={results} className={footerClassName}>
                    <>
                        {results?.state === 'complete' && resultsNumber === 0 && (
                            <NoResultsPage
                                searchContextsEnabled={searchContextsEnabled}
                                isSourcegraphDotCom={isSourcegraphDotCom}
                                isLightTheme={isLightTheme}
                                telemetryService={telemetryService}
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
