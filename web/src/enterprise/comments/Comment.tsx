import { NotificationType } from '@sourcegraph/extension-api-classes'
import H from 'history'
import CloseIcon from 'mdi-react/CloseIcon'
import PencilIcon from 'mdi-react/PencilIcon'
import React, { useCallback, useState } from 'react'
import { map, mapTo } from 'rxjs/operators'
import { Markdown } from '../../../../shared/src/components/Markdown'
import { ExtensionsControllerProps } from '../../../../shared/src/extensions/controller'
import { dataOrThrowErrors, gql } from '../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../shared/src/graphql/schema'
import { ActorLink } from '../../actor/ActorLink'
import { mutateGraphQL } from '../../backend/graphql'
import { Timestamp } from '../../components/time/Timestamp'
import { CommentForm } from './CommentForm'
import { WithLinkPreviews } from '../../../../shared/src/components/linkPreviews/WithLinkPreviews'
import { setElementTooltip } from '../../components/tooltip/Tooltip'
import { LINK_PREVIEW_CLASS } from '../../components/linkPreviews/styles'

const editComment = (
    args: GQL.IEditCommentOnMutationArguments
): Promise<Pick<GQL.Comment, 'body' | 'bodyHTML' | 'updatedAt'>> =>
    mutateGraphQL(
        gql`
            mutation EditComment($input: EditCommentInput!) {
                editComment(input: $input) {
                    body
                    bodyHTML
                    updatedAt
                }
            }
        `,
        args
    )
        .pipe(
            map(dataOrThrowErrors),
            map(data => data.editComment)
        )
        .toPromise()

const deleteComment = (args: GQL.IDeleteCommentOnMutationArguments): Promise<void> =>
    mutateGraphQL(
        gql`
            mutation DeleteComment($comment: ID!) {
                deleteComment(comment: $comment) {
                    alwaysNil
                }
            }
        `,
        args
    )
        .pipe(
            map(dataOrThrowErrors),
            mapTo(undefined)
        )
        .toPromise()

interface Props extends ExtensionsControllerProps {
    comment: Pick<
        GQL.Comment,
        '__typename' | 'id' | 'author' | 'body' | 'bodyHTML' | 'createdAt' | 'updatedAt' | 'viewerCanUpdate'
    >
    onCommentUpdate: ((update?: Partial<Pick<GQL.IComment, Exclude<keyof GQL.IComment, '__typename'>>>) => void) | null
    createdVerb?: string
    emptyBody?: string

    tag?: keyof JSX.IntrinsicElements // TODO!(sqs): might cause infinite loop per https://github.com/Microsoft/TypeScript/issues/28892#issuecomment-477927604
    className?: string
    history: H.History
}

/**
 * A comment with a rendered body and an edit mode.
 */
export const Comment: React.FunctionComponent<Props> = ({
    comment,
    onCommentUpdate,
    createdVerb = 'commented',
    emptyBody,
    tag: Tag = 'div',
    className = '',
    ...props
}) => {
    const [isEditing, setIsEditing] = useState(false)
    const onEdit = useCallback(() => setIsEditing(true), [])
    const onCancel = useCallback(() => setIsEditing(false), [])

    const [isEditLoading, setIsEditLoading] = useState(false)
    const onSubmit = useCallback(
        (body: string) => {
            // tslint:disable-next-line: no-floating-promises
            ;(async () => {
                setIsEditLoading(true)
                try {
                    if (!onCommentUpdate) {
                        throw new Error('onCommentUpdate callback not set')
                    }
                    onCommentUpdate(await editComment({ input: { id: comment.id, body } }))
                    setIsEditLoading(false)
                    setIsEditing(false)
                } catch (err) {
                    setIsEditLoading(false)
                    props.extensionsController.services.notifications.showMessages.next({
                        message: `Error editing comment: ${err.message}`,
                        type: NotificationType.Error,
                    })
                }
            })()
        },
        [comment.id, onCommentUpdate, props.extensionsController.services.notifications.showMessages]
    )

    const [isDeleteLoading, setIsDeleteLoading] = useState(false)
    const onDelete = useCallback(() => {
        if (!confirm('Really delete this comment?')) {
            return
        }
        // tslint:disable-next-line: no-floating-promises
        ;(async () => {
            setIsDeleteLoading(true)
            try {
                if (!onCommentUpdate) {
                    throw new Error('onCommentUpdate callback not set')
                }
                await deleteComment({ comment: comment.id })
                setIsDeleteLoading(false)
                setIsEditing(false)
                onCommentUpdate()
            } catch (err) {
                setIsDeleteLoading(false)
                props.extensionsController.services.notifications.showMessages.next({
                    message: `Error deleting comment: ${err.message}`,
                    type: NotificationType.Error,
                })
            }
        })()
    }, [comment.id, onCommentUpdate, props.extensionsController.services.notifications.showMessages])

    return (
        <Tag className={`comment card ${className}`}>
            <div className="card-header d-flex align-items-center justify-content-between">
                <span className="py-1">
                    <strong>
                        <ActorLink actor={comment.author} />
                    </strong>{' '}
                    {createdVerb} <Timestamp date={comment.createdAt} />{' '}
                    {comment.updatedAt !== comment.createdAt && <> &bull; edited</>}
                </span>
                {onCommentUpdate && comment.viewerCanUpdate && (
                    <span>
                        {!isEditing && (
                            <button
                                className="btn btn-sm btn-link text-muted p-1"
                                onClick={onEdit}
                                aria-label="Edit comment"
                                disabled={isDeleteLoading}
                            >
                                <PencilIcon className="icon-inline" />
                            </button>
                        )}
                        {comment.__typename === 'CommentReply' && (
                            <button
                                className="btn btn-sm btn-link text-muted p-1"
                                onClick={onDelete}
                                aria-label="Delete comment"
                                disabled={isDeleteLoading}
                                // TODO!(sqs): dont show on primary comment for campaigns/threads, add
                                // new viewerCanDeleteComment to graphql or something?
                            >
                                <CloseIcon className="icon-inline" />
                            </button>
                        )}
                    </span>
                )}
            </div>
            {!isEditing ? (
                <div className="card-body">
                    {comment.bodyHTML ? (
                        <WithLinkPreviews
                            dangerousInnerHTML={comment.bodyHTML}
                            extensionsController={props.extensionsController}
                            setElementTooltip={setElementTooltip}
                            linkPreviewContentClass={LINK_PREVIEW_CLASS}
                        >
                            {props => <Markdown {...props} />}
                        </WithLinkPreviews>
                    ) : (
                        <span className="text-muted font-style-italic">{emptyBody}</span>
                    )}
                </div>
            ) : (
                <div className="card-body">
                    <CommentForm
                        {...props}
                        initialBody={comment.body}
                        submitLabel="Save"
                        placeholder=""
                        onSubmit={onSubmit}
                        onCancel={onCancel}
                        autoFocus={true}
                        disabled={isEditLoading}
                    />
                </div>
            )}
        </Tag>
    )
}
