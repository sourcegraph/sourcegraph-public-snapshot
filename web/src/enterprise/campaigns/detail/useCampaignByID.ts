import { useCallback, useEffect, useState } from 'react'
import { map, startWith } from 'rxjs/operators'
import { dataOrThrowErrors, gql } from '../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { asError, ErrorLike, isErrorLike } from '../../../../../shared/src/util/errors'
import { queryGraphQL } from '../../../backend/graphql'

const LOADING: 'loading' = 'loading'

/**
 * A React hook that observes a campaign queried from the GraphQL API by ID.
 *
 * @param campaign The campaign ID.
 */
export const useCampaignByID = (
    campaign: GQL.ID
): [typeof LOADING | GQL.ICampaign | null | ErrorLike, (update?: Partial<GQL.ICampaign>) => void] => {
    const [updateSequence, setUpdateSequence] = useState(0)

    const [result, setResult] = useState<typeof LOADING | GQL.ICampaign | null | ErrorLike>(LOADING)
    useEffect(() => {
        const subscription = queryGraphQL(
            gql`
                query CampaignByID($campaign: ID!) {
                    node(id: $campaign) {
                        __typename
                        ... on Campaign {
                            id
                            name
                            body
                            bodyHTML
                            author {
                                ... on User {
                                    displayName
                                    username
                                    url
                                }
                            }
                            createdAt
                            updatedAt
                            viewerCanUpdate
                            isPreview
                            rules
                            url
                        }
                    }
                }
            `,
            { campaign }
        )
            .pipe(
                map(dataOrThrowErrors),
                map(data => {
                    if (!data.node || data.node.__typename !== 'Campaign') {
                        return null
                    }
                    return data.node
                }),
                startWith(LOADING)
            )
            .subscribe(setResult, err => setResult(asError(err)))
        return () => subscription.unsubscribe()
    }, [campaign, updateSequence])

    const onUpdate = useCallback(
        (update?: Partial<GQL.ICampaign>) => {
            if (update && result && result !== LOADING && !isErrorLike(result)) {
                // Apply immediate partial update.
                setResult({ ...result, ...update })
            } else {
                // Fetch from server.
                setUpdateSequence(updateSequence + 1)
            }
        },
        [result, updateSequence]
    )

    return [result, onUpdate]
}
