import { useCallback, useEffect, useState } from 'react'
import { map, startWith } from 'rxjs/operators'
import { dataOrThrowErrors, gql } from '../../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { asError, ErrorLike } from '../../../../../../shared/src/util/errors'
import { queryGraphQL } from '../../../../backend/graphql'
import { ThreadDiagnosticConnectionFragment } from '../../../threads/detail/diagnostics/useThreadDiagnostics'

const LOADING: 'loading' = 'loading'

type Result = typeof LOADING | GQL.IThreadDiagnosticConnection | ErrorLike

/**
 * A React hook that observes all diagnostics for all threads in a campaign (queried from the GraphQL API).
 *
 * @param campaign The campaign whose diagnostics to observe.
 */
export const useCampaignDiagnostics = (campaign: Pick<GQL.ICampaign, 'id'>): [Result, () => void] => {
    const [updateSequence, setUpdateSequence] = useState(0)
    const incrementUpdateSequence = useCallback(() => setUpdateSequence(updateSequence + 1), [updateSequence])

    const [result, setResult] = useState<Result>(LOADING)
    useEffect(() => {
        const subscription = queryGraphQL(
            gql`
                query CampaignDiagnostics($campaign: ID!) {
                    node(id: $campaign) {
                        __typename
                        ... on Campaign {
                            diagnostics {
                                ...ThreadDiagnosticConnectionFragment
                            }
                        }
                    }
                }
                ${ThreadDiagnosticConnectionFragment}
            `,
            { campaign: campaign.id }
        )
            .pipe(
                map(dataOrThrowErrors),
                map(data => {
                    if (!data.node || data.node.__typename !== 'Campaign') {
                        throw new Error('not a campaign')
                    }
                    return data.node.diagnostics
                }),
                startWith(LOADING)
            )
            .subscribe(setResult, err => setResult(asError(err)))
        return () => subscription.unsubscribe()
    }, [campaign.id, updateSequence])
    return [result, incrementUpdateSequence]
}
