import { map } from 'rxjs/operators'
import { dataOrThrowErrors, gql } from '../../../../../shared/src/graphql/graphql'
import { queryGraphQL, mutateGraphQL } from '../../../backend/graphql'
import { Observable } from 'rxjs'
import {
    ID,
    ICampaign,
    IUpdateCampaignInput,
    ICreateCampaignInput,
    IChangesetsOnCampaignArguments,
    IEmptyResponse,
    IExternalChangeset,
    IFileDiffConnection,
    IPreviewFileDiffConnection,
    IPatchSet,
    IPatchesOnCampaignArguments,
    IPatchConnection,
} from '../../../../../shared/src/graphql/schema'
import { DiffStatFields, PreviewFileDiffFields, FileDiffFields } from '../../../backend/diff'
import { Connection, FilteredConnectionQueryArgs } from '../../../components/FilteredConnection'

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
        }
        openChangesets {
            totalCount
        }
        patches {
            totalCount
        }
        patchSet {
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
        diffStat {
            ...DiffStatFields
        }
    }

    ${DiffStatFields}
`

const patchSetFragment = gql`
    fragment PatchSetFields on PatchSet {
        __typename
        id
        diffStat {
            ...DiffStatFields
        }
        patches {
            totalCount
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

export async function retryCampaign(campaignID: ID): Promise<ICampaign> {
    const result = await mutateGraphQL(
        gql`
            mutation RetryCampaign($campaign: ID!) {
                retryCampaign(campaign: $campaign) {
                    ...CampaignFields
                }
            }

            ${campaignFragment}
        `,
        { campaign: campaignID }
    ).toPromise()
    return dataOrThrowErrors(result).retryCampaign
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

export const fetchPatchSetById = (patchSet: ID): Observable<IPatchSet | null> =>
    queryGraphQL(
        gql`
            query PatchSetByID($patchSet: ID!) {
                node(id: $patchSet) {
                    __typename
                    ... on PatchSet {
                        ...PatchSetFields
                    }
                }
            }
            ${patchSetFragment}
        `,
        { patchSet }
    ).pipe(
        map(dataOrThrowErrors),
        map(({ node }) => {
            if (!node) {
                return null
            }
            if (node.__typename !== 'PatchSet') {
                throw new Error(`The given ID is a ${node.__typename}, not a PatchSet`)
            }
            return node
        })
    )

export const queryChangesets = (
    campaign: ID,
    { first, state, reviewState, checkState }: IChangesetsOnCampaignArguments
): Observable<Connection<IExternalChangeset>> =>
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
                                head {
                                    abbrevName
                                    target {
                                        oid
                                    }
                                }
                                base {
                                    abbrevName
                                    target {
                                        oid
                                    }
                                }
                                externalID
                                createdAt
                                updatedAt
                                nextSyncAt
                                diff {
                                    fileDiffs {
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
            return node.changesets
        })
    )
export const queryPatchesFromCampaign = (
    campaign: ID,
    { first }: IPatchesOnCampaignArguments
): Observable<IPatchConnection> =>
    queryGraphQL(
        gql`
            query CampaignPatches($campaign: ID!, $first: Int) {
                node(id: $campaign) {
                    __typename
                    ... on Campaign {
                        patches(first: $first) {
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

            ${DiffStatFields}
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
            return node.patches
        })
    )

export const queryPatchesFromPatchSet = (
    patchSet: ID,
    { first }: IPatchesOnCampaignArguments
): Observable<IPatchConnection> =>
    queryGraphQL(
        gql`
            query PatchSetPatches($patchSet: ID!, $first: Int) {
                node(id: $patchSet) {
                    __typename
                    ... on PatchSet {
                        patches(first: $first) {
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

            ${DiffStatFields}
        `,
        { patchSet, first }
    ).pipe(
        map(dataOrThrowErrors),
        map(({ node }) => {
            if (!node) {
                throw new Error(`PatchSet with ID ${patchSet} does not exist`)
            }
            if (node.__typename !== 'PatchSet') {
                throw new Error(`The given ID is a ${node.__typename}, not a PatchSet`)
            }
            return node.patches
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

export async function publishChangeset(patch: ID): Promise<IEmptyResponse> {
    const result = await mutateGraphQL(
        gql`
            mutation PublishChangeset($patch: ID!) {
                publishChangeset(patch: $patch) {
                    alwaysNil
                }
            }
        `,
        { patch }
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

export const queryExternalChangesetFileDiffs = (
    externalChangeset: ID,
    { first, after, isLightTheme }: FilteredConnectionQueryArgs & { isLightTheme: boolean }
): Observable<IFileDiffConnection> =>
    queryGraphQL(
        gql`
            query ExternalChangesetFileDiffs(
                $externalChangeset: ID!
                $first: Int
                $after: String
                $isLightTheme: Boolean!
            ) {
                node(id: $externalChangeset) {
                    __typename
                    ... on ExternalChangeset {
                        diff {
                            fileDiffs(first: $first, after: $after) {
                                nodes {
                                    ...FileDiffFields
                                }
                                totalCount
                                pageInfo {
                                    hasNextPage
                                    endCursor
                                }
                                diffStat {
                                    ...DiffStatFields
                                }
                            }
                        }
                    }
                }
            }

            ${FileDiffFields}

            ${DiffStatFields}
        `,
        { externalChangeset, first, after, isLightTheme }
    ).pipe(
        map(dataOrThrowErrors),
        map(({ node }) => {
            if (!node) {
                throw new Error(`Changeset with ID ${externalChangeset} does not exist`)
            }
            if (node.__typename !== 'ExternalChangeset') {
                throw new Error(`The given ID is a ${node.__typename}, not an ExternalChangeset`)
            }
            if (!node.diff) {
                throw new Error('The given Changeset has no diff')
            }
            return node.diff.fileDiffs
        })
    )

export const queryPatchFileDiffs = (
    patch: ID,
    { first, after, isLightTheme }: FilteredConnectionQueryArgs & { isLightTheme: boolean }
): Observable<IPreviewFileDiffConnection> =>
    queryGraphQL(
        gql`
            query PatchFileDiffs($patch: ID!, $first: Int, $after: String, $isLightTheme: Boolean!) {
                node(id: $patch) {
                    __typename
                    ... on Patch {
                        diff {
                            fileDiffs(first: $first, after: $after) {
                                nodes {
                                    ...PreviewFileDiffFields
                                }
                                totalCount
                                pageInfo {
                                    hasNextPage
                                    endCursor
                                }
                                diffStat {
                                    ...DiffStatFields
                                }
                            }
                        }
                    }
                }
            }

            ${PreviewFileDiffFields}

            ${DiffStatFields}
        `,
        { patch, first, after, isLightTheme }
    ).pipe(
        map(dataOrThrowErrors),
        map(({ node }) => {
            if (!node) {
                throw new Error(`Patch with ID ${patch} does not exist`)
            }
            if (node.__typename !== 'Patch') {
                throw new Error(`The given ID is a ${node.__typename}, not a Patch`)
            }
            if (!node.diff) {
                throw new Error('The given Patch has no diff')
            }
            return node.diff.fileDiffs
        })
    )
