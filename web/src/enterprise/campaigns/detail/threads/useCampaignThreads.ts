import { useCallback, useEffect, useState } from 'react'
import { map, startWith } from 'rxjs/operators'
import { dataOrThrowErrors, gql } from '../../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { asError, ErrorLike } from '../../../../../../shared/src/util/errors'
import { queryGraphQL } from '../../../../backend/graphql'

const LOADING: 'loading' = 'loading'

/**
 * A React hook that observes all threads in a campaign (queried from the GraphQL API).
 *
 * @param campaign The campaign whose threads to observe.
 */
export const useCampaignThreads = (
    campaign: Pick<GQL.ICampaign, 'id'>
): [typeof LOADING | GQL.IDiscussionThreadConnection | ErrorLike, () => void] => {
    const [updateSequence, setUpdateSequence] = useState(0)
    const incrementUpdateSequence = useCallback(() => setUpdateSequence(updateSequence + 1), [updateSequence])

    const [threadsOrError, setThreadsOrError] = useState<typeof LOADING | GQL.IDiscussionThreadConnection | ErrorLike>(
        LOADING
    )
    useEffect(() => {
        const subscription = queryGraphQL(
            gql`
                query CampaignThreads($campaign: ID!) {
                    node(id: $campaign) {
                        ... on Campaign {
                            threads {
                                nodes {
                                    id
                                    title
                                    url
                                    status
                                    type
                                }
                                totalCount
                            }
                        }
                    }
                }
            `,
            { campaign: campaign.id }
        )
            .pipe(
                map(dataOrThrowErrors),
                map(data => {
                    if (!data.node || data.node.__typename !== 'Campaign') {
                        throw new Error('not a campaign')
                    }
                    return data.node.threads
                }),
                startWith(LOADING)
            )
            .subscribe(setThreadsOrError, err => setThreadsOrError(asError(err)))
        return () => subscription.unsubscribe()
    }, [campaign, updateSequence])
    return [threadsOrError, incrementUpdateSequence]
}
