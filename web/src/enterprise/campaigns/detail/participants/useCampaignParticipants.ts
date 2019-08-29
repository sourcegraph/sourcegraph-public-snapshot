import { useEffect, useState } from 'react'
import { map, startWith } from 'rxjs/operators'
import { dataOrThrowErrors, gql } from '../../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { asError, ErrorLike } from '../../../../../../shared/src/util/errors'
import { ActorFragment, ActorQuery } from '../../../../actor/graphql'
import { queryGraphQL } from '../../../../backend/graphql'

const LOADING: 'loading' = 'loading'

type Result = typeof LOADING | GQL.IParticipantConnection | ErrorLike

/**
 * A React hook that observes all participants of a campaign (queried from the GraphQL API).
 *
 * @param campaign The campaign whose participants to observe.
 */
export const useCampaignParticipants = (campaign: Pick<GQL.ICampaign, 'id'>): Result => {
    const [result, setResult] = useState<Result>(LOADING)
    useEffect(() => {
        const subscription = queryGraphQL(
            gql`
                query CampaignParticipants($campaign: ID!) {
                    node(id: $campaign) {
                        __typename
                        ... on Campaign {
                            participants {
                                edges {
                                    actor {
                                        ${ActorQuery}
                                    }
                                    reasons
                                }
                                totalCount
                            }
                        }
                    }
                }
                ${ActorFragment}
            `,
            { campaign: campaign.id }
        )
            .pipe(
                map(dataOrThrowErrors),
                map(data => {
                    if (!data.node || data.node.__typename !== 'Campaign') {
                        throw new Error('not a campaign')
                    }
                    return data.node.participants
                }),
                startWith(LOADING)
            )
            .subscribe(setResult, err => setResult(asError(err)))
        return () => subscription.unsubscribe()
    }, [campaign])
    return result
}
