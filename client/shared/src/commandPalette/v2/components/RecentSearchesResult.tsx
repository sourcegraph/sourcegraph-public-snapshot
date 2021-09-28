import React, { useMemo } from 'react'
import { EMPTY, Observable, of } from 'rxjs'
import { map } from 'rxjs/operators'

import { EventLogsDataResult, EventLogsDataVariables, Scalars } from '../../../graphql-operations'
import { dataOrThrowErrors, gql } from '../../../graphql/graphql'
import { PlatformContext, PlatformContextProps } from '../../../platform/context'
import { useObservable } from '../../../util/useObservable'

interface RecentSearchesResultProps extends PlatformContextProps<'requestGraphQL'> {
    value: string
    onClick: () => void
    getAuthenticatedUserID: Observable<string | null>
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

export const RecentSearchesResult: React.FC<RecentSearchesResultProps> = ({
    value,
    onClick,
    platformContext,
    getAuthenticatedUserID,
}) => {
    console.log('TODO')

    // Need authenticated user, else show message encouraging users to sign up/in.

    const authenticatedUserID = useObservable(getAuthenticatedUserID)

    const recentSearches = useObservable(
        useMemo(() => (authenticatedUserID ? fetchRecentSearches(authenticatedUserID, 10, platformContext) : EMPTY), [
            // TODO: error handling
            authenticatedUserID,
            platformContext,
        ])
    )

    console.log({ authenticatedUserID, recentSearches })

    if (authenticatedUserID === undefined) {
        // loading
    }

    if (authenticatedUserID === null) {
        // tell users to sign up
    }

    return (
        <div>
            <h1>{value}</h1>
        </div>
    )
}
