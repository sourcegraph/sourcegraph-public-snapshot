import { map } from 'rxjs/operators'
import { dataOrThrowErrors, gql } from '../../../../../shared/src/graphql/graphql'
import { queryGraphQL, mutateGraphQL } from '../../../backend/graphql'
import { Observable } from 'rxjs'
import {
    ID,
    ICampaign,
    IUpdateCampaignInput,
    ICreateCampaignInput,
    ICampaignPlan,
    IChangesetPlansOnCampaignArguments,
    IChangesetPlanConnection,
    IChangesetsOnCampaignArguments,
    IEmptyResponse,
    IChangesetPlan,
    IExternalChangeset,
} from '../../../../../shared/src/graphql/schema'
import { DiffStatFields, FileDiffHunkRangeFields, PreviewFileDiffFields, FileDiffFields } from '../../../backend/diff'
import { Connection } from '../../../components/FilteredConnection'

export type CampaignType = 'comby' | 'credentials' | 'regexSearchReplace'

const campaignFragment = gql`
    fragment CampaignFields on Campaign {
        __typename
        id
        name
        description
        author {
            username
            avatarURL
        }
        status {
            completedCount
            pendingCount
            state
            errors
        }
        createdAt
        updatedAt
        publishedAt
        closedAt
        viewerCanAdminister
        changesets {
            totalCount
            nodes {
                __typename
                id
                state
                diff {
                    fileDiffs {
                        diffStat {
                            ...DiffStatFields
                        }
                    }
                }
            }
        }
        changesetPlans {
            totalCount
            nodes {
                id
                __typename
                diff {
                    fileDiffs {
                        totalCount
                        diffStat {
                            ...DiffStatFields
                        }
                    }
                }
            }
        }
        plan {
            id
        }
        # TODO move to separate query and configure from/to
        changesetCountsOverTime {
            date
            merged
            closed
            openApproved
            openChangesRequested
            openPending
            total
        }
    }

    ${DiffStatFields}
`

