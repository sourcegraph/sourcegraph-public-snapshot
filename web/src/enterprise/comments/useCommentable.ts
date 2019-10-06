import { useCallback, useEffect, useState } from 'react'
import { map, startWith } from 'rxjs/operators'
import { dataOrThrowErrors, gql } from '../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../shared/src/graphql/schema'
import { asError, ErrorLike } from '../../../../shared/src/util/errors'
import { ActorFragment, ActorQuery } from '../../actor/graphql'
import { queryGraphQL } from '../../backend/graphql'

const replyCommentFieldsFragment = gql`
    fragment CommentReplyFields on CommentReply {
        __typename
        id
        body
        bodyHTML
        author {
            ${ActorQuery}
        }
        createdAt
        updatedAt
        viewerCanUpdate
    }
    ${ActorFragment}
`

const LOADING: 'loading' = 'loading'

type Result =
    | typeof LOADING
    | Pick<GQL.Commentable, 'viewerCanComment' | 'viewerCannotCommentReasons' | 'comments'>
    | ErrorLike

/**
 * A React hook that observes all comments on a commentable object (queried from the GraphQL API).
 *
 * @param commentable The commentable object whose comments to observe.
 */
export const useCommentable = (commentable: Pick<GQL.Commentable, 'id'>): [Result, () => void] => {
    const [updateSequence, setUpdateSequence] = useState(0)
    const incrementUpdateSequence = useCallback(() => setUpdateSequence(updateSequence + 1), [updateSequence])

    const [result, setResult] = useState<Result>(LOADING)
    useEffect(() => {
        const subscription = queryGraphQL(
            gql`
                query Commentable($commentable: ID!) {
                    commentable(id: $commentable) {
                        comments {
                            nodes {
                                __typename
                                ... on CommentReply {
                                    ...CommentReplyFields
                                }
                            }
                            totalCount
                        }
                        viewerCanComment
                        viewerCannotCommentReasons
                    }
                }
                ${replyCommentFieldsFragment}
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
