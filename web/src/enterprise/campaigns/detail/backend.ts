import { map } from 'rxjs/operators'
import { dataOrThrowErrors, gql } from '../../../../../shared/src/graphql/graphql'
import { queryGraphQL, mutateGraphQL } from '../../../backend/graphql'
import { Observable } from 'rxjs'
import {
    Changeset,
    ID,
    ICampaign,
    IChangesetsOnCampaignArguments,
    IExternalChangeset,
} from '../../../../../shared/src/graphql/schema'
import { DiffStatFields, FileDiffFields } from '../../../backend/diff'
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
        branch
        createdAt
        updatedAt
        closedAt
        viewerCanAdminister
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
            total
        }
        diffStat {
            ...DiffStatFields
        }
    }

    ${DiffStatFields}
`

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
    { first, state, reviewState, checkState }: IChangesetsOnCampaignArguments
): Observable<Connection<Changeset>> =>
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

                                state
                                createdAt
                                updatedAt
                                nextSyncAt

                                ... on HiddenExternalChangeset {
                                    id
                                }
                                ... on ExternalChangeset {
                                    id
                                    title
                                    body
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
                                    diff {
                                        __typename
                                        ... on PreviewRepositoryComparison {
                                            fileDiffs {
                                                diffStat {
                                                    ...DiffStatFields
                                                }
                                            }
                                        }
                                        ... on RepositoryComparison {
                                            fileDiffs {
                                                diffStat {
                                                    ...DiffStatFields
                                                }
                                            }
                                        }
                                    }
                                    diffStat {
                                        added
                                        changed
                                        deleted
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

export const queryExternalChangesetWithFileDiffs = (
    externalChangeset: ID,
    { first, after, isLightTheme }: FilteredConnectionQueryArgs & { isLightTheme: boolean }
): Observable<IExternalChangeset> =>
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
                                    diffStat {
                                        ...DiffStatFields
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
                                    diffStat {
                                        ...DiffStatFields
                                    }
                                }
                            }
                        }
                    }
                }
            }

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
            return node
        })
    )
