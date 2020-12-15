import { diffStatFields, fileDiffFields } from '../../../../backend/diff'
import { gql, dataOrThrowErrors } from '../../../../../../shared/src/graphql/graphql'
import {
    CampaignSpecChangesetSpecsResult,
    CampaignSpecChangesetSpecsVariables,
    ChangesetSpecFileDiffsVariables,
    ChangesetSpecFileDiffsResult,
    ChangesetSpecFileDiffsFields,
} from '../../../../graphql-operations'
import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'
import { requestGraphQL } from '../../../../backend/graphql'

export const changesetSpecFieldsFragment = gql`
    fragment CommonChangesetSpecFields on ChangesetSpec {
        expiresAt
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

export const queryChangesetSpecs = ({
    campaignSpec,
    first,
    after,
}: CampaignSpecChangesetSpecsVariables): Observable<
    (CampaignSpecChangesetSpecsResult['node'] & { __typename: 'CampaignSpec' })['applyPreview']
> =>
    requestGraphQL<CampaignSpecChangesetSpecsResult, CampaignSpecChangesetSpecsVariables>(
        gql`
            query CampaignSpecChangesetSpecs($campaignSpec: ID!, $first: Int, $after: String) {
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
                # operations
                # delta {
                #     titleChanged
                # }
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
