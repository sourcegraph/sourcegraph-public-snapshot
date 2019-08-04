import { useCallback, useEffect, useState } from 'react'
import { map, startWith } from 'rxjs/operators'
import { dataOrThrowErrors, gql } from '../../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { asError, ErrorLike } from '../../../../../../shared/src/util/errors'
import { queryGraphQL } from '../../../../backend/graphql'

const LOADING: 'loading' = 'loading'

type Result = typeof LOADING | GQL.ICampaignBurndownChart | ErrorLike

/**
 * A React hook that observes the burndown chart for a campaign (queried from the GraphQL API).
 *
 * @param campaign The campaign whose burndown chart to observe.
 */
export const useCampaignBurndownChart = (campaign: Pick<GQL.ICampaign, 'id'>): [Result, () => void] => {
    const [updateSequence, setUpdateSequence] = useState(0)
    const incrementUpdateSequence = useCallback(() => setUpdateSequence(updateSequence + 1), [updateSequence])

    const [result, setResult] = useState<Result>(LOADING)
    useEffect(() => {
        const subscription = queryGraphQL(
            gql`
                query CampaignBurndownChart($campaign: ID!) {
                    node(id: $campaign) {
                        __typename
                        ... on Campaign {
                            burndownChart {
                                dates
                                openThreads
                                mergedThreads
                                closedThreads
                                openApprovedThreads
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
                    return data.node.burndownChart
                }),
                startWith(LOADING)
            )
            .subscribe(setResult, err => setResult(asError(err)))
        return () => subscription.unsubscribe()
    }, [campaign, updateSequence])
    return [result, incrementUpdateSequence]
}
