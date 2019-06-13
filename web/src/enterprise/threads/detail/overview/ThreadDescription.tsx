import H from 'history'
import React, { useCallback, useState } from 'react'
import { WithLinkPreviews } from '../../../../../../shared/src/components/linkPreviews/WithLinkPreviews'
import { Markdown } from '../../../../../../shared/src/components/Markdown'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { ErrorLike } from '../../../../../../shared/src/util/errors'
import { LINK_PREVIEW_CLASS } from '../../../../components/linkPreviews/styles'
import { Timestamp } from '../../../../components/time/Timestamp'
import { setElementTooltip } from '../../../../components/tooltip/Tooltip'
import { EditCommentForm } from '../../../../repo/blob/discussions/EditCommentForm'
import { PersonLink } from '../../../../user/PersonLink'

interface Props extends ExtensionsControllerProps {
    thread: GQL.IDiscussionThread
    onThreadUpdate: (thread: GQL.IDiscussionThread) => void

    className?: string
    location: H.Location
    history: H.History
}

export const ThreadDescription: React.FunctionComponent<Props> = ({ thread, onThreadUpdate, className, ...props }) => {
    const comment = thread.comments.nodes[0]

    const [isEditing, setIsEditing] = useState(false)
    const onEditClick = useCallback(() => setIsEditing(true), [])

    const onCommentUpdate = useCallback(
        (thread: GQL.IDiscussionThread | null) => {
            setIsEditing(false)
            if (thread) {
                onThreadUpdate(thread)
            }
        },
        [onThreadUpdate]
    )

    return comment ? (
        <div className={className}>
            {!isEditing ? (
                <>
                    <WithLinkPreviews
                        dangerousInnerHTML={comment.html}
                        extensionsController={props.extensionsController}
                        setElementTooltip={setElementTooltip}
                        linkPreviewContentClass={LINK_PREVIEW_CLASS}
                    >
                        {props => <Markdown {...props} />}
                    </WithLinkPreviews>
                    <div className="d-flex align-items-center mb-2">
                        <small className="text-muted">
                            Edited by{' '}
                            <strong>
                                <PersonLink user={comment.author} />
                            </strong>{' '}
                            <Timestamp date={comment.updatedAt} />
                        </small>
                        <button type="button" className="btn btn-link btn-sm ml-2 py-0 px-1" onClick={onEditClick}>
                            Edit
                        </button>
                    </div>
                </>
            ) : (
                <EditCommentForm {...props} comment={comment} onCommentUpdate={onCommentUpdate} />
            )}
        </div>
    ) : null
}
