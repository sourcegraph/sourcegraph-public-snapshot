import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'
import { gql } from '../../../shared/src/graphql/graphql'
import * as GQL from '../../../shared/src/graphql/schema'
import { createAggregateError } from '../../../shared/src/util/errors'
import { memoizeObservable } from '../../../shared/src/util/memoizeObservable'
import { mutateGraphQL, queryGraphQL } from '../backend/graphql'

const discussionCommentFieldsFragment = gql`
    fragment DiscussionCommentFields on DiscussionComment {
        id
        idWithoutKind
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
        idWithoutKind
        author {
            ...UserFields
        }
        title
        targets(first: 1) {
            nodes {
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
            if (!data || !data.discussions || !data.discussions.createThread || (errors && errors.length > 0)) {
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
            if (!data || !data.discussionThreads || (errors && errors.length > 0)) {
                throw createAggregateError(errors)
            }
            return data.discussionThreads
        })
    )
}

/**
 * Fetches a discussion thread and its comments.
 */
export function fetchDiscussionThreadAndComments(threadIDWithoutKind: string): Observable<GQL.IDiscussionThread> {
    return queryGraphQL(
        gql`
            query DiscussionThreadComments($threadIDWithoutKind: String!) {
                discussionThread(idWithoutKind: $threadIDWithoutKind) {
                    ...DiscussionThreadFields
                    comments {
                        totalCount
                        nodes {
                            ...DiscussionCommentFields
                        }
                    }
                }
            }
            ${discussionThreadFieldsFragment}
            ${discussionCommentFieldsFragment}
        `,
        { threadIDWithoutKind }
    ).pipe(
        map(({ data, errors }) => {
            if (!data || !data.discussionThread || (errors && errors.length > 0)) {
                throw createAggregateError(errors)
            }
            return data.discussionThread
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
            if (!data || !data.discussions || !data.discussions.addCommentToThread || (errors && errors.length > 0)) {
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
            if (!data || !data.discussions || !data.discussions.updateComment || (errors && errors.length > 0)) {
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
export const renderMarkdown = memoizeObservable(
    (ctx: { markdown: string; options?: GQL.IMarkdownOptions }): Observable<string> =>
        queryGraphQL(
            gql`
                query RenderMarkdown($markdown: String!, $options: MarkdownOptions) {
                    renderMarkdown(markdown: $markdown, options: $options)
                }
            `,
            ctx
        ).pipe(
            map(({ data, errors }) => {
                if (!data || (errors && errors.length > 0)) {
                    throw createAggregateError(errors)
                }
                return data.renderMarkdown
            })
        ),
    ctx => `${ctx.markdown}:${ctx.options}`
)
