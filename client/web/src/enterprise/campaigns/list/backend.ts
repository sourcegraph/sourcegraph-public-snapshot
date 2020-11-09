import { map } from 'rxjs/operators'
import { dataOrThrowErrors, gql } from '../../../../../shared/src/graphql/graphql'
import { Observable } from 'rxjs'
import {
    CampaignsVariables,
    CampaignsResult,
    CampaignsByNamespaceResult,
    CampaignsByNamespaceVariables,
} from '../../../graphql-operations'
import { requestGraphQL } from '../../../backend/graphql'

const ListCampaignFragment = gql`
    fragment ListCampaign on Campaign {
        id
        url
        name
        namespace {
            namespaceName
            url
        }
        description
        createdAt
        closedAt
        changesetsStats {
            open
            closed
            merged
        }
    }
`

export const queryCampaigns = ({
    first,
    after,
    state,
    viewerCanAdminister,
}: Partial<CampaignsVariables>): Observable<CampaignsResult['campaigns']> =>
    requestGraphQL<CampaignsResult, CampaignsVariables>(
        gql`
            query Campaigns($first: Int, $after: String, $state: CampaignState, $viewerCanAdminister: Boolean) {
                campaigns(first: $first, after: $after, state: $state, viewerCanAdminister: $viewerCanAdminister) {
                    nodes {
                        ...ListCampaign
                    }
                    pageInfo {
                        endCursor
                        hasNextPage
                    }
                    totalCount
                }
            }

            ${ListCampaignFragment}
        `,
        {
            first: first ?? null,
            after: after ?? null,
            state: state ?? null,
            viewerCanAdminister: viewerCanAdminister ?? null,
        }
    ).pipe(
        map(dataOrThrowErrors),
        map(data => data.campaigns)
    )

export const queryCampaignsByNamespace = ({
    namespaceID,
    first,
    after,
    state,
    viewerCanAdminister,
}: CampaignsByNamespaceVariables): Observable<CampaignsResult['campaigns']> =>
    requestGraphQL<CampaignsByNamespaceResult, CampaignsByNamespaceVariables>(
        gql`
            query CampaignsByNamespace(
                $namespaceID: ID!
                $first: Int
                $after: String
                $state: CampaignState
                $viewerCanAdminister: Boolean
            ) {
                node(id: $namespaceID) {
                    __typename
                    ... on User {
                        campaigns(
                            first: $first
                            after: $after
                            state: $state
                            viewerCanAdminister: $viewerCanAdminister
                        ) {
                            ...CampaignsFields
                        }
                    }
                    ... on Org {
                        campaigns(
                            first: $first
                            after: $after
                            state: $state
                            viewerCanAdminister: $viewerCanAdminister
                        ) {
                            ...CampaignsFields
                        }
                    }
                }
            }

            fragment CampaignsFields on CampaignConnection {
                nodes {
                    ...ListCampaign
                }
                pageInfo {
                    endCursor
                    hasNextPage
                }
                totalCount
            }

            ${ListCampaignFragment}
        `,
        { first, after, state, viewerCanAdminister, namespaceID }
    ).pipe(
        map(dataOrThrowErrors),
        map(data => {
            if (!data.node) {
                throw new Error('Namespace not found')
            }

            if (data.node.__typename !== 'Org' && data.node.__typename !== 'User') {
                throw new Error(`Requested node is a ${data.node.__typename}, not a User or Org`)
            }
            return data.node.campaigns
        })
    )
