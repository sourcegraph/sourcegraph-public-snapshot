import React, { useCallback, useEffect, useMemo, useState } from 'react'

import { gql } from '@apollo/client'
import classNames from 'classnames'
import { streamComputeQuery } from '../../../../shared/src/search/stream'

import { SyntaxHighlightedSearchQuery } from '@sourcegraph/search-ui'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { buildSearchURLQuery } from '@sourcegraph/shared/src/util/url'
import { Link, useObservable } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../auth'
import { Timestamp } from '../../components/time/Timestamp'
import { RecentSearchesPanelFragment, SearchPatternType } from '../../graphql-operations'
import { EventLogResult } from '../backend'

import { EmptyPanelContainer } from './EmptyPanelContainer'
import { HomePanelsFetchMore, RECENT_SEARCHES_TO_LOAD } from './HomePanels'
import { LoadingPanelView } from './LoadingPanelView'
import { PanelContainer } from './PanelContainer'
import { ShowMoreButton } from './ShowMoreButton'

import styles from './RecentSearchesPanel.module.scss'
import { LocationDescriptor } from 'history'

interface RecentSearch {
    count: number
    searchText: string
    timestamp: string
    url: string
}

interface Props extends TelemetryProps {
    className?: string
    authenticatedUser: AuthenticatedUser | null
    recentSearches: RecentSearchesPanelFragment | null
    /** Function that returns current time (for stability in visual tests). */
    now?: () => Date
    fetchMore: HomePanelsFetchMore
}

export const recentSearchesPanelFragment = gql`
    fragment RecentSearchesPanelFragment on User {
        recentSearchesLogs: eventLogs(first: $firstRecentSearches, eventName: "SearchResultsQueried") {
            nodes {
                argument
                timestamp
                url
            }
            pageInfo {
                hasNextPage
            }
            totalCount
        }
    }
`

