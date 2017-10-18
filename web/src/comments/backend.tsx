import 'rxjs/add/operator/map'
import { Observable } from 'rxjs/Observable'
import { queryGraphQL } from '../backend/graphql'

/**
 * Fetches shared item by ULID
 *
 * @return Observable that emits the item or `null` if it doesn't exist
 */
export function fetchSharedItem(ulid: string): Observable<GQL.ISharedItem | null> {
    return queryGraphQL(`
        query SharedItem($ulid: String!) {
            root {
                sharedItem(ulid: $ulid) {
                    author {
                        displayName
                    }
                    thread {
                        id
                        repo {
                            id
                            remoteUri
                        }
                        file
                        revision
                        title
                        startLine
                        endLine
                        startCharacter
                        endCharacter
                        rangeLength
                        createdAt
                        archivedAt
                        lines {
                            htmlBefore
                            html
                            htmlAfter
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
                    }
                    comment {
                        id
                    }
                }
            }
        }
    `, { ulid })
        .map(({ data, errors }) => {
            if (!data || !data.root) {
                throw Object.assign(new Error((errors || []).map(e => e.message).join('\n')), { errors })
            }
            return data.root.sharedItem
        })
}
