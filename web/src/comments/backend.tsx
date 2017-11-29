import { Observable } from 'rxjs/Observable'
import { map } from 'rxjs/operators/map'
import { mutateGraphQL, queryGraphQL } from '../backend/graphql'

export const EPERMISSIONDENIED = 'EPERMISSIONDENIED'
class PermissionDeniedError extends Error {
    public readonly code = EPERMISSIONDENIED
    constructor() {
        super(`permission denied`)
    }
}

const gqlThread = `
    id
    repo {
        id
        remoteUri
    }
    file
    repoRevision
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
        author{
            displayName
            username
            avatarURL
        }
        createdAt
        contents
    }
`

/**
 * Fetches shared item by ULID
 *
 * @return Observable that emits the item or `null` if it doesn't exist
 */
export function fetchSharedItem(ulid: string, isLightTheme: boolean): Observable<GQL.ISharedItem | null> {
    return queryGraphQL(
        `query SharedItem($ulid: String!, $isLightTheme: Boolean!) {
                root {
                    sharedItem(ulid: $ulid) {
                        author {
                            displayName
                        }
                        public
                        thread {
                            ${gqlThread}
                        }
                        comment {
                            id
                            title
                        }
                    }
                }
            }
        `,
        { ulid, isLightTheme }
    ).pipe(
        map(({ data, errors }) => {
            if (!data || !data.root || errors) {
                // TODO(slimsag): string comparison is bad practice, remove this
                if (errors && errors[0].message === 'permission denied') {
                    throw new PermissionDeniedError()
                }
                throw Object.assign(new Error((errors || []).map(e => e.message).join('\n')), { errors })
            }
            return data.root.sharedItem
        })
    )
}

/**
 * Adds a comment to the specified thread.
 *
 * @return Observable that emits the updated thread.
 */
export function addCommentToThread(
    threadID: number,
    contents: string,
    ulid: string,
    isLightTheme: boolean
): Observable<GQL.ISharedItemThread> {
    return mutateGraphQL(
        `mutation AddCommentToThread($threadID: Int!, $contents: String!, $ulid: String!, $isLightTheme: Boolean!) {
            addCommentToThreadShared(threadID: $threadID, contents: $contents, ulid: $ulid) {
                ${gqlThread}
            }
        }`,
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
