import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import H from 'history'
import React, { useCallback } from 'react'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { isErrorLike } from '../../../../../../shared/src/util/errors'
import { Comment } from '../../../comments/Comment'
import { useComments } from '../../../comments/useComments'

interface Props extends ExtensionsControllerProps {
    campaign: Pick<GQL.ICampaign, 'id'>

    className?: string
    history: H.History
}

const LOADING = 'loading' as const

/**
 * The activity related to a campaign.
 */
export const CampaignActivity: React.FunctionComponent<Props> = ({ campaign, className = '', ...props }) => {
    const [comments, onCommentsUpdate] = useComments(campaign)

    const onCommentUpdate = useCallback(() => onCommentsUpdate(), [onCommentsUpdate])
    return (
        <div className={`campaign-activity ${className}`}>
            {comments === LOADING ? (
                <LoadingSpinner className="icon-inline" />
            ) : isErrorLike(comments) ? (
                <div className="alert alert-danger">{comments.message}</div>
            ) : (
                <ol className="list-unstyled">
                    {comments.comments.nodes.map(comment => (
                        <Comment {...props} key={comment.id} comment={comment} onCommentUpdate={onCommentUpdate} />
                    ))}
                </ol>
            )}
        </div>
    )
}
