import { map } from 'rxjs/operators'
import { dataOrThrowErrors, gql } from '../../../../../../shared/src/graphql/graphql'
import { queryGraphQL } from '../../../../backend/graphql'
import { Observable } from 'rxjs'
import { Connection } from '../../../../components/FilteredConnection'
import { CampaignNodeProps } from '../../list/CampaignNode'
import { CampaignsResult, CampaignsVariables } from '../../../../graphql-operations'

export const queryCampaigns = ({
    first,
    state,
    viewerCanAdminister,
}: CampaignsVariables): Observable<Connection<CampaignNodeProps['node']>> =>
    queryGraphQL<CampaignsResult>(
        gql`
            query Campaigns($first: Int, $state: CampaignState, $viewerCanAdminister: Boolean) {
                campaigns(first: $first, state: $state, viewerCanAdminister: $viewerCanAdminister) {
                    nodes {
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
                    totalCount
                }
            }
        `,
        { first, state, viewerCanAdminister }
    ).pipe(
        map(dataOrThrowErrors),
        map(data => data.campaigns)
    )
