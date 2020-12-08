import { diffStatFields, fileDiffFields } from '../../../../backend/diff'
import { gql, dataOrThrowErrors } from '../../../../../../shared/src/graphql/graphql'
import {
    ChangesetSpecFileDiffsVariables,
    ChangesetSpecFileDiffsResult,
    ChangesetSpecFileDiffsFields,
    CampaignSpecChangesetApplyPreviewResult,
    CampaignSpecChangesetApplyPreviewVariables,
} from '../../../../graphql-operations'
import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'
import { requestGraphQL } from '../../../../backend/graphql'

const changesetSpecFieldsFragment = gql`
    fragment ChangesetSpecFields on ChangesetSpec {
        __typename
        ...HiddenChangesetSpecFields
        ...VisibleChangesetSpecFields
    }

    fragment CommonChangesetSpecFields on ChangesetSpec {
        id
        expiresAt
        type
    }

    fragment HiddenChangesetSpecFields on HiddenChangesetSpec {
        ...CommonChangesetSpecFields
    }

    fragment VisibleChangesetSpecFields on VisibleChangesetSpec {
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

const changesetApplyPreviewFieldsFragment = gql`
    fragment ChangesetApplyPreviewFields on ChangesetApplyPreview {
        operations
        delta {
            titleChanged
        }
        changesetSpec {
            ...ChangesetSpecFields
        }
        changeset {
            __typename
            id
            ... on ExternalChangeset {
                title
            }
        }
    }

    ${changesetSpecFieldsFragment}
`

export const queryChangesetApplyPreviews = ({
    campaignSpec,
    first,
    after,
}: CampaignSpecChangesetApplyPreviewVariables): Observable<
    (CampaignSpecChangesetApplyPreviewResult['node'] & { __typename: 'CampaignSpec' })['applyPreview']
> =>
    requestGraphQL<CampaignSpecChangesetApplyPreviewResult, CampaignSpecChangesetApplyPreviewVariables>(
        gql`
            query CampaignSpecChangesetApplyPreview($campaignSpec: ID!, $first: Int, $after: String) {
                node(id: $campaignSpec) {
                    __typename
                    ... on CampaignSpec {
                        applyPreview(first: $first, after: $after) {
                            totalCount
                            pageInfo {
                                endCursor
                                hasNextPage
                            }
                            nodes {
                                ...ChangesetApplyPreviewFields
                            }
                        }
                    }
                }
            }

            ${changesetApplyPreviewFieldsFragment}
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

export const changesetSpecFileDiffsFields = gql`
    fragment ChangesetSpecFileDiffsFields on VisibleChangesetSpec {
        description {
            __typename
            ... on GitBranchChangesetDescription {
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
                    }
                }
            }
        }
    }

    ${fileDiffFields}
`

export const queryChangesetSpecFileDiffs = ({
    changesetSpec,
    first,
    after,
    isLightTheme,
}: ChangesetSpecFileDiffsVariables): Observable<
    (ChangesetSpecFileDiffsFields['description'] & { __typename: 'GitBranchChangesetDescription' })['diff']
> =>
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
            return node.description.diff
        })
    )
