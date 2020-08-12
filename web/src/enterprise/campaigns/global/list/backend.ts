import { map } from 'rxjs/operators'
import { dataOrThrowErrors, gql, requestGraphQL } from '../../../../../../shared/src/graphql/graphql'
import { Observable } from 'rxjs'
import {
    CampaignsVariables,
    CampaignsResult,
    CampaignsByUserResult,
    CampaignsByUserVariables,
    CampaignsByOrgResult,
    CampaignsByOrgVariables,
} from '../../../../graphql-operations'

const ListCampaignFragment = gql`
    fragment ListCampaign on Campaign {
        id
        name
        namespace {
            namespaceName
            url
        }
        description
        createdAt
        closedAt
        changesets {
            stats {
                open
                closed
                merged
            }
        }
    }
`

export const queryCampaigns = ({
    first,
    state,
    viewerCanAdminister,
}: Partial<CampaignsVariables>): Observable<CampaignsResult['campaigns']> =>
    requestGraphQL<CampaignsResult, CampaignsVariables>({
        request: gql`
            query Campaigns($first: Int, $state: CampaignState, $viewerCanAdminister: Boolean) {
                campaigns(first: $first, state: $state, viewerCanAdminister: $viewerCanAdminister) {
                    nodes {
                        ...ListCampaign
                    }
                    totalCount
                }
            }

            ${ListCampaignFragment}
        `,
        variables: { first: first ?? null, state: state ?? null, viewerCanAdminister: viewerCanAdminister ?? null },
    }).pipe(
        map(dataOrThrowErrors),
        map(data => data.campaigns)
    )

export const queryCampaignsByUser = ({
    userID,
    first,
    state,
    viewerCanAdminister,
}: CampaignsByUserVariables): Observable<CampaignsResult['campaigns']> =>
    requestGraphQL<CampaignsByUserResult, CampaignsByUserVariables>({
        request: gql`
            query CampaignsByUser($userID: ID!, $first: Int, $state: CampaignState, $viewerCanAdminister: Boolean) {
                node(id: $userID) {
                    __typename
                    ... on User {
                        campaigns(first: $first, state: $state, viewerCanAdminister: $viewerCanAdminister) {
                            nodes {
                                ...ListCampaign
                            }
                            totalCount
                        }
                    }
                }
            }

            ${ListCampaignFragment}
        `,
        variables: { first, state, viewerCanAdminister, userID },
    }).pipe(
        map(dataOrThrowErrors),
        map(data => {
            if (!data.node) {
                throw new Error('User not found')
            }
            if (data.node.__typename !== 'User') {
                throw new Error(`Requested node is a ${data.node.__typename}, not a User`)
            }
            return data.node.campaigns
        })
    )

export const queryCampaignsByOrg = ({
    orgID,
    first,
    state,
    viewerCanAdminister,
}: CampaignsByOrgVariables): Observable<CampaignsResult['campaigns']> =>
    requestGraphQL<CampaignsByOrgResult, CampaignsByOrgVariables>({
        request: gql`
            query CampaignsByOrg($orgID: ID!, $first: Int, $state: CampaignState, $viewerCanAdminister: Boolean) {
                node(id: $orgID) {
                    __typename
                    ... on Org {
                        campaigns(first: $first, state: $state, viewerCanAdminister: $viewerCanAdminister) {
                            nodes {
                                ...ListCampaign
                            }
                            totalCount
                        }
                    }
                }
            }

            ${ListCampaignFragment}
        `,
        variables: { first, state, viewerCanAdminister, orgID },
    }).pipe(
        map(dataOrThrowErrors),
        map(data => {
            if (!data.node) {
                throw new Error('Org not found')
            }
            if (data.node.__typename !== 'Org') {
                throw new Error(`Requested node is a ${data.node.__typename}, not an Org`)
            }
            return data.node.campaigns
        })
    )