export const RecentSearchesPanel: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    className,
    now,
    telemetryService,
    recentSearches,
    fetchMore,
}) => {
    const [searchEventLogs, setSearchEventLogs] = useState<null | RecentSearchesPanelFragment['recentSearchesLogs']>(
        recentSearches?.recentSearchesLogs ?? null
    )
    useEffect(() => setSearchEventLogs(recentSearches?.recentSearchesLogs ?? null), [
        recentSearches?.recentSearchesLogs,
    ])

    const [itemsToLoad, setItemsToLoad] = useState(RECENT_SEARCHES_TO_LOAD)

    const processedResults = useMemo(() => (searchEventLogs === null ? null : processRecentSearches(searchEventLogs)), [
        searchEventLogs,
    ])
    useEffect(() => {
        // Only log the first load (when items to load is equal to the page size)
        if (processedResults && itemsToLoad === RECENT_SEARCHES_TO_LOAD) {
            telemetryService.log(
                'RecentSearchesPanelLoaded',
                { empty: processedResults.length === 0 },
                { empty: processedResults.length === 0 }
            )
        }
    }, [processedResults, telemetryService, itemsToLoad])

    const logSearchClicked = useCallback(() => telemetryService.log('RecentSearchesPanelSearchClicked'), [
        telemetryService,
    ])

    // a constant to hold git commits history
    // get the user's git commit history
    // call streamComputeQuery from stream

    const URL = 'content:output((.|\n)* -> $repo) author:"adham.khalifa@sourcegraph.com" type:commit after:"1 year ago" count:all'
    const computeStreamUrl = '/.api/compute/stream'
    const query = streamComputeQuery(URL, computeStreamUrl)
    // use useobservable
    const gitCommit = useObservable(useMemo(() =>
    streamComputeQuery(URL, computeStreamUrl), [query]))
    //console.log(gitCommit)


    /*
        Get the user’s code search history.
        If the user’s search history is not empty
        Show that
        If the user’s search history is empty
        Check if the user has a git commit history (email)
        Show that
        Else
        Call the empty history view
    */

    const loadingDisplay = <LoadingPanelView text="Loading recent searches" />
    const emptyDisplay = (
        <EmptyPanelContainer className="text-muted">
            <small className="mb-2">
                Your recent searches will be displayed here. Here are a few searches to get you started:
            </small>

            <ul className={styles.examplesList}>
                <li className={styles.examplesListItem}>
                    <small>
                        <Link
                            to={
                                '/search?' +
                                buildSearchURLQuery(
                                    'lang:c if(:[eval_match]) { :[statement_match] }',
                                    SearchPatternType.structural,
                                    false
                                )
                            }
                            className="text-monospace"
                        >
                            <SyntaxHighlightedSearchQuery query="lang:c if(:[eval_match]) { :[statement_match] }" />
                        </Link>
                    </small>
                </li>
                <li className={styles.examplesListItem}>
                    <small>
                        <Link
                            to={
                                '/search?' +
                                buildSearchURLQuery(
                                    'lang:java type:diff after:"1 week ago"',
                                    SearchPatternType.literal,
                                    false
                                )
                            }
                            className="text-monospace"
                        >
                            <SyntaxHighlightedSearchQuery query='lang:java type:diff after:"1 week ago"' />
                        </Link>
                    </small>
                </li>
                <li className={styles.examplesListItem}>
                    <small>
                        <Link
                            to={'/search?' + buildSearchURLQuery('lang:java', SearchPatternType.literal, false)}
                            className="text-monospace"
                        >
                            <SyntaxHighlightedSearchQuery query="lang:java" />
                        </Link>
                    </small>
                </li>
            </ul>
        </EmptyPanelContainer>
    )

    async function loadMoreItems(): Promise<void> {
        telemetryService.log('RecentSearchesPanelShowMoreClicked')
        const newItemsToLoad = itemsToLoad + RECENT_SEARCHES_TO_LOAD
        setItemsToLoad(newItemsToLoad)

        const { data } = await fetchMore({
            firstRecentSearches: newItemsToLoad,
        })

        if (data === undefined) {
            return
        }
        const node = data.node
        if (node === null || node.__typename !== 'User') {
            return
        }
        setSearchEventLogs(node.recentSearchesLogs)
    }

    const contentDisplay = (
        <>
            <table className={classNames('mt-2', styles.resultsTable)}>
                <thead>
                    <tr className={styles.resultsTableRow}>
                        <th>
                            <small>Search</small>
                        </th>
                        <th>
                            <small>Date</small>
                        </th>
                    </tr>
                </thead>
                <tbody>
                    {processedResults?.map((recentSearch, index) => (
                        <tr key={index} className={styles.resultsTableRow}>
                            <td>
                                <small className={styles.recentQuery}>
                                    <Link to={recentSearch.url} onClick={logSearchClicked}>
                                        <SyntaxHighlightedSearchQuery query={recentSearch.searchText} />
                                    </Link>
                                </small>
                            </td>
                            <td className={styles.resultsTableDateCol}>
                                <Timestamp noAbout={true} date={recentSearch.timestamp} now={now} strict={true} />
                            </td>
                        </tr>
                     ))}
                </tbody>
            </table>
            {searchEventLogs?.pageInfo.hasNextPage && <ShowMoreButton onClick={loadMoreItems} />}
        </>
    )

    // A new display for git commits
    // convert gitCommit to a list of git commits
    const gitCommitDisplay = (
        <>
            <table className={classNames('mt-2', styles.resultsTable)}>
                <thead>
                    <tr className={styles.resultsTableRow}>
                        <th>
                            <small>Commit</small>
                        </th>
                        <th>
                            <small>Date</small>
                        </th>
                    </tr>
                </thead>
                <tbody>
                    {((gitCommit: { url: LocationDescriptor<any>; searchText: string; timestamp: string | number | Date }, index: React.Key | null | undefined) => (
                        <tr key={index} className={styles.resultsTableRow}>
                            <td>
                                <small className={styles.recentQuery}>
                                    <Link to={gitCommit.url} onClick={logSearchClicked}>
                                        <SyntaxHighlightedSearchQuery query={gitCommit.searchText} />
                                    </Link>
                                </small>
                            </td>
                            <td className={styles.resultsTableDateCol}>
                                <Timestamp noAbout={true} date={gitCommit.timestamp} now={now} strict={true} />
                            </td>
                        </tr>
                        ))}
                </tbody>
            </table>
            {searchEventLogs?.pageInfo.hasNextPage && <ShowMoreButton onClick={loadMoreItems} />}
        </>
    )

    // get the user's code search history
    const codeSearchHistory = processedResults?.map((recentSearch) => recentSearch.searchText)

    // if the user's code search history is not empty
    // show that
    // if the user's code search history is empty
    // check if the user has a git commit history (email)
    // show that
    // else
    // call the empty history view

    if (processedResults !== null && processedResults.length > 0) {
        // return the contentDisplay
        return contentDisplay
        // check if the user has a git commit history (email)
    } else if (processedResults !== null && processedResults.length === 0 && gitCommit !== null) {
        // return the contentDisplay
        gitCommitDisplay
    }


    return (
        <PanelContainer
            className={classNames(className, 'recent-searches-panel')}
            title="Recent searches"
            state={processedResults ? (processedResults.length > 0 ? 'populated' : 'empty') : 'loading'}
            loadingContent={loadingDisplay}
            populatedContent={contentDisplay}
            emptyContent={emptyDisplay}
        />
    )
}



function processRecentSearches(eventLogResult?: EventLogResult): RecentSearch[] | null {
    if (!eventLogResult) {
        return null
    }

    const recentSearches: RecentSearch[] = []

    for (const node of eventLogResult.nodes) {
        if (node.argument && node.url) {
            const parsedArguments = JSON.parse(node.argument)
            const searchText: string | undefined = parsedArguments?.code_search?.query_data?.combined

            if (searchText) {
                if (recentSearches.length > 0 && recentSearches[recentSearches.length - 1].searchText === searchText) {
                    recentSearches[recentSearches.length - 1].count += 1
                } else {
                    const parsedUrl = new URL(node.url)
                    recentSearches.push({
                        count: 1,
                        url: parsedUrl.pathname + parsedUrl.search, // Strip domain from URL so clicking on it doesn't reload page
                        searchText,
                        timestamp: node.timestamp,
                    })
                }
            }
        }
    }

    return recentSearches
}

