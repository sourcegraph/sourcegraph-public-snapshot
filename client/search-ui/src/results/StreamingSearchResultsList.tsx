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

import { smartSearchClickedEvent } from '../util/events'

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
    showSearchContext: boolean
    /** Clicking on a match opens the link in a new tab. */
    openMatchesInNewTab?: boolean
    /** Available to web app through JS Context */
    assetsRoot?: string

    extensionsController?: Pick<ExtensionsController, 'extHostAPI'> | null
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

    /**
     * For A/B testing on Sourcegraph.com. To be removed at latest by 12/2022.
     */
    smartSearchEnabled?: boolean
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
    showSearchContext,
    assetsRoot,
    platformContext,
    extensionsController,
    hoverifier,
    openMatchesInNewTab,
    executedQuery,
    resultClassName,
    smartSearchEnabled: smartSearchEnabled,
}) => {
    const resultsNumber = results?.results.length || 0
    const { itemsToShow, handleBottomHit } = useItemsToShow(executedQuery, resultsNumber)
    const location = useLocation()

    const logSearchResultClicked = useCallback(
        (index: number, type: string) => {
            telemetryService.log('SearchResultClicked')

            // This data ends up in Prometheus and is not part of the ping payload.
            telemetryService.log('search.ranking.result-clicked', { index, type })

            // Lucky search A/B test events on Sourcegraph.com. To be removed at latest by 12/2022.
            if (smartSearchEnabled && !(results?.alert?.kind === 'lucky-search-queries')) {
                telemetryService.log('SearchResultClickedAutoNone')
            }

            if (
                smartSearchEnabled &&
                results?.alert?.kind === 'lucky-search-queries' &&
                results?.alert?.title &&
                results.alert.proposedQueries
            ) {
                const event = smartSearchClickedEvent(
                    results.alert.title,
                    results.alert.proposedQueries.map(entry => entry.description || '')
                )

                telemetryService.log(event)
            }
        },
        [telemetryService, results, smartSearchEnabled]
    )

    const renderResult = useCallback(
        (result: SearchMatch, index: number): JSX.Element => {
            switch (result.type) {
                case 'content':
                case 'path':
                case 'symbol':
                    return (
                        <FileSearchResult
                            index={index}
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
                            as="li"
                        />
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
        },
        [
            location,
            telemetryService,
            allExpanded,
            fetchHighlightedFileLineRanges,
            settingsCascade,
            platformContext,
            extensionsController,
            hoverifier,
            openMatchesInNewTab,
            resultClassName,
            logSearchResultClicked,
        ]
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
    if (item.type === 'content') {
        const lineStart = item.lineMatches.length > 0 ? item.lineMatches[0].lineNumber : 0
        return `file:${getMatchUrl(item)}:${lineStart}`
    }
    if (item.type === 'symbol') {
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
