import * as H from 'history'
import AlphaSBoxIcon from 'mdi-react/AlphaSBoxIcon'
import FileDocumentIcon from 'mdi-react/FileDocumentIcon'
import FileIcon from 'mdi-react/FileIcon'
import SourceCommitIcon from 'mdi-react/SourceCommitIcon'
import SourceRepositoryIcon from 'mdi-react/SourceRepositoryIcon'
import React, { useCallback, useEffect, useState } from 'react'
import { Observable } from 'rxjs'

import { FetchFileParameters } from '@sourcegraph/shared/src/components/CodeExcerpt'
import { FileMatch } from '@sourcegraph/shared/src/components/FileMatch'
import { displayRepoName } from '@sourcegraph/shared/src/components/RepoFileLink'
import { VirtualList } from '@sourcegraph/shared/src/components/VirtualList'
import {
    AggregateStreamingSearchResults,
    FileLineMatch,
    FileSymbolMatch,
    SearchMatch,
    getMatchUrl,
} from '@sourcegraph/shared/src/search/stream'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'

import { SearchResult } from '../../components/SearchResult'
import { eventLogger } from '../../tracking/eventLogger'

import { StreamingSearchResultFooter } from './StreamingSearchResultsFooter'

const initialItemsToShow = 15
const incrementalItemsToShow = 10

export interface StreamingSearchResultsListProps extends ThemeProps, SettingsCascadeProps, TelemetryProps {
    results?: AggregateStreamingSearchResults

    location: H.Location

    allExpanded: boolean

    fetchHighlightedFileLineRanges: (parameters: FetchFileParameters, force?: boolean) => Observable<string[][]>
}

export const StreamingSearchResultsList: React.FunctionComponent<StreamingSearchResultsListProps> = ({
    results,
    location,
    isLightTheme,
    allExpanded,
    fetchHighlightedFileLineRanges,
    settingsCascade,
    telemetryService,
}) => {
    const [itemsToShow, setItemsToShow] = useState(initialItemsToShow)
    const onBottomHit = useCallback(
        () => setItemsToShow(items => Math.min(results?.results.length || 0, items + incrementalItemsToShow)),
        [results?.results.length]
    )

    // Reset scroll visibility state when new search is started
    useEffect(() => {
        setItemsToShow(initialItemsToShow)
    }, [location.search])

    const itemKey = useCallback((item: SearchMatch): string => {
        if (item.type === 'file' || item.type === 'symbol') {
            return `file:${getMatchUrl(item)}`
        }
        return getMatchUrl(item)
    }, [])

    const logSearchResultClicked = useCallback(() => telemetryService.log('SearchResultClicked'), [telemetryService])

    const renderResult = useCallback(
        (result: SearchMatch): JSX.Element => {
            switch (result.type) {
                case 'file':
                case 'symbol':
                    return (
                        <FileMatch
                            location={location}
                            eventLogger={eventLogger}
                            icon={getFileMatchIcon(result)}
                            result={result}
                            onSelect={logSearchResultClicked}
                            expanded={false}
                            showAllMatches={false}
                            isLightTheme={isLightTheme}
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
                            isLightTheme={isLightTheme}
                        />
                    )
                case 'repo':
                    return (
                        <SearchResult
                            icon={SourceRepositoryIcon}
                            result={result}
                            repoName={result.repository}
                            isLightTheme={isLightTheme}
                        />
                    )
            }
        },
        [isLightTheme, location, logSearchResultClicked, allExpanded, fetchHighlightedFileLineRanges, settingsCascade]
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

            {itemsToShow >= (results?.results.length || 0) && <StreamingSearchResultFooter results={results} />}
        </>
    )
}

function getFileMatchIcon(result: FileLineMatch | FileSymbolMatch): React.ComponentType<{ className?: string }> {
    if (result.type === 'file' && result.lineMatches && result.lineMatches.length > 0) {
        return FileDocumentIcon
    }
    if (result.type === 'symbol' && result.symbols && result.symbols.length > 0) {
        return AlphaSBoxIcon
    }
    return FileIcon
}
