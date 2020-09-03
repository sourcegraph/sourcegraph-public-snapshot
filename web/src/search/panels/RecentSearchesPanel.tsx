import React, { useState, useEffect } from 'react'
import classNames from 'classnames'
import { AuthenticatedUser } from '../../auth'
import { dataOrThrowErrors, gql, requestGraphQL } from '../../../../shared/src/graphql/graphql'
import { map } from 'rxjs/operators'
import { PanelContainer } from './PanelContainer'
import { RecentSearchesPanelDataResult, RecentSearchesPanelDataVariables } from '../../graphql-operations'
import { Maybe } from '../../../../shared/src/graphql-operations'

interface EventLogResult {
    totalCount: number
    nodes: { argument: Maybe<string>; timestamp: string; url: string }[]
    pageInfo: { endCursor: Maybe<string>; hasNextPage: boolean }
}

const getData = ({ userId, first }: RecentSearchesPanelDataVariables): Promise<EventLogResult> => {
    const result = requestGraphQL<RecentSearchesPanelDataResult, RecentSearchesPanelDataVariables>({
        request: gql`
            query RecentSearchesPanelData($userId: ID!, $first: Int) {
                node(id: $userId) {
                    ... on User {
                        recentSearches: eventLogs(first: $first, eventName: "SearchResultsQueried") {
                            nodes {
                                argument
                                timestamp
                                url
                            }
                            pageInfo {
                                endCursor
                                hasNextPage
                            }
                            totalCount
                        }
                    }
                }
            }
        `,
        variables: { userId, first: first ?? null },
    })

    return result
        .pipe(
            map(dataOrThrowErrors),
            map(
                (data: RecentSearchesPanelDataResult): EventLogResult => {
                    if (!data.node) {
                        throw new Error('User not found')
                    }
                    return data.node.recentSearches
                }
            )
        )
        .toPromise()
}

export const RecentSearchesPanel: React.FunctionComponent<{
    className?: string
    authenticatedUser: AuthenticatedUser | null
}> = ({ className, authenticatedUser }) => {
    const [state, setState] = useState<'loading' | 'populated' | 'empty'>('loading')
    const [eventLogResult, setEventLogResult] = useState<EventLogResult | null>(null)

    useEffect(() => {
        const getDataAsync = async (): Promise<void> => {
            if (!authenticatedUser) {
                return
            }
            const data = await getData({ userId: authenticatedUser.id, first: 100 })
            setEventLogResult(data)
            if (data.totalCount === 0) {
                setState('empty')
            } else {
                setState('populated')
            }
        }

        // eslint-disable-next-line @typescript-eslint/no-floating-promises
        getDataAsync()
    }, [authenticatedUser])

    const loadingDisplay = <div>Loading</div>
    const contentDisplay = <div>{JSON.stringify(eventLogResult)}</div>
    const emptyDisplay = <div>Empty</div>

    return (
        <PanelContainer
            className={classNames(className, 'recent-searches-panel')}
            title="Recent searches"
            state={state}
            loadingContent={loadingDisplay}
            populatedContent={contentDisplay}
            emptyContent={emptyDisplay}
        />
    )
}
