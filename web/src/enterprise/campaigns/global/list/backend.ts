import { map } from 'rxjs/operators'
import { dataOrThrowErrors, gql } from '../../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { queryGraphQL } from '../../../../backend/graphql'
import { Observable } from 'rxjs'

export const queryCampaigns = ({
    first,
    state,
    hasPatchSet,
}: GQL.ICampaignsOnQueryArguments): Observable<GQL.ICampaignConnection> =>
    queryGraphQL(
        gql`
            query Campaigns($first: Int, $state: CampaignState, $hasPatchSet: Boolean) {
                campaigns(first: $first, state: $state, hasPatchSet: $hasPatchSet) {
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
                        patches {
                            totalCount
                        }
                    }
                    totalCount
                }
            }
        `,
        { first, state, hasPatchSet }
    ).pipe(
        map(dataOrThrowErrors),
        map(data => data.campaigns)
    )

export const queryCampaignsCount = (): Observable<number> =>
    queryGraphQL(
        gql`
            query Campaigns {
                campaigns(first: 1) {
                    totalCount
                }
            }
        `
    ).pipe(
        map(dataOrThrowErrors),
        map(data => data.campaigns.totalCount)
    )
