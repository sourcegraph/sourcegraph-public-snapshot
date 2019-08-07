import { useCallback, useEffect, useState } from 'react'
import { map, startWith } from 'rxjs/operators'
import { dataOrThrowErrors, gql, queryAndFragmentForUnion } from '../../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { asError, ErrorLike } from '../../../../../../shared/src/util/errors'
import { actorFragment, actorQuery } from '../../../../actor/graphql'
import { queryGraphQL } from '../../../../backend/graphql'
import { queryAndFragmentForThread } from '../../../threadlike/util/graphql'

const LOADING: 'loading' = 'loading'

const { fragment: threadFragment, query: threadQuery } = queryAndFragmentForThread([
    '__typename',
    'id',
    'title',
    'url',
])

const { fragment: eventFragment, query: eventQuery } = queryAndFragmentForUnion<
    GQL.ThreadTimelineItem['__typename'],
    keyof GQL.ThreadTimelineItem
>(
    [
        'CreateThreadEvent',
        'AddThreadToCampaignEvent',
        'RemoveThreadFromCampaignEvent',
        'ReviewEvent',
        'RequestReviewEvent',
        'CloseThreadEvent',
        'ReopenThreadEvent',
        'CommentOnThreadEvent',
    ],
    ['id', 'createdAt'],
    [`actor { ${actorQuery} }`, `thread { ${threadQuery} }`],
    [actorFragment]
)

type Result = typeof LOADING | GQL.IThreadTimelineItemConnection | ErrorLike

/**
 * A React hook that observes all timeline items for a issue (queried from the GraphQL API).
 *
 * @param issue The issue whose timeline items to observe.
 */
export const useIssueTimelineItems = (issue: Pick<GQL.IThread, 'id'>): [Result, () => void] => {
    const [updateSequence, setUpdateSequence] = useState(0)
    const incrementUpdateSequence = useCallback(() => setUpdateSequence(updateSequence + 1), [updateSequence])

    const [result, setResult] = useState<Result>(LOADING)
    useEffect(() => {
        const subscription = queryGraphQL(
            gql`
                query IssueTimelineItems($issue: ID!) {
                    node(id: $issue) {
                        __typename
                        ... on Issue {
                            timelineItems {
                                nodes {
                                    ${eventQuery}
																		... on AddThreadToCampaignEvent {
                                        campaign { name url }
                                    }
                                    ... on RemoveThreadFromCampaignEvent {
                                        campaign { name url }
                                    }
                                }
                                totalCount
                            }
                        }
                    }
                }
								${threadFragment}
								${eventFragment}
            `,
            { issue: issue.id }
        )
            .pipe(
                map(dataOrThrowErrors),
                map(data => {
                    if (!data.node || data.node.__typename !== 'Issue') {
                        throw new Error('not an issue')
                    }
                    return data.node.timelineItems
                }),
                startWith(LOADING)
            )
            .subscribe(setResult, err => setResult(asError(err)))
        return () => subscription.unsubscribe()
    }, [issue, updateSequence])
    return [result, incrementUpdateSequence]
}
