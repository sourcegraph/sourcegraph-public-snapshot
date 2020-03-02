import { map } from 'rxjs/operators'
import { dataOrThrowErrors, gql } from '../../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { queryGraphQL } from '../../../../backend/graphql'
import { Observable } from 'rxjs'

export const queryCampaigns = ({ first, state }: GQL.ICampaignsOnQueryArguments): Observable<GQL.ICampaignConnection> =>
    queryGraphQL(
        gql`
            query Campaigns($first: Int, $state: CampaignState) {
                campaigns(first: $first, state: $state) {
                    nodes {
                        id
                        name
                        description
                        url
                        createdAt
                        closedAt
                        changesets {
                            totalCount
                            nodes {
                                state
                            }
                        }
                        changesetPlans {
                            totalCount
                        }
                    }
                    totalCount
                }
            }
        `,
        { first, state }
    ).pipe(
        map(dataOrThrowErrors),
        map(data => data.campaigns)
    )
