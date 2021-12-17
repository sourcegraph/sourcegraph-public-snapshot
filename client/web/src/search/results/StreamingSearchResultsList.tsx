import classNames from 'classnames'
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
    ContentMatch,
    SymbolMatch,
    PathMatch,
    SearchMatch,
    getMatchUrl,
} from '@sourcegraph/shared/src/search/stream'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'

import { SearchContextProps } from '..'
import { AuthenticatedUser } from '../../auth'
import { SearchResult } from '../../components/SearchResult'
import { SearchUserNeedsCodeHost } from '../../user/settings/codeHosts/OrgUserNeedsCodeHost'

import { NoResultsPage } from './NoResultsPage'
import styles from './StreamingSearchResults.module.scss'
import { StreamingSearchResultFooter } from './StreamingSearchResultsFooter'

const initialItemsToShow = 15
const incrementalItemsToShow = 10

export interface StreamingSearchResultsListProps
    extends ThemeProps,
        SettingsCascadeProps,
        TelemetryProps,
        Pick<SearchContextProps, 'searchContextsEnabled' | 'selectedSearchContextSpec'> {
    isSourcegraphDotCom: boolean
    results?: AggregateStreamingSearchResults
    location: H.Location
    allExpanded: boolean
    fetchHighlightedFileLineRanges: (parameters: FetchFileParameters, force?: boolean) => Observable<string[][]>
    authenticatedUser: AuthenticatedUser | null
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
        if (item.type === 'content' || item.type === 'symbol') {
            return `file:${getMatchUrl(item)}`
        }
        return getMatchUrl(item)
    }, [])

    const logSearchResultClicked = useCallback(() => telemetryService.log('SearchResultClicked'), [telemetryService])

    const renderResult = useCallback(
        (result: SearchMatch): JSX.Element => {
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
                            onSelect={logSearchResultClicked}
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
                        />
                    )
                case 'repo':
                    return (
                        <SearchResult
                            icon={SourceRepositoryIcon}
                            result={result}
                            repoName={result.repository}
                            telemetryService={telemetryService}
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
        ]
    )

    return (
        <>
            <div
                className={classNames(
                    styles.streamingSearchResultsContentCentered,
                    'd-flex flex-column align-items-center'
                )}
            >
                <div className="align-self-stretch">
                    {isSourcegraphDotCom &&
                        searchContextsEnabled &&
                        authenticatedUser &&
                        results?.state === 'complete' &&
                        results?.results.length === 0 && (
                            <SearchUserNeedsCodeHost
                                user={authenticatedUser}
                                orgSearchContext={selectedSearchContextSpec}
                            />
                        )}
                </div>
            </div>
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
                <StreamingSearchResultFooter results={results}>
                    <>
                        {results?.state === 'complete' && results?.results.length === 0 && (
                            <NoResultsPage
                                searchContextsEnabled={searchContextsEnabled}
                                isSourcegraphDotCom={isSourcegraphDotCom}
                                isLightTheme={isLightTheme}
                                telemetryService={telemetryService}
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
