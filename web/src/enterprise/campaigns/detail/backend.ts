import { map } from 'rxjs/operators'
import { dataOrThrowErrors, gql, requestGraphQL } from '../../../../../shared/src/graphql/graphql'
import { Observable } from 'rxjs'
import { diffStatFields, fileDiffFields } from '../../../backend/diff'
import {
    CampaignFields,
    CampaignByIDResult,
    CampaignChangesetsVariables,
    CampaignChangesetsResult,
    CampaignByIDVariables,
    ExternalChangesetFileDiffsResult,
    ExternalChangesetFileDiffsVariables,
    ExternalChangesetFileDiffsFields,
    SyncChangesetResult,
    SyncChangesetVariables,
    Scalars,
    ChangesetCountsOverTimeVariables,
    ChangesetCountsOverTimeFields,
    ChangesetCountsOverTimeResult,
} from '../../../graphql-operations'

const campaignFragment = gql`
    fragment CampaignFields on Campaign {
        __typename
        id
        name
        namespace {
            namespaceName
            url
        }
        description
        initialApplier {
            username
            url
        }
        createdAt
        updatedAt
        closedAt
        viewerCanAdminister
        changesets {
            stats {
                total
                closed
                merged
                open
                unpublished
            }
        }
        diffStat {
            ...DiffStatFields
        }
    }

    ${diffStatFields}
`

const changesetLabelFragment = gql`
    fragment ChangesetLabelFields on ChangesetLabel {
        color
        description
        text
    }
`

