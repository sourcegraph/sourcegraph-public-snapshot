import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import H from 'history'
import React, { useCallback } from 'react'
import { ExtensionsControllerProps } from '../../../../shared/src/extensions/controller'
import * as GQL from '../../../../shared/src/graphql/schema'
import { isErrorLike } from '../../../../shared/src/util/errors'
import { Comment } from './Comment'
import { NewCommentReplyForm } from './NewCommentReplyForm'
import { useCommentable } from './useCommentable'

interface Props extends ExtensionsControllerProps {
    commentable: Pick<GQL.Commentable, 'id'>

    className?: string
    history: H.History
}

const LOADING = 'loading' as const

/**
 * A list of comments and a form to add a new reply comment on a commentable.
 */
export const CommentList: React.FunctionComponent<Props> = ({ commentable: commentableId, className, ...props }) => {
    const [commentable, onCommentableUpdate] = useCommentable(commentableId)
    const onCommentUpdate = useCallback(() => onCommentableUpdate(), [onCommentableUpdate])
    return (
        <div className={`campaign-activity ${className}`}>
            {commentable === LOADING ? (
                <LoadingSpinner className="icon-inline" />
            ) : isErrorLike(commentable) ? (
                <div className="alert alert-danger">{commentable.message}</div>
            ) : (
                <>
                    <ol className="list-unstyled mb-0">
                        {commentable.comments.nodes
                            .filter(comment => comment.__typename === 'CommentReply')
                            .map(comment => (
                                <Comment
                                    {...props}
                                    key={comment.id}
                                    comment={comment}
                                    onCommentUpdate={onCommentUpdate}
                                    tag="li"
                                    className="mb-4"
                                />
                            ))}
                    </ol>
                    {/* TODO!(sqs): be consistent about what a comment object is called - a "commentable" or a "comment object"? in the graphql api, and go/ts var names */}
                    {commentable.viewerCanComment ? (
                        <NewCommentReplyForm
                            {...props}
                            commentable={commentableId}
                            onCommentableUpdate={onCommentableUpdate}
                        />
                    ) : (
                        <div className="alert alert-info mt-3">
                            Unable to comment:{' '}
                            {commentable.viewerCannotCommentReasons.join(' ') /* TODO!(sqs): make this a nicer ui */}
                        </div>
                    )}
                </>
            )}
        </div>
    )
}
