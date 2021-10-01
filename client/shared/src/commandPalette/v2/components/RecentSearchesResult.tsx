import { sortBy } from 'lodash'
import React, { useMemo } from 'react'
import { useHistory } from 'react-router'
import { EMPTY, Observable, of } from 'rxjs'
import { map } from 'rxjs/operators'
import stringScore from 'string-score'

import { HighlightedMatches } from '../../../components/HighlightedMatches'
import { EventLogsDataResult, EventLogsDataVariables, Scalars } from '../../../graphql-operations'
import { dataOrThrowErrors, gql } from '../../../graphql/graphql'
import { PlatformContext, PlatformContextProps } from '../../../platform/context'
import { useObservable } from '../../../util/useObservable'

import { Message } from './Message'
import { NavigableList } from './NavigableList'
import listStyles from './NavigableList.module.scss'
import styles from './RecentSearchesResult.module.scss'

interface RecentSearchesResultProps
    extends PlatformContextProps<'requestGraphQL' | 'clientApplication' | 'sourcegraphURL'> {
    value: string
    onClick: () => void
    currentUserID: Observable<string | null>
}

export interface EventLogResult {
    totalCount: number
    nodes: { argument: string | null; timestamp: string; url: string }[]
    pageInfo: { hasNextPage: boolean }
}

// TODO: extract and share with web (shared search panels)
const eventLogsDataQuery = gql`
    query EventLogsData($userId: ID!, $first: Int, $eventName: String!) {
        node(id: $userId) {
            ... on User {
                eventLogs(first: $first, eventName: $eventName) {
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
        }
    }
`

function fetchEvents(
    userId: Scalars['ID'],
    first: number,
    eventName: string,
    platformContext: Pick<PlatformContext, 'requestGraphQL'>
): Observable<EventLogResult | null> {
    if (!userId) {
        return of(null)
    }

    const result = platformContext.requestGraphQL<EventLogsDataResult, EventLogsDataVariables>({
        request: eventLogsDataQuery,
        variables: { userId, first: first ?? null, eventName },
        mightContainPrivateInfo: true,
    })

    return result.pipe(
        map(dataOrThrowErrors),
        map(
            (data: EventLogsDataResult): EventLogResult => {
                if (!data.node) {
                    throw new Error('User not found')
                }
                return data.node.eventLogs
            }
        )
    )
}

export function fetchRecentSearches(
    userId: Scalars['ID'],
    first: number,
    platformContext: Pick<PlatformContext, 'requestGraphQL'>
): Observable<EventLogResult | null> {
    return fetchEvents(userId, first, 'SearchResultsQueried', platformContext)
}

// TODO: pagination, narrowing with search should be enough for now
const RESULT_COUNT = 200

export const RecentSearchesResult: React.FC<RecentSearchesResultProps> = ({
    value,
    onClick,
    platformContext,
    currentUserID,
}) => {
    const history = useHistory()

    const authenticatedUserID = useObservable(useMemo(() => currentUserID, [currentUserID]))
    console.log({ authenticatedUserID })
    // TODO: error handling
    const recentSearches = useObservable(
        useMemo(
            () =>
                authenticatedUserID ? fetchRecentSearches(authenticatedUserID, RESULT_COUNT, platformContext) : EMPTY,
            [authenticatedUserID, platformContext]
        )
    )

    if (authenticatedUserID === undefined) {
        return <Message>Loading...</Message>
    }

    if (authenticatedUserID === null) {
        return <Message>Sign in to view recent searches</Message>
    }

    // Navigable list, throttle (w leading n trailing) search query

    // on bext, this will take you to SG! also, the first/default navigable item
    // should be the input value, essentially making this an easy "search on sourcegraph"
    // path!

    const searches = filterAndRankResults(value, (recentSearches && processRecentSearches(recentSearches)) || [])

    const onSearch = (query: string): void => {
        if (platformContext.clientApplication === 'sourcegraph') {
            history.push({
                pathname: '/search',
                search: new URLSearchParams([['q', query]]).toString(),
                state: history.location.state,
            })
        } else {
            window.location.href = `${platformContext.sourcegraphURL}/search?q=${query}`
        }
        onClick()
    }

    return (
        <div>
            <NavigableList items={[null, ...searches]}>
                {(search, { active }) => {
                    if (search === null) {
                        return (
                            <NavigableList.Item active={active} onClick={() => onSearch(value)}>
                                <span className={listStyles.itemContainer}>
                                    <strong className={styles.queryPrompt}>Execute search with query: </strong> {value}
                                </span>
                            </NavigableList.Item>
                        )
                    }

                    return (
                        <NavigableList.Item active={active} onClick={() => onSearch(search.searchText)}>
                            {value ? (
                                <HighlightedMatches
                                    containerClassName={styles.textContainer}
                                    text={search.searchText}
                                    pattern={value}
                                />
                            ) : (
                                search.searchText
                            )}
                        </NavigableList.Item>
                    )
                }}
            </NavigableList>
        </div>
    )
}

// TODO: share with recent searches panel
interface RecentSearch {
    count: number
    searchText: string
    timestamp: string
    url: string
}

function processRecentSearches(eventLogResult?: EventLogResult): RecentSearch[] | null {
    if (!eventLogResult) {
        return null
    }

    const recentSearches: RecentSearch[] = []

    for (const node of eventLogResult.nodes) {
        if (node.argument) {
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

function filterAndRankResults(query: string, recentSearches: RecentSearch[]): RecentSearch[] {
    if (!query) {
        // Should already be sorted by time
        return recentSearches
    }

    const filteredSearches = recentSearches
        .map(search => ({
            ...search,
            score: stringScore(search.searchText, query, 0),
        }))
        .filter(({ score }) => score > 0)

    const sortedSearches = sortBy(filteredSearches, 'score', 'count')

    // Depduplicate by search text
    const usedSearchText = new Set<string>()
    const searches: RecentSearch[] = []
    for (const search of sortedSearches) {
        if (!usedSearchText.has(search.searchText)) {
            searches.push(search)
            usedSearchText.add(search.searchText)
        }
    }
    return searches
}
