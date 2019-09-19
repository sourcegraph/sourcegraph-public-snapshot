import { map } from 'rxjs/operators'
import { dataOrThrowErrors, gql } from '../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { queryGraphQL } from '../../../backend/graphql'
import { useObservable } from '../../../util/useObservable'
import { useMemo } from 'react'

/**
 * A React hook that fetches a campaign queried from the GraphQL API by ID.
 *
 * @param campaign The campaign ID.
 */
export const useCampaignByID = (campaign: GQL.ID): undefined | GQL.ICampaign | null =>
    useObservable(
        useMemo(
            () =>
                queryGraphQL(
                    gql`
                        query CampaignByID($campaign: ID!) {
                            node(id: $campaign) {
                                __typename
                                ... on Campaign {
                                    id
                                    namespace {
                                        id
                                        namespaceName
                                    }
                                    author {
                                        username
                                        avatarURL
                                    }
                                    name
                                    description
                                    createdAt
                                    updatedAt
                                    url
                                    changesets {
                                        nodes {
                                            id
                                            title
                                            body
                                            state
                                            reviewState
                                            repository {
                                                name
                                                url
                                            }
                                            externalURL {
                                                url
                                            }
                                            createdAt
                                        }
                                    }
                                }
                            }
                        }
                    `,
                    { campaign }
                ).pipe(
                    map(dataOrThrowErrors),
                    map(({ node }) => {
                        if (!node) {
                            return null
                        }
                        if (node.__typename !== 'Campaign') {
                            throw new Error(`The given ID is a ${node.__typename}, not a Campaign`)
                        }
                        return node
                    })
                ),
            [campaign]
        )
    )
