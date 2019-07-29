import { useEffect, useState } from 'react'
import { map, startWith } from 'rxjs/operators'
import { dataOrThrowErrors, gql } from '../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { asError, ErrorLike } from '../../../../../shared/src/util/errors'
import { queryGraphQL } from '../../../backend/graphql'

const LOADING: 'loading' = 'loading'

/**
 * A React hook that observes campaigns queried from the GraphQL API.
 *
 * @param namespace The (optional) namespace in which to observe the campaigns defined.
 */
export const useCampaigns = (
    namespace?: Pick<GQL.INamespace, 'id'>
): typeof LOADING | GQL.ICampaignConnection | ErrorLike => {
    const [campaigns, setCampaigns] = useState<typeof LOADING | GQL.ICampaignConnection | ErrorLike>(LOADING)
    useEffect(() => {
        const results = namespace
            ? queryGraphQL(
                  gql`
                      query CampaignsDefinedInNamespace($namespace: ID!) {
                          namespace(id: $namespace) {
                              campaigns {
                                  nodes {
                                      id
                                      name
                                      description
                                      url
                                  }
                                  totalCount
                              }
                          }
                      }
                  `,
                  { namespace: namespace.id }
              ).pipe(
                  map(dataOrThrowErrors),
                  map(data => {
                      if (!data.namespace) {
                          throw new Error('not a namespace')
                      }
                      return data.namespace.campaigns
                  })
              )
            : queryGraphQL(gql`
                  query Campaigns {
                      campaigns {
                          nodes {
                              id
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
              )
        const subscription = results.pipe(startWith(LOADING)).subscribe(setCampaigns, err => setCampaigns(asError(err)))
        return () => subscription.unsubscribe()
    }, [namespace])
    return campaigns
}
