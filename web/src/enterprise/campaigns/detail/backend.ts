import { map } from 'rxjs/operators'
import { dataOrThrowErrors, gql } from '../../../../../shared/src/graphql/graphql'
import { queryGraphQL, mutateGraphQL } from '../../../backend/graphql'
import { Observable } from 'rxjs'
import {
    ID,
    ICampaign,
    IUpdateCampaignInput,
    ICreateCampaignInput,
    IChangesetConnection,
    IChangesetsOnCampaignArguments,
} from '../../../../../shared/src/graphql/schema'

const campaignFragment = gql`
    fragment CampaignFields on Campaign {
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
            totalCount
        }
        # TODO move to separate query and configure from/to
        changesetCountsOverTime {
            date
            merged
            closed
            openApproved
            openChangesRequested
            openPending
        }
    }
`

export async function updateCampaign(update: IUpdateCampaignInput): Promise<ICampaign> {
    const result = await mutateGraphQL(
        gql`
            mutation UpdateCampaign($update: UpdateCampaignInput!) {
                updateCampaign(input: $update) {
                    ...CampaignFields
                }
            }
            ${campaignFragment}
        `,
        { update }
    ).toPromise()
    return dataOrThrowErrors(result).updateCampaign
}

export async function createCampaign(input: ICreateCampaignInput): Promise<ICampaign> {
    const result = await mutateGraphQL(
        gql`
            mutation CreateCampaign($input: CreateCampaignInput!) {
                createCampaign(input: $input) {
                    id
                    url
                }
            }
        `,
        { input }
    ).toPromise()
    return dataOrThrowErrors(result).createCampaign
}

export async function deleteCampaign(campaign: ID): Promise<void> {
    const result = await mutateGraphQL(
        gql`
            mutation DeleteCampaign($campaign: ID!) {
                deleteCampaign(campaign: $campaign) {
                    alwaysNil
                }
            }
        `,
        { campaign }
    ).toPromise()
    dataOrThrowErrors(result)
}

export const fetchCampaignById = (campaign: ID): Observable<ICampaign | null> =>
    queryGraphQL(
        gql`
            query CampaignByID($campaign: ID!) {
                node(id: $campaign) {
                    __typename
                    ... on Campaign {
                        ...CampaignFields
                    }
                }
            }
            ${campaignFragment}
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
    )

export const queryChangesets = (
    campaign: ID,
    { first }: IChangesetsOnCampaignArguments
): Observable<IChangesetConnection> =>
    queryGraphQL(
        gql`
            query CampaignChangesets($campaign: ID!, $first: Int) {
                node(id: $campaign) {
                    __typename
                    ... on Campaign {
                        changesets(first: $first) {
                            totalCount
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
        { campaign, first }
    ).pipe(
        map(dataOrThrowErrors),
        map(({ node }) => {
            if (!node) {
                throw new Error(`Campaign with ID ${campaign} does not exist`)
            }
            if (node.__typename !== 'Campaign') {
                throw new Error(`The given ID is a ${node.__typename}, not a Campaign`)
            }
            return node.changesets
        })
    )
