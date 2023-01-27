import React, { useCallback, useState } from 'react'

import classNames from 'classnames'
import { useLocation } from 'react-router'
import { Observable } from 'rxjs'

import { TraceSpanProvider } from '@sourcegraph/observability-client'
import { FetchFileParameters } from '@sourcegraph/shared/src/backend/file'
import { FilePrefetcher, PrefetchableFile } from '@sourcegraph/shared/src/components/PrefetchableFile'
import { displayRepoName } from '@sourcegraph/shared/src/components/RepoLink'
import { VirtualList } from '@sourcegraph/shared/src/components/VirtualList'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { QueryState, SearchContextProps } from '@sourcegraph/shared/src/search'
import {
    AggregateStreamingSearchResults,
    SearchMatch,
    getMatchUrl,
    getRevision,
} from '@sourcegraph/shared/src/search/stream'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'

import { CommitSearchResult } from '../components/CommitSearchResult'
import { FileContentSearchResult } from '../components/FileContentSearchResult'
import { FilePathSearchResult } from '../components/FilePathSearchResult'
import { RepoSearchResult } from '../components/RepoSearchResult'
import { SymbolSearchResult } from '../components/SymbolSearchResult'
import { smartSearchClickedEvent } from '../util/events'

import { NoResultsPage } from './NoResultsPage'
import { StreamingSearchResultFooter } from './StreamingSearchResultsFooter'
import { useItemsToShow } from './use-items-to-show'
import { useSearchResultsKeyboardNavigation } from './useSearchResultsKeyboardNavigation'

