import React, { useCallback, useState } from 'react'

import classNames from 'classnames'
import type { Observable } from 'rxjs'

import { TraceSpanProvider } from '@sourcegraph/observability-client'
import type { FetchFileParameters } from '@sourcegraph/shared/src/backend/file'
import { type FilePrefetcher, PrefetchableFile } from '@sourcegraph/shared/src/components/PrefetchableFile'
import { displayRepoName } from '@sourcegraph/shared/src/components/RepoLink'
import { VirtualList } from '@sourcegraph/shared/src/components/VirtualList'
import type { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import type {
    BuildSearchQueryURLParameters,
    QueryState,
    SearchContextProps,
    SearchMode,
    SubmitSearchParameters,
} from '@sourcegraph/shared/src/search'
import {
    type AggregateStreamingSearchResults,
    getMatchUrl,
    getRevision,
    type SearchMatch,
} from '@sourcegraph/shared/src/search/stream'
import type { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import type { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import {
    CommitSearchResult,
    FileContentSearchResult,
    FilePathSearchResult,
    RepoSearchResult,
    SymbolSearchResult,
} from '../components'
import { OwnerSearchResult } from '../components/OwnerSearchResult'

import { NoResultsPage } from './NoResultsPage'
import { StreamingSearchResultFooter } from './StreamingSearchResultsFooter'
import { useItemsToShow } from './use-items-to-show'
import { useSearchResultsKeyboardNavigation } from './useSearchResultsKeyboardNavigation'

import resultContainerStyles from '../components/ResultContainer.module.scss'
import styles from './StreamingSearchResultsList.module.scss'

export interface StreamingSearchResultsListProps
    extends SettingsCascadeProps,
        TelemetryProps,
        Pick<SearchContextProps, 'searchContextsEnabled'>,
        PlatformContextProps<'requestGraphQL'> {
    isSourcegraphDotCom: boolean
    results?: AggregateStreamingSearchResults
    allExpanded: boolean
    fetchHighlightedFileLineRanges: (parameters: FetchFileParameters, force?: boolean) => Observable<string[][]>
    /** Clicking on a match opens the link in a new tab. */
    openMatchesInNewTab?: boolean

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

    /**
     * The query state to be used for the query examples and owner search.
     * If not provided, the query examples and owner search will not
     * allow modifying the query.
     */
    queryState?: QueryState
    buildSearchURLQueryFromQueryState?: (queryParameters: BuildSearchQueryURLParameters) => string

    searchMode?: SearchMode
    setSearchMode?: (mode: SearchMode) => void
    submitSearch?: (parameters: SubmitSearchParameters) => void
    searchQueryFromURL?: string
    caseSensitive?: boolean

    selectedSearchContextSpec?: string

    /**
     * An optional callback invoked whenever a search result is clicked.
     * It's passed the index of the result in the list and the result type.
     */
    logSearchResultClicked?: (index: number, type: string, resultsLength: number) => void

    enableRepositoryMetadata?: boolean
}

export const StreamingSearchResultsList: React.FunctionComponent<
    React.PropsWithChildren<StreamingSearchResultsListProps>
> = ({
    results,
    allExpanded,
    fetchHighlightedFileLineRanges,
    settingsCascade,
    telemetryService,
    isSourcegraphDotCom,
    searchContextsEnabled,
    platformContext,
    openMatchesInNewTab,
    executedQuery,
    resultClassName,
    prefetchFile,
    prefetchFileEnabled,
    enableKeyboardNavigation,
    showQueryExamplesOnNoResultsPage,
    queryState,
    buildSearchURLQueryFromQueryState,
    searchMode,
    setSearchMode,
    submitSearch,
    caseSensitive,
    searchQueryFromURL,
    logSearchResultClicked,
    enableRepositoryMetadata,
}) => {
    const resultsNumber = results?.results.length || 0
    const { itemsToShow, handleBottomHit } = useItemsToShow(executedQuery, resultsNumber)
    const [rootRef, setRootRef] = useState<HTMLElement | null>(null)

    const renderResult = useCallback(
        (result: SearchMatch, index: number): JSX.Element => {
            function renderResultContent(): JSX.Element {
                switch (result.type) {
                    case 'content':
                    case 'symbol':
                    case 'path': {
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
                                        telemetryService={telemetryService}
                                        result={result}
                                        onSelect={() => logSearchResultClicked?.(index, 'fileMatch', resultsNumber)}
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
                                        onSelect={() => logSearchResultClicked?.(index, 'symbolMatch', resultsNumber)}
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
                                        onSelect={() => logSearchResultClicked?.(index, 'filePathMatch', resultsNumber)}
                                        repoDisplayName={displayRepoName(result.repository)}
                                        containerClassName={resultClassName}
                                        telemetryService={telemetryService}
                                        settingsCascade={settingsCascade}
                                    />
                                )}
                            </PrefetchableFile>
                        )
                    }
                    case 'commit': {
                        return (
                            <CommitSearchResult
                                index={index}
                                result={result}
                                platformContext={platformContext}
                                onSelect={() => logSearchResultClicked?.(index, 'commit', resultsNumber)}
                                openInNewTab={openMatchesInNewTab}
                                containerClassName={resultClassName}
                                as="li"
                            />
                        )
                    }
                    case 'repo': {
                        return (
                            <RepoSearchResult
                                index={index}
                                result={result}
                                onSelect={() => logSearchResultClicked?.(index, 'repo', resultsNumber)}
                                containerClassName={resultClassName}
                                buildSearchURLQueryFromQueryState={buildSearchURLQueryFromQueryState}
                                queryState={queryState}
                                enableRepositoryMetadata={enableRepositoryMetadata}
                                as="li"
                            />
                        )
                    }
                    case 'person':
                    case 'team': {
                        return (
                            <OwnerSearchResult
                                index={index}
                                result={result}
                                as="li"
                                onSelect={() => logSearchResultClicked?.(index, 'person', resultsNumber)}
                                containerClassName={resultClassName}
                                telemetryService={telemetryService}
                                queryState={queryState}
                                buildSearchURLQueryFromQueryState={buildSearchURLQueryFromQueryState}
                            />
                        )
                    }
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
            resultsNumber,
            telemetryService,
            allExpanded,
            fetchHighlightedFileLineRanges,
            settingsCascade,
            openMatchesInNewTab,
            resultClassName,
            platformContext,
            queryState,
            enableRepositoryMetadata,
            buildSearchURLQueryFromQueryState,
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
                <StreamingSearchResultFooter results={results}>
                    <>
                        {results?.state === 'complete' && resultsNumber === 0 && (
                            <NoResultsPage
                                searchContextsEnabled={searchContextsEnabled}
                                isSourcegraphDotCom={isSourcegraphDotCom}
                                telemetryService={telemetryService}
                                showSearchContext={searchContextsEnabled}
                                showQueryExamples={showQueryExamplesOnNoResultsPage}
                                searchMode={searchMode}
                                setSearchMode={setSearchMode}
                                submitSearch={submitSearch}
                                caseSensitive={caseSensitive}
                                searchQueryFromURL={searchQueryFromURL}
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
