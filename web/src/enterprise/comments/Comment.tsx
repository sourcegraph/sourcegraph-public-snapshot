import PencilIcon from 'mdi-react/PencilIcon'
import React, { useCallback, useState } from 'react'
import { Markdown } from '../../../../shared/src/components/Markdown'
import * as GQL from '../../../../shared/src/graphql/schema'
import { Timestamp } from '../../components/time/Timestamp'
import { PersonLink } from '../../user/PersonLink'

interface Props {
    comment: GQL.IComment
    createdVerb?: string

    className?: string
}

/**
 * A comment with a rendered body and an edit mode.
 */
export const Comment: React.FunctionComponent<Props> = ({ comment, createdVerb = 'commented', className = '' }) => {
    const [isEditing, setIsEditing] = useState(false)
    const onEditClick = useCallback(() => setIsEditing(true), [])
    return (
        <div className={`comment card ${className}`}>
            <div className="card-header d-flex align-items-center justify-content-between">
                <span>
                    <strong>
                        <PersonLink user={comment.author as GQL.IUser /* TODO!(sqs) */} />
                    </strong>{' '}
                    {createdVerb} <Timestamp date={comment.createdAt} />{' '}
                    {comment.updatedAt !== comment.createdAt && <> &bull; edited</>}
                </span>
                <button className="btn btn-sm btn-link text-muted" onClick={onEditClick}>
                    <PencilIcon className="icon-inline" />
                </button>
            </div>
            {!isEditing ? (
                <div className="card-body">
                    {comment.bodyHTML ? (
                        <Markdown dangerousInnerHTML={comment.bodyHTML} />
                    ) : (
                        <span className="text-muted">(empty)</span>
                    )}
                </div>
            ) : (
                a
            )}
        </div>
    )
}
