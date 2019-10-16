import { map } from 'rxjs/operators'
import { dataOrThrowErrors, gql } from '../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { queryGraphQL } from '../../../backend/graphql'
import { useObservable } from '../../../util/useObservable'
import { useMemo } from 'react'

/**
 * A React hook that observes campaigns queried from the GraphQL API.
 *
 * @param _namespace The (optional) namespace in which to observe the campaigns defined. TODO(sqs): This is not yet implemented.
 */
export const useCampaigns = (_namespace?: unknown): undefined | GQL.ICampaignConnection =>
    useObservable(
        useMemo(
            () =>
                queryGraphQL(gql`
                    query Campaigns {
                        campaigns {
                            nodes {
                                id
                                namespace {
                                    namespaceName
                                }
                                name
                                description
                                url
                            }
                            totalCount
                        }
                    }
                `).pipe(
                    map(dataOrThrowErrors),
                    map(data => data.campaigns)
                ),
            []
        )
    )
