import { applyEdits, Edit } from '@sqs/jsonc-parser'
import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'
import { gql } from '../../../shared/src/graphql/graphql'
import * as GQL from '../../../shared/src/graphql/schema'
import { createAggregateError } from '../../../shared/src/util/errors'
import { memoizeObservable } from '../../../shared/src/util/memoizeObservable'
import { mutateGraphQL, queryGraphQL } from '../backend/graphql'
import { ThreadSettings } from '../enterprise/threads/settings'

const discussionCommentFieldsFragment = gql`
    fragment DiscussionCommentFields on DiscussionComment {
        id
        idWithoutKind
        author {
            ...UserFields
        }
        html
        contents
        inlineURL
        createdAt
        updatedAt
        reports
        canReport
        canDelete
        canClearReports
    }
`

export const discussionThreadTargetFieldsFragment = gql`
    fragment DiscussionThreadTargetFields on DiscussionThreadTarget {
        __typename
        ... on DiscussionThreadTargetRepo {
            id
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
            isIgnored
            url
        }
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
        targets {
            nodes {
                __typename
                ...DiscussionThreadTargetFields
            }
            totalCount
        }
        settings
        type
        status
        url
        inlineURL
        createdAt
        updatedAt
        archivedAt
    }

    fragment UserFields on User {
        displayName
        username
        avatarURL
        url
    }
    ${discussionThreadTargetFieldsFragment}
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
 * Updates an existing discussion thread.
 *
 * @return Observable that emits the updated discussion thread and its comments, or `null` if it was deleted.
 */
export async function updateThread(
    input: GQL.IDiscussionThreadUpdateInput & { delete: true }
): Promise<GQL.IDiscussionThread | null>
export async function updateThread(
    input: Pick<GQL.IDiscussionThreadUpdateInput, Exclude<keyof GQL.IDiscussionThreadUpdateInput, 'delete'>>
): Promise<GQL.IDiscussionThread>
export async function updateThread(input: GQL.IDiscussionThreadUpdateInput): Promise<GQL.IDiscussionThread | null> {
    return mutateGraphQL(
        gql`
            mutation UpdateThread($input: DiscussionThreadUpdateInput!) {
                discussions {
                    updateThread(input: $input) {
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
    )
        .pipe(
            map(({ data, errors }) => {
                if (!data || !data.discussions || (errors && errors.length > 0)) {
                    throw createAggregateError(errors)
                }
                return data.discussions.updateThread
            })
        )
        .toPromise()
}

/**
 * Updates the settings of a discussion thread.
 *
 * @return The updated discussion thread.
 */
export async function updateThreadSettings(
    thread: Pick<GQL.IDiscussionThread, 'id' | 'settings'>,
    edits: Edit[]
): Promise<GQL.IDiscussionThread>
export async function updateThreadSettings(
    thread: Pick<GQL.IDiscussionThread, 'id'>,
    // tslint:disable-next-line: unified-signatures
    newSettings: string
): Promise<GQL.IDiscussionThread>
export async function updateThreadSettings(
    thread: Pick<GQL.IDiscussionThread, 'id'>,
    // tslint:disable-next-line: unified-signatures
    newSettings: ThreadSettings
): Promise<GQL.IDiscussionThread>
export async function updateThreadSettings(
    thread: Pick<GQL.IDiscussionThread, 'id'> & { settings?: any },
    arg: string | Edit[] | ThreadSettings
): Promise<GQL.IDiscussionThread> {
    return updateThread({
        threadID: thread.id,
        settings:
            typeof arg === 'string'
                ? arg
                : Array.isArray(arg)
                ? applyEdits(thread.settings, arg)
                : JSON.stringify(arg, null, 2),
    })
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
 * Add a target to an existing thread.
 *
 * @return Observable that emits the target upon success.
 */
export function addTargetToThread(
    args: GQL.IAddTargetToThreadOnDiscussionsMutationArguments
): Observable<GQL.DiscussionThreadTarget> {
    return mutateGraphQL(
        gql`
            mutation AddTargetToThread($threadID: ID!, $target: DiscussionThreadTargetInput!) {
                discussions {
                    addTargetToThread(threadID: $threadID, target: $target) {
                        __typename
                        ...DiscussionThreadTargetFields
                    }
                }
            }
            ${discussionThreadTargetFieldsFragment}
        `,
        args
    ).pipe(
        map(({ data, errors }) => {
            if (!data || !data.discussions || !data.discussions.addTargetToThread) {
                throw createAggregateError(errors)
            }
            return data.discussions.addTargetToThread
        })
    )
}

/**
 * Updates a target in an existing thread.
 *
 * @return Observable that emits the updated target upon success, or null if the target was removed.
 */
export function updateTargetInThread(
    input: Pick<GQL.IDiscussionThreadTargetUpdateInput, Exclude<keyof GQL.IDiscussionThreadTargetUpdateInput, 'delete'>>
): Observable<GQL.DiscussionThreadTarget>
export function updateTargetInThread(input: GQL.IDiscussionThreadTargetUpdateInput & { delete: true }): Observable<null>
export function updateTargetInThread(
    input: GQL.IDiscussionThreadTargetUpdateInput
): Observable<GQL.DiscussionThreadTarget | null> {
    return mutateGraphQL(
        gql`
            mutation UpdateTargetInThread($input: DiscussionThreadTargetUpdateInput!) {
                discussions {
                    updateTargetInThread(input: $input) {
                        __typename
                        ...DiscussionThreadTargetFields
                    }
                }
            }
            ${discussionThreadTargetFieldsFragment}
        `,
        { input }
    ).pipe(
        map(({ data, errors }) => {
            if (!data || !data.discussions || (errors && errors.length > 0)) {
                throw createAggregateError(errors)
            }
            return data.discussions.updateTargetInThread
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
