import { useCallback, useEffect, useState } from 'react'
import { map, startWith } from 'rxjs/operators'
import { dataOrThrowErrors, gql, queryAndFragmentForUnion } from '../../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { asError, ErrorLike } from '../../../../../../shared/src/util/errors'
import { queryGraphQL } from '../../../../backend/graphql'
import { queryAndFragmentForThreadOrIssueOrChangeset } from '../../../threadlike/util/graphql'

const LOADING: 'loading' = 'loading'

const { fragment: threadFragment, query: threadQuery } = queryAndFragmentForThreadOrIssueOrChangeset([
    '__typename',
    'id',
    'title',
    'url',
])

const { fragment: eventFragment, query: eventQuery } = queryAndFragmentForUnion<
    GQL.Event['__typename'],
    keyof GQL.Event
>(
    ['AddThreadToCampaignEvent', 'CreateThreadEvent', 'RemoveThreadFromCampaignEvent'],
    ['id', 'createdAt'],
    ['actor { ... on User { id username displayName url } }', `thread { ${threadQuery} }`]
)

type Result = typeof LOADING | GQL.IEventConnection | ErrorLike

/**
 * A React hook that observes all timeline items for a campaign (queried from the GraphQL API).
 *
 * @param campaign The campaign whose timeline items to observe.
 */
export const useCampaignTimelineItems = (campaign: Pick<GQL.ICampaign, 'id'>): [Result, () => void] => {
    const [updateSequence, setUpdateSequence] = useState(0)
    const incrementUpdateSequence = useCallback(() => setUpdateSequence(updateSequence + 1), [updateSequence])

    const [result, setResult] = useState<Result>(LOADING)
    useEffect(() => {
        const subscription = queryGraphQL(
            gql`
                query CampaignTimelineItems($campaign: ID!) {
                    node(id: $campaign) {
                        __typename
                        ... on Campaign {
                            timelineItems {
                                nodes {
                                    ${eventQuery}
                                }
                                totalCount
                            }
                        }
                    }
                }
								${threadFragment}
								${eventFragment}
            `,
            { campaign: campaign.id }
        )
            .pipe(
                map(dataOrThrowErrors),
                map(data => {
                    if (!data.node || data.node.__typename !== 'Campaign') {
                        throw new Error('not a campaign')
                    }
                    return data.node.timelineItems
                }),
                startWith(LOADING)
            )
            .subscribe(setResult, err => setResult(asError(err)))
        return () => subscription.unsubscribe()
    }, [campaign, updateSequence])
    return [result, incrementUpdateSequence]
}
