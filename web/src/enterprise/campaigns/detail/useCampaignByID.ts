import { useEffect, useState } from 'react'
import { map, startWith } from 'rxjs/operators'
import { dataOrThrowErrors, gql } from '../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { asError, ErrorLike } from '../../../../../shared/src/util/errors'
import { queryGraphQL } from '../../../backend/graphql'

const LOADING: 'loading' = 'loading'

/**
 * A React hook that observes a campaign queried from the GraphQL API by ID.
 *
 * @param id The scope in which to observe the campaign.
 */
export const useCampaignByID = (id: GQL.ID): typeof LOADING | GQL.ICampaign | null | ErrorLike => {
    const [campaignOrError, setCampaignOrError] = useState<typeof LOADING | GQL.ICampaign | null | ErrorLike>(LOADING)
    useEffect(() => {
        const subscription = queryGraphQL(
            gql`
                query CampaignByID($campaign: ID!) {
                    node(id: $campaign) {
                        __typename
                        ... on Campaign {
                            id
                            name
                            description
                            url
                        }
                    }
                }
            `,
            { campaign: id }
        )
            .pipe(
                map(dataOrThrowErrors),
                map(data => {
                    if (!data.node || data.node.__typename !== 'Campaign') {
                        return null
                    }
                    return data.node
                }),
                startWith(LOADING)
            )
            .subscribe(setCampaignOrError, err => setCampaignOrError(asError(err)))
        return () => subscription.unsubscribe()
    }, [id])
    return campaignOrError
}
