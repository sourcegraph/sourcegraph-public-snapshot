import { diffStatFields, fileDiffFields } from '../../../../backend/diff'
import { gql, dataOrThrowErrors } from '../../../../../../shared/src/graphql/graphql'
import {
    ChangesetSpecFileDiffsVariables,
    ChangesetSpecFileDiffsResult,
    CampaignSpecApplyPreviewConnectionFields,
    CampaignSpecApplyPreviewResult,
    CampaignSpecApplyPreviewVariables,
    ChangesetSpecFileDiffConnectionFields,
} from '../../../../graphql-operations'
import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'
import { requestGraphQL } from '../../../../backend/graphql'

const changesetSpecFieldsFragment = gql`
    fragment CommonChangesetSpecFields on ChangesetSpec {
        type
    }

    fragment HiddenChangesetSpecFields on HiddenChangesetSpec {
        __typename
        id
        ...CommonChangesetSpecFields
    }

    fragment VisibleChangesetSpecFields on VisibleChangesetSpec {
        __typename
        id
        ...CommonChangesetSpecFields
        description {
            __typename
            ...ExistingChangesetReferenceFields
            ...GitBranchChangesetDescriptionFields
        }
    }

    fragment ExistingChangesetReferenceFields on ExistingChangesetReference {
        baseRepository {
            name
            url
        }
        externalID
    }

    fragment GitBranchChangesetDescriptionFields on GitBranchChangesetDescription {
        baseRepository {
            name
            url
        }
        title
        published
        body
        commits {
            author {
                avatarURL
                email
                displayName
                user {
                    username
                    displayName
                    url
                }
            }
            subject
            body
        }
        baseRef
        headRef
        diffStat {
            ...DiffStatFields
        }
    }

    ${diffStatFields}
`

const campaignSpecApplyPreviewConnectionFieldsFragment = gql`
    fragment CampaignSpecApplyPreviewConnectionFields on ChangesetApplyPreviewConnection {
        totalCount
        pageInfo {
            endCursor
            hasNextPage
        }
        nodes {
            ...ChangesetApplyPreviewFields
        }
    }

    fragment ChangesetApplyPreviewFields on ChangesetApplyPreview {
        __typename
        ... on HiddenChangesetApplyPreview {
            ...HiddenChangesetApplyPreviewFields
        }
        ... on VisibleChangesetApplyPreview {
            ...VisibleChangesetApplyPreviewFields
        }
    }

    fragment HiddenChangesetApplyPreviewFields on HiddenChangesetApplyPreview {
        __typename
        targets {
            __typename
            ... on HiddenApplyPreviewTargetsAttach {
                changesetSpec {
                    ...HiddenChangesetSpecFields
                }
            }
            ... on HiddenApplyPreviewTargetsUpdate {
                changesetSpec {
                    ...HiddenChangesetSpecFields
                }
                changeset {
                    id
                }
            }
            ... on HiddenApplyPreviewTargetsDetach {
                changeset {
                    id
                }
            }
        }
    }

    fragment VisibleChangesetApplyPreviewFields on VisibleChangesetApplyPreview {
        __typename
        operations
        delta {
            titleChanged
        }
        targets {
            __typename
            ... on VisibleApplyPreviewTargetsAttach {
                changesetSpec {
                    ...VisibleChangesetSpecFields
                }
            }
            ... on VisibleApplyPreviewTargetsUpdate {
                changesetSpec {
                    ...VisibleChangesetSpecFields
                }
                changeset {
                    id
                    title
                }
            }
            ... on VisibleApplyPreviewTargetsDetach {
                changeset {
                    id
                    title
                    repository {
                        url
                        name
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

    ${changesetSpecFieldsFragment}
`

export const queryChangesetApplyPreview = ({
    campaignSpec,
    first,
    after,
}: CampaignSpecApplyPreviewVariables): Observable<CampaignSpecApplyPreviewConnectionFields> =>
    requestGraphQL<CampaignSpecApplyPreviewResult, CampaignSpecApplyPreviewVariables>(
        gql`
            query CampaignSpecApplyPreview($campaignSpec: ID!, $first: Int, $after: String) {
                node(id: $campaignSpec) {
                    __typename
                    ... on CampaignSpec {
                        applyPreview(first: $first, after: $after) {
                            ...CampaignSpecApplyPreviewConnectionFields
                        }
                    }
                }
            }

            ${campaignSpecApplyPreviewConnectionFieldsFragment}
        `,
        { campaignSpec, first, after }
    ).pipe(
        map(dataOrThrowErrors),
        map(({ node }) => {
            if (!node) {
                throw new Error(`CampaignSpec with ID ${campaignSpec} does not exist`)
            }
            if (node.__typename !== 'CampaignSpec') {
                throw new Error(`The given ID is a ${node.__typename}, not a CampaignSpec`)
            }
            return node.applyPreview
        })
    )

const changesetSpecFileDiffsFields = gql`
    fragment ChangesetSpecFileDiffsFields on VisibleChangesetSpec {
        description {
            __typename
            ... on GitBranchChangesetDescription {
                diff {
                    fileDiffs(first: $first, after: $after) {
                        ...ChangesetSpecFileDiffConnectionFields
                    }
                }
            }
        }
    }

    fragment ChangesetSpecFileDiffConnectionFields on FileDiffConnection {
        nodes {
            ...FileDiffFields
        }
        totalCount
        pageInfo {
            hasNextPage
            endCursor
        }
    }

    ${fileDiffFields}
`

export const queryChangesetSpecFileDiffs = ({
    changesetSpec,
    first,
    after,
    isLightTheme,
}: ChangesetSpecFileDiffsVariables): Observable<ChangesetSpecFileDiffConnectionFields> =>
    requestGraphQL<ChangesetSpecFileDiffsResult, ChangesetSpecFileDiffsVariables>(
        gql`
            query ChangesetSpecFileDiffs($changesetSpec: ID!, $first: Int, $after: String, $isLightTheme: Boolean!) {
                node(id: $changesetSpec) {
                    __typename
                    ...ChangesetSpecFileDiffsFields
                }
            }

            ${changesetSpecFileDiffsFields}
        `,
        { changesetSpec, first, after, isLightTheme }
    ).pipe(
        map(dataOrThrowErrors),
        map(({ node }) => {
            if (!node) {
                throw new Error(`ChangesetSpec with ID ${changesetSpec} does not exist`)
            }
            if (node.__typename !== 'VisibleChangesetSpec') {
                throw new Error(`The given ID is a ${node.__typename}, not a VisibleChangesetSpec`)
            }
            if (node.description.__typename !== 'GitBranchChangesetDescription') {
                throw new Error('The given ChangesetSpec is no GitBranchChangesetDescription')
            }
            return node.description.diff.fileDiffs
        })
    )
