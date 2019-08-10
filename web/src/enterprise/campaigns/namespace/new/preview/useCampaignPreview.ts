import { useEffect, useState } from 'react'
import { map, startWith } from 'rxjs/operators'
import { dataOrThrowErrors, gql } from '../../../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../../../shared/src/graphql/schema'
import { asError, ErrorLike } from '../../../../../../../shared/src/util/errors'
import { queryGraphQL } from '../../../../../backend/graphql'

export const CampaignPreviewFragment = gql`
    fragment CampaignPreviewFragment on CampaignPreview {
        name
        diagnostics {
            nodes {
                type
                data
            }
            totalCount
        }
    }
`

const LOADING: 'loading' = 'loading'

type Result = typeof LOADING | GQL.ICampaignPreview | ErrorLike

/**
 * A React hook that observes a campaign preview queried from the GraphQL API.
 */
export const useCampaignPreview = (input: GQL.ICreateCampaignInput): Result => {
    const [result, setResult] = useState<Result>(LOADING)
    useEffect(() => {
        const subscription = queryGraphQL(
            gql`
                query CampaignPreview($input: CreateCampaignInput!) {
                    campaignPreview(input: $input) {
                        ...CampaignPreviewFragment
                    }
                }
                ${CampaignPreviewFragment}
            `,
            { input }
        )
            .pipe(
                map(dataOrThrowErrors),
                map(data => data.campaignPreview)
            )
            .pipe(startWith(LOADING))
            .subscribe(setResult, err => setResult(asError(err)))
        return () => subscription.unsubscribe()
    }, [input])
    return result
}
