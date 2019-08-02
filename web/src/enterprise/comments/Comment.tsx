import { NotificationType } from '@sourcegraph/extension-api-classes'
import H from 'history'
import PencilIcon from 'mdi-react/PencilIcon'
import React, { useCallback, useState } from 'react'
import { map, mapTo } from 'rxjs/operators'
import { Markdown } from '../../../../shared/src/components/Markdown'
import { ExtensionsControllerProps } from '../../../../shared/src/extensions/controller'
import { dataOrThrowErrors, gql } from '../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../shared/src/graphql/schema'
import { mutateGraphQL } from '../../backend/graphql'
import { Timestamp } from '../../components/time/Timestamp'
import { PersonLink } from '../../user/PersonLink'
import { CommentForm } from './CommentForm'

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

interface Props extends ExtensionsControllerProps {
    comment: GQL.Comment
    onCommentUpdate: (update: Partial<GQL.Comment>) => void
    createdVerb?: string

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

    return (
        <div className={`comment card ${className}`}>
            <div className="card-header d-flex align-items-center justify-content-between">
                <span className="py-1">
                    <strong>
                        <PersonLink user={comment.author as GQL.IUser /* TODO!(sqs) */} />
                    </strong>{' '}
                    {createdVerb} <Timestamp date={comment.createdAt} />{' '}
                    {comment.updatedAt !== comment.createdAt && <> &bull; edited</>}
                </span>
                {!isEditing && comment.viewerCanUpdate && (
                    <button className="btn btn-sm btn-link text-muted p-1" onClick={onEdit}>
                        <PencilIcon className="icon-inline" />
                    </button>
                )}
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
                <div className="card-body">
                    <CommentForm
                        {...props}
                        initialBody={comment.body}
                        submitLabel="Save"
                        placeholder=""
                        onSubmit={onSubmit}
                        onCancel={onCancel}
                        disabled={isEditLoading}
                    />
                </div>
            )}
        </div>
    )
}
