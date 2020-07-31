import { map } from 'rxjs/operators'
import { dataOrThrowErrors, gql, requestGraphQL } from '../../../../../../shared/src/graphql/graphql'
import { Observable } from 'rxjs'
import { CampaignsVariables, CampaignsResult } from '../../../../graphql-operations'

const ListCampaignFragment = gql`
    fragment ListCampaign on Campaign {
        id
        name
        description
        createdAt
        closedAt
        author {
            username
        }
        changesets {
            stats {
                open
                closed
                merged
            }
        }
    }
`

export const queryCampaigns = ({
    first,
    state,
    viewerCanAdminister,
}: CampaignsVariables): Observable<CampaignsResult['campaigns']> =>
    requestGraphQL<CampaignsResult, CampaignsVariables>({
        request: gql`
            query Campaigns($first: Int, $state: CampaignState, $viewerCanAdminister: Boolean) {
                campaigns(first: $first, state: $state, viewerCanAdminister: $viewerCanAdminister) {
                    nodes {
                        ...ListCampaign
                    }
                    totalCount
                }
            }

            ${ListCampaignFragment}
        `,
        variables: { first, state, viewerCanAdminister },
    }).pipe(
        map(dataOrThrowErrors),
        map(data => data.campaigns)
    )
