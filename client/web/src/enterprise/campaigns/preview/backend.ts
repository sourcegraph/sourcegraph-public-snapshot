import { diffStatFields } from '../../../backend/diff'
import { gql, dataOrThrowErrors } from '../../../../../shared/src/graphql/graphql'
import {
    Scalars,
    CampaignSpecFields,
    CampaignSpecByIDVariables,
    CampaignSpecByIDResult,
    CreateCampaignVariables,
    CreateCampaignResult,
    ApplyCampaignResult,
    ApplyCampaignVariables,
} from '../../../graphql-operations'
import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'
import { requestGraphQL } from '../../../backend/graphql'

export const viewerCampaignsCodeHostsFragment = gql`
    fragment ViewerCampaignsCodeHostsFields on CampaignsCodeHostConnection {
        totalCount
        nodes {
            externalServiceURL
            externalServiceKind
        }
    }
`

const supersedingCampaignSpecFragment = gql`
    fragment SupersedingCampaignSpecFields on CampaignSpec {
        createdAt
        applyURL
    }
`

export const campaignSpecFragment = gql`
    fragment CampaignSpecFields on CampaignSpec {
        id
        description {
            name
            description
        }
        appliesToCampaign {
            id
            name
            url
        }
        createdAt
        creator {
            username
            url
        }
        expiresAt
        namespace {
            namespaceName
            url
        }
        viewerCanAdminister
        diffStat {
            ...DiffStatFields
        }
        supersedingCampaignSpec {
            ...SupersedingCampaignSpecFields
        }
        viewerCampaignsCodeHosts(onlyWithoutCredential: true) {
            ...ViewerCampaignsCodeHostsFields
        }
    }

    ${viewerCampaignsCodeHostsFragment}

    ${diffStatFields}

    ${supersedingCampaignSpecFragment}
`

export const fetchCampaignSpecById = (campaignSpec: Scalars['ID']): Observable<CampaignSpecFields | null> =>
    requestGraphQL<CampaignSpecByIDResult, CampaignSpecByIDVariables>(
        gql`
            query CampaignSpecByID($campaignSpec: ID!) {
                node(id: $campaignSpec) {
                    __typename
                    ... on CampaignSpec {
                        ...CampaignSpecFields
                    }
                }
            }
            ${campaignSpecFragment}
        `,
        { campaignSpec }
    ).pipe(
        map(dataOrThrowErrors),
        map(({ node }) => {
            if (!node) {
                return null
            }
            if (node.__typename !== 'CampaignSpec') {
                throw new Error(`The given ID is a ${node.__typename}, not a CampaignSpec`)
            }
            return node
        })
    )

export const createCampaign = ({
    campaignSpec,
}: CreateCampaignVariables): Promise<CreateCampaignResult['createCampaign']> =>
    requestGraphQL<CreateCampaignResult, CreateCampaignVariables>(
        gql`
            mutation CreateCampaign($campaignSpec: ID!) {
                createCampaign(campaignSpec: $campaignSpec) {
                    id
                    url
                }
            }
        `,
        { campaignSpec }
    )
        .pipe(
            map(dataOrThrowErrors),
            map(data => data.createCampaign)
        )
        .toPromise()

export const applyCampaign = ({
    campaignSpec,
    campaign,
}: ApplyCampaignVariables): Promise<ApplyCampaignResult['applyCampaign']> =>
    requestGraphQL<ApplyCampaignResult, ApplyCampaignVariables>(
        gql`
            mutation ApplyCampaign($campaignSpec: ID!, $campaign: ID!) {
                applyCampaign(campaignSpec: $campaignSpec, ensureCampaign: $campaign) {
                    id
                    url
                }
            }
        `,
        { campaignSpec, campaign }
    )
        .pipe(
            map(dataOrThrowErrors),
            map(data => data.applyCampaign)
        )
        .toPromise()