export const fetchCampaignById = (campaign: Scalars['ID']): Observable<CampaignFields | null> =>
    requestGraphQL<CampaignByIDResult, CampaignByIDVariables>({
        request: gql`
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
        variables: { campaign },
    }).pipe(
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

export const hiddenExternalChangesetFieldsFragment = gql`
    fragment HiddenExternalChangesetFields on HiddenExternalChangeset {
        __typename
        id
        createdAt
        updatedAt
        nextSyncAt
        externalState
        publicationState
        reconcilerState
    }
`
export const externalChangesetFieldsFragment = gql`
    fragment ExternalChangesetFields on ExternalChangeset {
        __typename
        id
        title
        body
        publicationState
        reconcilerState
        externalState
        reviewState
        checkState
        error
        labels {
            ...ChangesetLabelFields
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
        diffStat {
            ...DiffStatFields
        }
        createdAt
        updatedAt
        nextSyncAt
    }

    ${diffStatFields}

    ${changesetLabelFragment}
`

export const changesetFieldsFragment = gql`
    fragment ChangesetFields on Changeset {
        __typename
        ... on HiddenExternalChangeset {
            ...HiddenExternalChangesetFields
        }
        ... on ExternalChangeset {
            ...ExternalChangesetFields
        }
    }

    ${hiddenExternalChangesetFieldsFragment}

    ${externalChangesetFieldsFragment}
`

export const queryChangesets = ({
    campaign,
    first,
    externalState,
    reviewState,
    checkState,
}: CampaignChangesetsVariables): Observable<
    (CampaignChangesetsResult['node'] & { __typename: 'Campaign' })['changesets']
> =>
    requestGraphQL<CampaignChangesetsResult, CampaignChangesetsVariables>({
        request: gql`
            query CampaignChangesets(
                $campaign: ID!
                $first: Int
                $externalState: ChangesetExternalState
                $reviewState: ChangesetReviewState
                $checkState: ChangesetCheckState
            ) {
                node(id: $campaign) {
                    __typename
                    ... on Campaign {
                        changesets(
                            first: $first
                            externalState: $externalState
                            reviewState: $reviewState
                            checkState: $checkState
                        ) {
                            totalCount
                            pageInfo {
                                endCursor
                                hasNextPage
                            }
                            nodes {
                                ...ChangesetFields
                            }
                        }
                    }
                }
            }

            ${changesetFieldsFragment}
        `,
        variables: { campaign, first, externalState, reviewState, checkState },
    }).pipe(
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

export async function syncChangeset(changeset: Scalars['ID']): Promise<void> {
    const result = await requestGraphQL<SyncChangesetResult, SyncChangesetVariables>({
        request: gql`
            mutation SyncChangeset($changeset: ID!) {
                syncChangeset(changeset: $changeset) {
                    alwaysNil
                }
            }
        `,
        variables: { changeset },
    }).toPromise()
    dataOrThrowErrors(result)
}

// Because thats the name in the API:
// eslint-disable-next-line unicorn/prevent-abbreviations
export const gitRefSpecFields = gql`
    fragment GitRefSpecFields on GitRevSpec {
        __typename
        ... on GitObject {
            oid
        }
        ... on GitRef {
            target {
                oid
            }
        }
        ... on GitRevSpecExpr {
            object {
                oid
            }
        }
    }
`

export const externalChangesetFileDiffsFields = gql`
    fragment ExternalChangesetFileDiffsFields on ExternalChangeset {
        diff {
            __typename
            ... on RepositoryComparison {
                range {
                    base {
                        ...GitRefSpecFields
                    }
                    head {
                        ...GitRefSpecFields
                    }
                }
                fileDiffs(first: $first, after: $after) {
                    nodes {
                        ...FileDiffFields
                    }
                    totalCount
                    pageInfo {
                        hasNextPage
                        endCursor
                    }
                }
            }
            ... on PreviewRepositoryComparison {
                fileDiffs(first: $first, after: $after) {
                    nodes {
                        ...FileDiffFields
                    }
                    totalCount
                    pageInfo {
                        hasNextPage
                        endCursor
                    }
                }
            }
        }
    }

    ${fileDiffFields}

    ${gitRefSpecFields}
`

export const queryExternalChangesetWithFileDiffs = ({
    externalChangeset,
    first,
    after,
    isLightTheme,
}: ExternalChangesetFileDiffsVariables): Observable<ExternalChangesetFileDiffsFields> =>
    requestGraphQL<ExternalChangesetFileDiffsResult, ExternalChangesetFileDiffsVariables>({
        request: gql`
            query ExternalChangesetFileDiffs(
                $externalChangeset: ID!
                $first: Int
                $after: String
                $isLightTheme: Boolean!
            ) {
                node(id: $externalChangeset) {
                    __typename
                    ...ExternalChangesetFileDiffsFields
                }
            }

            ${externalChangesetFileDiffsFields}
        `,
        variables: { externalChangeset, first, after, isLightTheme },
    }).pipe(
        map(dataOrThrowErrors),
        map(({ node }) => {
            if (!node) {
                throw new Error(`Changeset with ID ${externalChangeset} does not exist`)
            }
            if (node.__typename !== 'ExternalChangeset') {
                throw new Error(`The given ID is a ${node.__typename}, not an ExternalChangeset`)
            }
            return node
        })
    )

const changesetCountsOverTimeFragment = gql`
    fragment ChangesetCountsOverTimeFields on ChangesetCounts {
        date
        merged
        closed
        openApproved
        openChangesRequested
        openPending
        total
    }
`

export const queryChangesetCountsOverTime = ({
    campaign,
}: ChangesetCountsOverTimeVariables): Observable<ChangesetCountsOverTimeFields[]> =>
    requestGraphQL<ChangesetCountsOverTimeResult, ChangesetCountsOverTimeVariables>({
        request: gql`
            query ChangesetCountsOverTime($campaign: ID!) {
                node(id: $campaign) {
                    __typename
                    ... on Campaign {
                        changesetCountsOverTime {
                            ...ChangesetCountsOverTimeFields
                        }
                    }
                }
            }

            ${changesetCountsOverTimeFragment}
        `,
        variables: { campaign },
    }).pipe(
        map(dataOrThrowErrors),
        map(({ node }) => {
            if (!node) {
                throw new Error(`Campaign with ID ${campaign} does not exist`)
            }
            if (node.__typename !== 'Campaign') {
                throw new Error(`The given ID is a ${node.__typename}, not a Campaign`)
            }
            return node.changesetCountsOverTime
        })
    )
