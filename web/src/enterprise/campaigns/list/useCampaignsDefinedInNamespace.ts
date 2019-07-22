import { useCallback, useEffect, useState } from 'react'
import { map, startWith } from 'rxjs/operators'
import { dataOrThrowErrors, gql } from '../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { asError, ErrorLike } from '../../../../../shared/src/util/errors'
import { queryGraphQL } from '../../../backend/graphql'

const LOADING: 'loading' = 'loading'

/**
 * A React hook that observes all campaigns queried from the GraphQL API defined in a particular
 * namespace.
 *
 * @param namespace The namespace in which to observe the campaigns defined.
 */
export const useCampaignsDefinedInNamespace = (
    namespace: Pick<GQL.INamespace, 'id'>
): [typeof LOADING | GQL.ICampaignConnection | ErrorLike, () => void] => {
    const [updateSequence, setUpdateSequence] = useState(0)
    const incrementUpdateSequence = useCallback(() => setUpdateSequence(updateSequence + 1), [updateSequence])

    const [campaignsOrError, setCampaignsOrError] = useState<typeof LOADING | GQL.ICampaignConnection | ErrorLike>(
        LOADING
    )
    useEffect(() => {
        const subscription = queryGraphQL(
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
        )
            .pipe(
                map(dataOrThrowErrors),
                map(data => {
                    if (!data.namespace) {
                        throw new Error('not a namespace')
                    }
                    return data.namespace.campaigns
                }),
                startWith(LOADING)
            )
            .subscribe(setCampaignsOrError, err => setCampaignsOrError(asError(err)))
        return () => subscription.unsubscribe()
    }, [namespace, updateSequence])
    return [campaignsOrError, incrementUpdateSequence]
}
