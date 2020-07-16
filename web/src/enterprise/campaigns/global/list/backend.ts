import { map } from 'rxjs/operators'
import { dataOrThrowErrors, gql } from '../../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { queryGraphQL } from '../../../../backend/graphql'
import { Observable } from 'rxjs'

export const queryCampaigns = ({
    first,
    state,
    viewerCanAdminister,
}: GQL.ICampaignsOnQueryArguments): Observable<GQL.ICampaignConnection> =>
    queryGraphQL(
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

export const queryCampaignsCount = (): Observable<number> =>
    queryGraphQL(
        gql`
            query CampaignsCount {
                campaigns(first: 1) {
                    totalCount
                }
            }
        `
    ).pipe(
        map(dataOrThrowErrors),
        map(data => data.campaigns.totalCount)
    )
