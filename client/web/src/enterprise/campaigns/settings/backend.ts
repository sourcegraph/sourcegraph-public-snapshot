import { Observable } from 'rxjs'
import { map, mapTo } from 'rxjs/operators'
import { dataOrThrowErrors, gql } from '../../../../../shared/src/graphql/graphql'
import { requestGraphQL } from '../../../backend/graphql'
import {
    CampaignsCodeHostsFields,
    CreateCampaignsCredentialResult,
    CreateCampaignsCredentialVariables,
    DeleteCampaignsCredentialResult,
    DeleteCampaignsCredentialVariables,
    Scalars,
    UserCampaignsCodeHostsResult,
    UserCampaignsCodeHostsVariables,
} from '../../../graphql-operations'

export const campaignsCredentialFieldsFragment = gql`
    fragment CampaignsCredentialFields on CampaignsCredential {
        id
        createdAt
    }
`

export function createCampaignsCredential(args: CreateCampaignsCredentialVariables): Promise<void> {
    return requestGraphQL<CreateCampaignsCredentialResult, CreateCampaignsCredentialVariables>(
        gql`
            mutation CreateCampaignsCredential(
                $user: ID!
                $credential: String!
                $externalServiceKind: ExternalServiceKind!
                $externalServiceURL: String!
            ) {
                createCampaignsCredential(
                    user: $user
                    credential: $credential
                    externalServiceKind: $externalServiceKind
                    externalServiceURL: $externalServiceURL
                ) {
                    ...CampaignsCredentialFields
                }
            }

            ${campaignsCredentialFieldsFragment}
        `,
        args
    )
        .pipe(map(dataOrThrowErrors), mapTo(undefined))
        .toPromise()
}

export function deleteCampaignsCredential(id: Scalars['ID']): Promise<void> {
    return requestGraphQL<DeleteCampaignsCredentialResult, DeleteCampaignsCredentialVariables>(
        gql`
            mutation DeleteCampaignsCredential($id: ID!) {
                deleteCampaignsCredential(campaignsCredential: $id) {
                    alwaysNil
                }
            }
        `,
        { id }
    )
        .pipe(map(dataOrThrowErrors), mapTo(undefined))
        .toPromise()
}

export const queryUserCampaignsCodeHosts = ({
    user,
    first,
    after,
}: UserCampaignsCodeHostsVariables): Observable<CampaignsCodeHostsFields> =>
    requestGraphQL<UserCampaignsCodeHostsResult, UserCampaignsCodeHostsVariables>(
        gql`
            query UserCampaignsCodeHosts($user: ID!, $first: Int, $after: String) {
                node(id: $user) {
                    __typename
                    ... on User {
                        campaignsCodeHosts(first: $first, after: $after) {
                            ...CampaignsCodeHostsFields
                        }
                    }
                }
            }

            fragment CampaignsCodeHostsFields on CampaignsCodeHostConnection {
                totalCount
                pageInfo {
                    hasNextPage
                    endCursor
                }
                nodes {
                    ...CampaignsCodeHostFields
                }
            }

            fragment CampaignsCodeHostFields on CampaignsCodeHost {
                externalServiceKind
                externalServiceURL
                credential {
                    ...CampaignsCredentialFields
                }
            }

            ${campaignsCredentialFieldsFragment}
        `,
        {
            user,
            first,
            after,
        }
    ).pipe(
        map(dataOrThrowErrors),
        map(data => {
            if (data.node === null) {
                throw new Error('User not found')
            }
            if (data.node.__typename !== 'User') {
                throw new Error(`Node is a ${data.node.__typename}, not a User`)
            }
            return data.node.campaignsCodeHosts
        })
    )
