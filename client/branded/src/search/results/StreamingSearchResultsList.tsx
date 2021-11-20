import * as H from 'history'
import AlphaSBoxIcon from 'mdi-react/AlphaSBoxIcon'
import FileDocumentIcon from 'mdi-react/FileDocumentIcon'
import FileIcon from 'mdi-react/FileIcon'
import SourceCommitIcon from 'mdi-react/SourceCommitIcon'
import SourceRepositoryIcon from 'mdi-react/SourceRepositoryIcon'
import React, { useCallback, useEffect, useState } from 'react'
import { Observable } from 'rxjs'

import { SearchResult } from '@sourcegraph/branded/src/components/SearchResult'
import { NoResultsPage } from '@sourcegraph/branded/src/search/results/NoResultsPage'
import { StreamingSearchResultFooter } from '@sourcegraph/branded/src/search/results/StreamingSearchResultsFooter'
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

const initialItemsToShow = 15
const incrementalItemsToShow = 10

export interface StreamingSearchResultsListProps
    extends ThemeProps,
        SettingsCascadeProps,
        TelemetryProps,
        Pick<SearchContextProps, 'searchContextsEnabled' | 'showSearchContext'>,
        PlatformContextProps<'requestGraphQL'> {
    isSourcegraphDotCom: boolean
    results?: AggregateStreamingSearchResults
    location?: H.Location
    allExpanded: boolean
    fetchHighlightedFileLineRanges: (parameters: FetchFileParameters, force?: boolean) => Observable<string[][]>
    footerClassName?: string
    /** Available to web app through JS Context */
    assetsRoot?: string
    /** TODO EXPLAIN */
    executedQuery: string

    /** TODO explain */
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
    showSearchContext,
    platformContext,
    footerClassName,
    assetsRoot,
    executedQuery,
    onSelect,
}) => {
    const [itemsToShow, setItemsToShow] = useState(initialItemsToShow)
    const onBottomHit = useCallback(
        () => setItemsToShow(items => Math.min(results?.results.length || 0, items + incrementalItemsToShow)),
        [results?.results.length]
    )

    // Reset scroll visibility state when new search is started
    useEffect(() => {
        setItemsToShow(initialItemsToShow)
    }, [executedQuery])

    const itemKey = useCallback((item: SearchMatch): string => {
        if (item.type === 'content' || item.type === 'symbol') {
            return `file:${getMatchUrl(item)}`
        }
        return getMatchUrl(item)
    }, [])

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
                onShowMoreItems={onBottomHit}
                items={results?.results || []}
                itemProps={undefined}
                itemKey={itemKey}
                renderItem={renderResult}
            />

            {itemsToShow >= (results?.results.length || 0) && (
                <StreamingSearchResultFooter results={results} className={footerClassName}>
                    <>
                        {results?.state === 'complete' && results?.results.length === 0 && (
                            <NoResultsPage
                                searchContextsEnabled={searchContextsEnabled}
                                showSearchContext={showSearchContext}
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
