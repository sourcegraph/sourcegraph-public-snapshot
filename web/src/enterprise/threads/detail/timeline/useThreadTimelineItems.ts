import { useCallback, useEffect, useState } from 'react'
import { map, startWith } from 'rxjs/operators'
import { dataOrThrowErrors, gql, queryAndFragmentForUnion } from '../../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { asError, ErrorLike } from '../../../../../../shared/src/util/errors'
import { ActorFragment, ActorQuery } from '../../../../actor/graphql'
import { queryGraphQL } from '../../../../backend/graphql'
import { ThreadFragment } from '../../util/graphql'

const LOADING: 'loading' = 'loading'

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
        'MergeThreadEvent',
        'CloseThreadEvent',
        'ReopenThreadEvent',
        'CommentOnThreadEvent',
        'AddDiagnosticToThreadEvent',
        'RemoveDiagnosticFromThreadEvent',
    ],
    ['id', 'createdAt'],
    [`actor { ${ActorQuery} }`, `thread { ...ThreadFragment }`],
    [ActorFragment, ThreadFragment]
)

type Result = typeof LOADING | GQL.IThreadTimelineItemConnection | ErrorLike

/**
 * A React hook that observes all timeline items for a thread (queried from the GraphQL API).
 *
 * @param thread The thread whose timeline items to observe.
 */
export const useThreadTimelineItems = (thread: Pick<GQL.IThread, 'id'>): [Result, () => void] => {
    const [updateSequence, setUpdateSequence] = useState(0)
    const incrementUpdateSequence = useCallback(() => setUpdateSequence(updateSequence + 1), [updateSequence])

    const [result, setResult] = useState<Result>(LOADING)
    useEffect(() => {
        const subscription = queryGraphQL(
            gql`
                query ThreadTimelineItems($thread: ID!) {
                    node(id: $thread) {
                        __typename
                        ... on Thread {
                            timelineItems {
                                nodes {
                                    ${eventQuery}
									... on AddThreadToCampaignEvent {
                                        campaign { name url }
                                    }
                                    ... on RemoveThreadFromCampaignEvent {
                                        campaign { name url }
                                    }
                                    ... on AddDiagnosticToThreadEvent {
                                        diagnostic { type data }
                                    }
                                    ... on RemoveDiagnosticFromThreadEvent {
                                        diagnostic { type data }
                                    }
                                }
                                totalCount
                            }
                        }
                    }
                }
				${eventFragment}
            `,
            { thread: thread.id }
        )
            .pipe(
                map(dataOrThrowErrors),
                map(data => {
                    if (!data.node || data.node.__typename !== 'Thread') {
                        throw new Error('not an thread')
                    }
                    return data.node.timelineItems
                }),
                startWith(LOADING)
            )
            .subscribe(setResult, err => setResult(asError(err)))
        return () => subscription.unsubscribe()
    }, [thread, updateSequence])
    return [result, incrementUpdateSequence]
}
