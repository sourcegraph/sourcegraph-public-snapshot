import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'
import { mutateGraphQL, queryGraphQL } from '../backend/graphql'
import { gql } from '../backend/graphql'
import * as GQL from '../backend/graphqlschema'
import { createAggregateError } from '../util/errors'

const discussionCommentFieldsFragment = gql`
    fragment DiscussionCommentFields on DiscussionComment {
        id
        author {
            ...UserFields
        }
        html
        inlineURL
        createdAt
        updatedAt
        reports
        canReport
        canDelete
        canClearReports
    }
`

const discussionThreadFieldsFragment = gql`
    fragment DiscussionThreadFields on DiscussionThread {
        id
        author {
            ...UserFields
        }
        title
        target {
            __typename
            ... on DiscussionThreadTargetRepo {
                repository {
                    name
                }
                path
                branch {
                    displayName
                }
                revision {
                    displayName
                }
                selection {
                    startLine
                    startCharacter
                    endLine
                    endCharacter
                    linesBefore
                    lines
                    linesAfter
                }
            }
        }
        inlineURL
        createdAt
        updatedAt
        archivedAt
    }

    fragment UserFields on User {
        displayName
        username
        avatarURL
    }
`

/**
 * Creates a new discussion thread.
 *
 * @return Observable that emits the new discussion thread.
 */
export function createThread(input: GQL.IDiscussionThreadCreateInput): Observable<GQL.IDiscussionThread> {
    return mutateGraphQL(
        gql`
            mutation CreateThread($input: DiscussionThreadCreateInput!) {
                discussions {
                    createThread(input: $input) {
                        ...DiscussionThreadFields
                        comments {
                            totalCount
                        }
                    }
                }
            }
            ${discussionThreadFieldsFragment}
        `,
        { input }
    ).pipe(
        map(({ data, errors }) => {
            if (!data || !data.discussions || !data.discussions.createThread) {
                throw createAggregateError(errors)
            }
            return data.discussions.createThread
        })
    )
}

/**
 * Fetches discussion threads.
 */
export function fetchDiscussionThreads(opts: {
    first?: number
    query?: string
    threadID?: GQL.ID
    authorUserID?: GQL.ID
    targetRepositoryID?: GQL.ID
    targetRepositoryName?: string
    targetRepositoryGitCloneURL?: string
    targetRepositoryPath?: string
}): Observable<GQL.IDiscussionThreadConnection> {
    return queryGraphQL(
        gql`
            query DiscussionThreads(
                $first: Int
                $query: String
                $threadID: ID
                $authorUserID: ID
                $targetRepositoryID: ID
                $targetRepositoryName: String
                $targetRepositoryGitCloneURL: String
                $targetRepositoryPath: String
            ) {
                discussionThreads(
                    first: $first
                    query: $query
                    threadID: $threadID
                    authorUserID: $authorUserID
                    targetRepositoryID: $targetRepositoryID
                    targetRepositoryName: $targetRepositoryName
                    targetRepositoryGitCloneURL: $targetRepositoryGitCloneURL
                    targetRepositoryPath: $targetRepositoryPath
                ) {
                    totalCount
                    pageInfo {
                        hasNextPage
                    }
                    nodes {
                        ...DiscussionThreadFields
                        comments {
                            totalCount
                        }
                    }
                }
            }
            ${discussionThreadFieldsFragment}
        `,
        opts
    ).pipe(
        map(({ data, errors }) => {
            if (!data || !data.discussionThreads) {
                throw createAggregateError(errors)
            }
            return data.discussionThreads
        })
    )
}

/**
 * Fetches a discussion thread and its comments.
 */
export function fetchDiscussionThreadAndComments(threadID: GQL.ID): Observable<GQL.IDiscussionThread> {
    return queryGraphQL(
        gql`
            query DiscussionThreadComments($threadID: ID!) {
                discussionThreads(threadID: $threadID) {
                    totalCount
                    nodes {
                        ...DiscussionThreadFields
                        comments {
                            totalCount
                            nodes {
                                ...DiscussionCommentFields
                            }
                        }
                    }
                }
            }
            ${discussionThreadFieldsFragment}
            ${discussionCommentFieldsFragment}
        `,
        { threadID }
    ).pipe(
        map(({ data, errors }) => {
            if (
                !data ||
                !data.discussionThreads ||
                !data.discussionThreads.nodes ||
                data.discussionThreads.nodes.length !== 1
            ) {
                throw createAggregateError(errors)
            }
            return data.discussionThreads.nodes[0]
        })
    )
}

/**
 * Adds a comment to an existing discussion thread.
 *
 * @return Observable that emits the updated discussion thread and its comments.
 */
export function addCommentToThread(threadID: GQL.ID, contents: string): Observable<GQL.IDiscussionThread> {
    return mutateGraphQL(
        gql`
            mutation AddCommentToThread($threadID: ID!, $contents: String!) {
                discussions {
                    addCommentToThread(threadID: $threadID, contents: $contents) {
                        ...DiscussionThreadFields
                        comments {
                            totalCount
                            nodes {
                                ...DiscussionCommentFields
                            }
                        }
                    }
                }
            }
            ${discussionThreadFieldsFragment}
            ${discussionCommentFieldsFragment}
        `,
        { threadID, contents }
    ).pipe(
        map(({ data, errors }) => {
            if (!data || !data.discussions || !data.discussions.addCommentToThread) {
                throw createAggregateError(errors)
            }
            return data.discussions.addCommentToThread
        })
    )
}

/**
 * Updates an existing comment in a discussion thread.
 *
 * @return Observable that emits the updated discussion thread and its comments.
 */
export function updateComment(input: GQL.IDiscussionCommentUpdateInput): Observable<GQL.IDiscussionThread> {
    return mutateGraphQL(
        gql`
            mutation UpdateComment($input: DiscussionCommentUpdateInput!) {
                discussions {
                    updateComment(input: $input) {
                        ...DiscussionThreadFields
                        comments {
                            totalCount
                            nodes {
                                ...DiscussionCommentFields
                            }
                        }
                    }
                }
            }
            ${discussionThreadFieldsFragment}
            ${discussionCommentFieldsFragment}
        `,
        { input }
    ).pipe(
        map(({ data, errors }) => {
            if (!data || !data.discussions || !data.discussions.updateComment) {
                throw createAggregateError(errors)
            }
            return data.discussions.updateComment
        })
    )
}

/**
 * Renders Markdown to HTML.
 *
 * @return Observable that emits the HTML string, which is already sanitized and escaped and thus is always safe to render.
 */
export function renderMarkdown(markdown: string, options?: GQL.IMarkdownOptions): Observable<string> {
    return queryGraphQL(
        gql`
            query RenderMarkdown($markdown: String!, $options: MarkdownOptions) {
                renderMarkdown(markdown: $markdown, options: $options)
            }
        `,
        { markdown }
    ).pipe(
        map(({ data, errors }) => {
            if (!data || !data.renderMarkdown) {
                throw createAggregateError(errors)
            }
            return data.renderMarkdown
        })
    )
}
