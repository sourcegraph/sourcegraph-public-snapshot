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
                $credential: String!
                $externalServiceKind: ExternalServiceKind!
                $externalServiceURL: String!
            ) {
                createCampaignsCredential(
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
    username,
    first,
    after,
}: Partial<UserCampaignsCodeHostsVariables>): Observable<CampaignsCodeHostsFields> =>
    requestGraphQL<UserCampaignsCodeHostsResult, UserCampaignsCodeHostsVariables>(
        gql`
            query UserCampaignsCodeHosts($username: String!, $first: Int, $after: String) {
                user(username: $username) {
                    campaignsCodeHosts(first: $first, after: $after) {
                        ...CampaignsCodeHostsFields
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
            username: username ?? '', // TODO: why is this required?
            first: first ?? null,
            after: after ?? null,
        }
    ).pipe(
        map(dataOrThrowErrors),
        map(data => {
            if (data.user === null) {
                throw new Error('User not found')
            }
            return data.user.campaignsCodeHosts
        })
    )
