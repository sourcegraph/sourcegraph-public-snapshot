import { map } from 'rxjs/operators'
import { dataOrThrowErrors, gql } from '../../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { queryGraphQL } from '../../../../backend/graphql'
import { Observable } from 'rxjs'
import { FilteredConnectionQueryArgs } from '../../../../components/FilteredConnection'

export const queryCampaigns = ({ first }: FilteredConnectionQueryArgs): Observable<GQL.ICampaignConnection> =>
    queryGraphQL(
        gql`
            query Campaigns($first: Int) {
                campaigns(first: $first) {
                    nodes {
                        id
                        namespace {
                            namespaceName
                        }
                        name
                        description
                        url
                        closedAt
                    }
                    totalCount
                }
            }
        `,
        { first }
    ).pipe(
        map(dataOrThrowErrors),
        map(data => data.campaigns)
    )