import resultContainerStyles from '../components/ResultContainer.module.scss'
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
    /** Clicking on a match opens the link in a new tab. */
    openMatchesInNewTab?: boolean
    /** Available to web app through JS Context */
    assetsRoot?: string

    /**
     * Latest run query. Resets scroll visibility state when changed.
     * For example, `location.search` on web.
     */
    executedQuery: string
    /**
     * Classname to be applied to the container of a search result.
     */
    resultClassName?: string

    prefetchFile?: FilePrefetcher

    prefetchFileEnabled?: boolean

    enableKeyboardNavigation?: boolean

    showQueryExamplesOnNoResultsPage?: boolean

    /*
     * For updating the query from the QueryExamples on NoResultsPage. Only
     * needed if showQueryExamplesOnNoResultsPage is true.
     */
    setQueryState?: (query: QueryState) => void
    selectedSearchContextSpec?: string
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
    assetsRoot,
    platformContext,
    openMatchesInNewTab,
    executedQuery,
    resultClassName,
    prefetchFile,
    prefetchFileEnabled,
    enableKeyboardNavigation,
    showQueryExamplesOnNoResultsPage,
    setQueryState,
}) => {
    const resultsNumber = results?.results.length || 0
    const { itemsToShow, handleBottomHit } = useItemsToShow(executedQuery, resultsNumber)
    const location = useLocation()
    const [rootRef, setRootRef] = useState<HTMLElement | null>(null)

    const logSearchResultClicked = useCallback(
        (index: number, type: string) => {
            telemetryService.log('SearchResultClicked')

            // This data ends up in Prometheus and is not part of the ping payload.
            telemetryService.log('search.ranking.result-clicked', { index, type })

            // Lucky search A/B test events on Sourcegraph.com. To be removed at latest by 12/2022.
            if (
                !(
                    results?.alert?.kind === 'smart-search-additional-results' ||
                    results?.alert?.kind === 'smart-search-pure-results'
                )
            ) {
                telemetryService.log('SearchResultClickedAutoNone')
            }

            if (
                (results?.alert?.kind === 'smart-search-additional-results' ||
                    results?.alert?.kind === 'smart-search-pure-results') &&
                results?.alert?.title &&
                results.alert.proposedQueries
            ) {
                const event = smartSearchClickedEvent(
                    results.alert.kind,
                    results.alert.title,
                    results.alert.proposedQueries.map(entry => entry.description || '')
                )

                telemetryService.log(event)
            }
        },
        [telemetryService, results]
    )

    const renderResult = useCallback(
        (result: SearchMatch, index: number): JSX.Element => {
            function renderResultContent(): JSX.Element {
                switch (result.type) {
                    case 'content':
                    case 'symbol':
                    case 'path':
                        return (
                            <PrefetchableFile
                                isPrefetchEnabled={prefetchFileEnabled}
                                prefetch={prefetchFile}
                                filePath={result.path}
                                revision={getRevision(result.branches, result.commit)}
                                repoName={result.repository}
                                // PrefetchableFile adds an extra wrapper, so we lift the <li> up and match the ResultContainer styles.
                                // Better approach would be to use `as` to avoid wrapping, but that requires a larger refactor of the
                                // child components than is worth doing right now for this experimental feature
                                className={resultContainerStyles.resultContainer}
                                as="li"
                            >
                                {result.type === 'content' && (
                                    <FileContentSearchResult
                                        index={index}
                                        location={location}
                                        telemetryService={telemetryService}
                                        result={result}
                                        onSelect={() => logSearchResultClicked(index, 'fileMatch')}
                                        defaultExpanded={false}
                                        showAllMatches={false}
                                        allExpanded={allExpanded}
                                        fetchHighlightedFileLineRanges={fetchHighlightedFileLineRanges}
                                        repoDisplayName={displayRepoName(result.repository)}
                                        settingsCascade={settingsCascade}
                                        openInNewTab={openMatchesInNewTab}
                                        containerClassName={resultClassName}
                                    />
                                )}
                                {result.type === 'symbol' && (
                                    <SymbolSearchResult
                                        index={index}
                                        telemetryService={telemetryService}
                                        result={result}
                                        onSelect={() => logSearchResultClicked(index, 'symbolMatch')}
                                        fetchHighlightedFileLineRanges={fetchHighlightedFileLineRanges}
                                        repoDisplayName={displayRepoName(result.repository)}
                                        settingsCascade={settingsCascade}
                                        openInNewTab={openMatchesInNewTab}
                                        containerClassName={resultClassName}
                                    />
                                )}
                                {result.type === 'path' && (
                                    <FilePathSearchResult
                                        index={index}
                                        result={result}
                                        onSelect={() => logSearchResultClicked(index, 'filePathMatch')}
                                        repoDisplayName={displayRepoName(result.repository)}
                                        containerClassName={resultClassName}
                                        telemetryService={telemetryService}
                                    />
                                )}
                            </PrefetchableFile>
                        )
                    case 'commit':
                        return (
                            <CommitSearchResult
                                index={index}
                                result={result}
                                platformContext={platformContext}
                                onSelect={() => logSearchResultClicked(index, 'commit')}
                                openInNewTab={openMatchesInNewTab}
                                containerClassName={resultClassName}
                                as="li"
                            />
                        )
                    case 'repo':
                        return (
                            <RepoSearchResult
                                index={index}
                                result={result}
                                onSelect={() => logSearchResultClicked(index, 'repo')}
                                containerClassName={resultClassName}
                                as="li"
                            />
                        )
                }
            }

            return (
                <TraceSpanProvider
                    name="StreamingSearchResultsListItem"
                    attributes={{
                        type: result.type,
                        index,
                    }}
                >
                    {renderResultContent()}
                </TraceSpanProvider>
            )
        },
        [
            prefetchFileEnabled,
            prefetchFile,
            location,
            telemetryService,
            allExpanded,
            fetchHighlightedFileLineRanges,
            settingsCascade,
            openMatchesInNewTab,
            resultClassName,
            platformContext,
            logSearchResultClicked,
        ]
    )

    const [showFocusInputMessage, onVisibilityChange] = useSearchResultsKeyboardNavigation(
        rootRef,
        enableKeyboardNavigation
    )

    return (
        <>
            <VirtualList<SearchMatch>
                as="ol"
                aria-label="Search results"
                className={classNames('mt-2 mb-0', styles.list)}
                itemsToShow={itemsToShow}
                onShowMoreItems={handleBottomHit}
                items={results?.results || []}
                itemProps={undefined}
                itemKey={itemKey}
                renderItem={renderResult}
                onRef={setRootRef}
                onVisibilityChange={onVisibilityChange}
            />

            <div
                className={classNames(styles.focusInputMessage, showFocusInputMessage && styles.focusInputMessageShow)}
            >
                Press <span className={styles.focusInputMessageSlash}>/</span> to focus the search input
            </div>

            {itemsToShow >= resultsNumber && (
                <StreamingSearchResultFooter results={results} telemetryService={telemetryService}>
                    <>
                        {results?.state === 'complete' && resultsNumber === 0 && (
                            <NoResultsPage
                                searchContextsEnabled={searchContextsEnabled}
                                isSourcegraphDotCom={isSourcegraphDotCom}
                                isLightTheme={isLightTheme}
                                telemetryService={telemetryService}
                                showSearchContext={searchContextsEnabled}
                                assetsRoot={assetsRoot}
                                showQueryExamples={showQueryExamplesOnNoResultsPage}
                                setQueryState={setQueryState}
                            />
                        )}
                    </>
                </StreamingSearchResultFooter>
            )}
        </>
    )
}

function itemKey(item: SearchMatch): string {
    if (item.type === 'content') {
        const lineStart = item.chunkMatches
            ? item.chunkMatches.length > 0
                ? item.chunkMatches[0].contentStart.line
                : 0
            : 0
        return `file:${getMatchUrl(item)}:${lineStart}`
    }
    if (item.type === 'symbol') {
        return `file:${getMatchUrl(item)}`
    }
    return getMatchUrl(item)
}
