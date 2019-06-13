import H from 'history'
import React, { useCallback } from 'react'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { updateComment } from '../../../discussions/backend'
import { DiscussionsInput, TitleMode } from './DiscussionsInput'

interface Props extends ExtensionsControllerProps {
    comment: Pick<GQL.IDiscussionComment, 'id' | 'contents'>
    onCommentUpdate: (thread: GQL.IDiscussionThread | null) => void

    className?: string
    location: H.Location
    history: H.History
}

/**
 * A form to edit a comment in a discussion thread.
 */
export const EditCommentForm: React.FunctionComponent<Props> = ({ comment, onCommentUpdate, ...props }) => {
    const onSubmit = useCallback(
        async (_title: string, contents: string): Promise<void> => {
            onCommentUpdate(await updateComment({ commentID: comment.id, contents }).toPromise())
        },
        [onCommentUpdate, comment.id]
    )
    const onDiscard = useCallback(() => onCommentUpdate(null), [onCommentUpdate])
    return (
        <DiscussionsInput
            {...props}
            initialContents={comment.contents}
            titleMode={TitleMode.None}
            submitLabel="Update"
            onSubmit={onSubmit}
            onDiscard={onDiscard}
        />
    )
}
