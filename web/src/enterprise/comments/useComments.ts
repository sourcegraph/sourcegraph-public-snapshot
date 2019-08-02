import { useCallback, useEffect, useState } from 'react'
import { map, startWith } from 'rxjs/operators'
import { dataOrThrowErrors, gql } from '../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../shared/src/graphql/schema'
import { asError, ErrorLike } from '../../../../shared/src/util/errors'
import { queryGraphQL } from '../../backend/graphql'

const commentFieldsFragment = gql`
    fragment CommentFields on Comment {
        id
        body
        bodyHTML
        author {
            username
            displayName
            url
        }
        createdAt
        updatedAt
        viewerCanUpdate
    }
`

const LOADING: 'loading' = 'loading'

/**
 * A React hook that observes all comments on a commentable object (queried from the GraphQL API).
 *
 * @param commentable The commentable object whose comments to observe.
 */
export const useComments = (
    commentable: Pick<GQL.Commentable, 'id'>
): [typeof LOADING | Pick<GQL.Commentable, 'viewerCanComment' | 'comments'> | ErrorLike, () => void] => {
    const [updateSequence, setUpdateSequence] = useState(0)
    const incrementUpdateSequence = useCallback(() => setUpdateSequence(updateSequence + 1), [updateSequence])

    const [result, setResult] = useState<
        typeof LOADING | Pick<GQL.Commentable, 'viewerCanComment' | 'comments'> | ErrorLike
    >(LOADING)
    useEffect(() => {
        const subscription = queryGraphQL(
            gql`
                query Commentable($commentable: ID!) {
                    commentable(id: $commentable) {
                        comments {
                            nodes {
                                ...CommentsFields
                            }
                            totalCount
                        }
                        viewerCanComment
                    }
                }
                ${commentFieldsFragment}
            `,
            { commentable: commentable.id }
        )
            .pipe(
                map(dataOrThrowErrors),
                map(data => {
                    if (!data.commentable) {
                        throw new Error('commentable not found')
                    }
                    return data.commentable
                }),
                startWith(LOADING)
            )
            .subscribe(setResult, err => setResult(asError(err)))
        return () => subscription.unsubscribe()
    }, [commentable, updateSequence])
    return [result, incrementUpdateSequence]
}
