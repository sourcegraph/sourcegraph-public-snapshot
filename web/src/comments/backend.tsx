import { Observable } from 'rxjs/Observable'
import { map } from 'rxjs/operators/map'
import { gql, mutateGraphQL, queryGraphQL } from '../backend/graphql'
import { makeRepoURI, ParsedRepoURI } from '../repo/index'
import { memoizeObservable } from '../util/memoize'

export const EPERMISSIONDENIED = 'EPERMISSIONDENIED'
class PermissionDeniedError extends Error {
    public readonly code = EPERMISSIONDENIED
    constructor() {
        super(`permission denied`)
    }
}

const commonThreadFields = gql`
    id
    databaseID
    file
    repoRevision
    linesRevision
    branch
    title
    startLine
    endLine
    startCharacter
    endCharacter
    rangeLength
    createdAt
    archivedAt
    lines {
        htmlBefore(isLightTheme: $isLightTheme)
        html(isLightTheme: $isLightTheme)
        htmlAfter(isLightTheme: $isLightTheme)
        textBefore
        text
        textAfter
        textSelectionRangeStart
        textSelectionRangeLength
    }
    comments {
        id
        databaseID
        author {
            displayName
            username
            avatarURL
        }
        createdAt
        richHTML
    }
`

const sharedItemThreadFragment = gql`
    fragment SharedItemThreadFields on SharedItemThread {
        ${commonThreadFields}
        repo {
            id
            remoteUri
            repository {
                uri
                viewerCanAdminister
            }
        }
    }
`

const threadFragment = gql`
    fragment ThreadFields on Thread {
        ${commonThreadFields}
        repo {
            id
            canonicalRemoteID
            repository {
                uri
                viewerCanAdminister
            }
        }
    }
`

export function createSharedItemCacheKey(
    parsed: (ParsedRepoURI | { ulid: string }) & { isLightTheme: boolean }
): string {
    return (isULIDObject(parsed) ? parsed.ulid : makeRepoURI(parsed)) + parsed.isLightTheme
}

function isULIDObject(value: any): value is { ulid: string } {
    return value && value.ulid
}

/**
 * Fetches shared item by ULID
 *
 * @return Observable that emits the item or `null` if it doesn't exist
 */
export const fetchSharedItem = memoizeObservable(
    (args: { ulid: string; isLightTheme: boolean }): Observable<GQL.ISharedItem | null> =>
        queryGraphQL(
            gql`
                query SharedItem($ulid: String!, $isLightTheme: Boolean!) {
                    sharedItem(ulid: $ulid) {
                        author {
                            displayName
                        }
                        public
                        thread {
                            ...SharedItemThreadFields
                        }
                        comment {
                            id
                            title
                        }
                    }
                }
                ${sharedItemThreadFragment}
            `,
            args
        ).pipe(
            map(({ data, errors }) => {
                if (!data || errors) {
                    // TODO(slimsag): string comparison is bad practice, remove this
                    if (errors && errors[0].message === 'permission denied') {
                        throw new PermissionDeniedError()
                    }
                    throw Object.assign(new Error((errors || []).map(e => e.message).join('\n')), { errors })
                }
                return data.sharedItem
            })
        ),
    createSharedItemCacheKey
)

/**
 * Fetches thread by ID.
 *
 * @return Observable that emits the thread or `null` if it doesn't exist
 */
export function fetchThread(id: GQLID, isLightTheme: boolean): Observable<GQL.IThread | null> {
    return queryGraphQL(
        gql`
            query Thread($id: ID!, $isLightTheme: Boolean!) {
                node(id: $id) {
                    ...ThreadFields
                }
            }
            ${threadFragment}
        `,
        { id, isLightTheme }
    ).pipe(
        map(({ data, errors }) => {
            if (!data || !data.node || errors) {
                throw Object.assign(new Error((errors || []).map(e => e.message).join('\n')), { errors })
            }
            return (data.node as any) as GQL.IThread | null
        })
    )
}

/**
 * Adds a comment to the specified thread.
 *
 * @return Observable that emits the updated thread.
 */
export function addCommentToThread(
    threadID: GQLID,
    contents: string,
    ulid: string | undefined,
    isLightTheme: boolean
): Observable<GQL.ISharedItemThread | GQL.IThread> {
    if (ulid) {
        return mutateGraphQL(
            gql`
                mutation AddCommentToThread(
                    $threadID: ID!
                    $contents: String!
                    $ulid: String!
                    $isLightTheme: Boolean!
                ) {
                    addCommentToThreadShared(threadID: $threadID, contents: $contents, ulid: $ulid) {
                        ...SharedItemThreadFields
                    }
                }
                ${sharedItemThreadFragment}
            `,
            { threadID, contents, ulid, isLightTheme }
        ).pipe(
            map(({ data, errors }) => {
                if (!data || !data.addCommentToThreadShared) {
                    throw Object.assign(new Error((errors || []).map(e => e.message).join('\n')), { errors })
                }
                return data.addCommentToThreadShared
            })
        )
    }

    return mutateGraphQL(
        gql`
            mutation AddCommentToThread($threadID: ID!, $contents: String!, $isLightTheme: Boolean!) {
                addCommentToThread(threadID: $threadID, contents: $contents) {
                    ...ThreadFields
                }
            }
            ${threadFragment}
        `,
        { threadID, contents, isLightTheme }
    ).pipe(
        map(({ data, errors }) => {
            if (!data || !data.addCommentToThread) {
                throw Object.assign(new Error((errors || []).map(e => e.message).join('\n')), { errors })
            }
            return data.addCommentToThread
        })
    )
}
