import { useCallback, useEffect, useState } from 'react'
import { map, startWith } from 'rxjs/operators'
import { dataOrThrowErrors, gql } from '../../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { asError, ErrorLike } from '../../../../../../shared/src/util/errors'
import { queryGraphQL } from '../../../../backend/graphql'
import { queryAndFragmentForThread } from '../../../threadlike/util/graphql'

const LOADING: 'loading' = 'loading'

const { fragment, query } = queryAndFragmentForThread()

/**
 * A React hook that observes all threads in a campaign (queried from the GraphQL API).
 *
 * @param campaign The campaign whose threads to observe.
 */
export const useCampaignThreads = (
    campaign: Pick<GQL.ICampaign, 'id'>
): [typeof LOADING | GQL.IThreadConnection | ErrorLike, () => void] => {
    const [updateSequence, setUpdateSequence] = useState(0)
    const incrementUpdateSequence = useCallback(() => setUpdateSequence(updateSequence + 1), [updateSequence])

    const [result, setResult] = useState<typeof LOADING | GQL.IThreadConnection | ErrorLike>(LOADING)
    useEffect(() => {
        const subscription = queryGraphQL(
            gql`
                query CampaignThreadlikes($campaign: ID!) {
                    node(id: $campaign) {
                        __typename
                        ... on Campaign {
                            threads {
                                nodes {
                                    ${query}
                                }
                                totalCount
                            }
                        }
                    }
                }
                ${fragment}
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
            .subscribe(setResult, err => setResult(asError(err)))
        return () => subscription.unsubscribe()
    }, [campaign, updateSequence])
    return [result, incrementUpdateSequence]
}