const campaignPlanFragment = gql`
    fragment CampaignPlanFields on CampaignPlan {
        __typename
        id
        status {
            completedCount
            pendingCount
            state
            errors
        }
        changesetPlans {
            totalCount
            nodes {
                id
                __typename
                diff {
                    fileDiffs {
                        totalCount
                        diffStat {
                            ...DiffStatFields
                        }
                    }
                }
            }
        }
    }

    ${DiffStatFields}
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

export async function retryCampaign(campaignID: ID): Promise<void> {
    const result = await mutateGraphQL(
        gql`
            mutation RetryCampaign($campaign: ID!) {
                retryCampaign(campaign: $campaign) {
                    id
                }
            }
        `,
        { campaign: campaignID }
    ).toPromise()
    dataOrThrowErrors(result)
}

export async function closeCampaign(campaign: ID, closeChangesets = false): Promise<void> {
    const result = await mutateGraphQL(
        gql`
            mutation CloseCampaign($campaign: ID!, $closeChangesets: Boolean!) {
                closeCampaign(campaign: $campaign, closeChangesets: $closeChangesets) {
                    id
                }
            }
        `,
        { campaign, closeChangesets }
    ).toPromise()
    dataOrThrowErrors(result)
}

export async function deleteCampaign(campaign: ID, closeChangesets = false): Promise<void> {
    const result = await mutateGraphQL(
        gql`
            mutation DeleteCampaign($campaign: ID!, $closeChangesets: Boolean!) {
                deleteCampaign(campaign: $campaign, closeChangesets: $closeChangesets) {
                    alwaysNil
                }
            }
        `,
        { campaign, closeChangesets }
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

export const fetchCampaignPlanById = (campaignPlan: ID): Observable<ICampaignPlan | null> =>
    queryGraphQL(
        gql`
            query CampaignPlanByID($campaignPlan: ID!) {
                node(id: $campaignPlan) {
                    __typename
                    ... on CampaignPlan {
                        ...CampaignPlanFields
                    }
                }
            }
            ${campaignPlanFragment}
        `,
        { campaignPlan }
    ).pipe(
        map(dataOrThrowErrors),
        map(({ node }) => {
            if (!node) {
                return null
            }
            if (node.__typename !== 'CampaignPlan') {
                throw new Error(`The given ID is a ${node.__typename}, not a CampaignPlan`)
            }
            return node
        })
    )

export const queryChangesets = (
    campaign: ID,
    { first, state, reviewState, checkState }: IChangesetsOnCampaignArguments
): Observable<Connection<IExternalChangeset | IChangesetPlan>> =>
    queryGraphQL(
        gql`
            query CampaignChangesets(
                $campaign: ID!
                $first: Int
                $state: ChangesetState
                $reviewState: ChangesetReviewState
                $checkState: ChangesetCheckState
            ) {
                node(id: $campaign) {
                    __typename
                    ... on Campaign {
                        changesets(first: $first, state: $state, reviewState: $reviewState, checkState: $checkState) {
                            totalCount
                            nodes {
                                __typename
                                id
                                title
                                body
                                state
                                reviewState
                                checkState
                                labels {
                                    text
                                    description
                                    color
                                }
                                repository {
                                    id
                                    name
                                    url
                                }
                                externalURL {
                                    url
                                }
                                externalID
                                createdAt
                                updatedAt
                                diff {
                                    fileDiffs {
                                        nodes {
                                            ...FileDiffFields
                                        }
                                        totalCount
                                        pageInfo {
                                            hasNextPage
                                        }
                                        diffStat {
                                            ...DiffStatFields
                                        }
                                    }
                                }
                            }
                        }
                        changesetPlans(first: $first) {
                            totalCount
                            nodes {
                                __typename
                                id
                                repository {
                                    id
                                    name
                                    url
                                }
                                publicationEnqueued
                                diff {
                                    fileDiffs {
                                        nodes {
                                            ...PreviewFileDiffFields
                                        }
                                        totalCount
                                        pageInfo {
                                            hasNextPage
                                        }
                                        diffStat {
                                            ...DiffStatFields
                                        }
                                    }
                                }
                            }
                        }
                    }
                }
            }

            ${PreviewFileDiffFields}

            ${FileDiffFields}

            ${FileDiffHunkRangeFields}

            ${DiffStatFields}
        `,
        { campaign, first, state, reviewState, checkState }
    ).pipe(
        map(dataOrThrowErrors),
        map(({ node }) => {
            if (!node) {
                throw new Error(`Campaign with ID ${campaign} does not exist`)
            }
            if (node.__typename !== 'Campaign') {
                throw new Error(`The given ID is a ${node.__typename}, not a Campaign`)
            }
            return {
                totalCount: node.changesetPlans.totalCount + node.changesets.totalCount,
                nodes: [...node.changesetPlans.nodes, ...node.changesets.nodes],
            }
        })
    )

export const queryChangesetPlans = (
    campaignPlan: ID,
    { first }: IChangesetPlansOnCampaignArguments
): Observable<IChangesetPlanConnection> =>
    queryGraphQL(
        gql`
            query CampaignChangesets($campaignPlan: ID!, $first: Int) {
                node(id: $campaignPlan) {
                    __typename
                    ... on CampaignPlan {
                        changesetPlans(first: $first) {
                            totalCount
                            nodes {
                                __typename
                                id
                                repository {
                                    id
                                    name
                                    url
                                }
                                publicationEnqueued
                                diff {
                                    fileDiffs {
                                        nodes {
                                            ...PreviewFileDiffFields
                                        }
                                        totalCount
                                        pageInfo {
                                            hasNextPage
                                        }
                                        diffStat {
                                            ...DiffStatFields
                                        }
                                    }
                                }
                            }
                        }
                    }
                }
            }

            ${PreviewFileDiffFields}

            ${FileDiffHunkRangeFields}

            ${DiffStatFields}
        `,
        { campaignPlan, first }
    ).pipe(
        map(dataOrThrowErrors),
        map(({ node }) => {
            if (!node) {
                throw new Error(`CampaignPlan with ID ${campaignPlan} does not exist`)
            }
            if (node.__typename !== 'CampaignPlan') {
                throw new Error(`The given ID is a ${node.__typename}, not a Campaign`)
            }
            return node.changesetPlans
        })
    )

export async function publishCampaign(campaign: ID): Promise<ICampaign> {
    const result = await mutateGraphQL(
        gql`
            mutation PublishCampaign($campaign: ID!) {
                publishCampaign(campaign: $campaign) {
                    ...CampaignFields
                }
            }
            ${campaignFragment}
        `,
        { campaign }
    ).toPromise()
    return dataOrThrowErrors(result).publishCampaign
}

export async function publishChangeset(changesetPlan: ID): Promise<IEmptyResponse> {
    const result = await mutateGraphQL(
        gql`
            mutation PublishChangeset($changesetPlan: ID!) {
                publishChangeset(changesetPlan: $changesetPlan) {
                    alwaysNil
                }
            }
        `,
        { changesetPlan }
    ).toPromise()
    return dataOrThrowErrors(result).publishChangeset
}

export async function syncChangeset(changeset: ID): Promise<void> {
    const result = await mutateGraphQL(
        gql`
            mutation SyncChangeset($changeset: ID!) {
                syncChangeset(changeset: $changeset) {
                    alwaysNil
                }
            }
        `,
        { changeset }
    ).toPromise()
    dataOrThrowErrors(result)
}
