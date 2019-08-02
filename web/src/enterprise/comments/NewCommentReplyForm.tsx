import { NotificationType } from '@sourcegraph/extension-api-classes'
import H from 'history'
import React, { useCallback, useState } from 'react'
import { map, mapTo } from 'rxjs/operators'
import { ExtensionsControllerProps } from '../../../../shared/src/extensions/controller'
import { dataOrThrowErrors, gql } from '../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../shared/src/graphql/schema'
import { mutateGraphQL } from '../../backend/graphql'
import { CommentForm } from './CommentForm'

const createComment = (args: GQL.ICreateCommentOnMutationArguments): Promise<void> =>
    mutateGraphQL(
        gql`
            mutation CreateComment($input: CreateCommentInput!) {
                createComment(input: $input) {
                    id
                }
            }
        `,
        args
    )
        .pipe(
            map(dataOrThrowErrors),
            mapTo(void 0)
        )
        .toPromise()

interface Props extends ExtensionsControllerProps {
    commentable: Pick<GQL.Commentable, 'id'>
    onCommentableUpdate: () => void

    className?: string
    history: H.History
}

/**
 * A form to add a new reply comment to a commentable.
 */
export const NewCommentReplyForm: React.FunctionComponent<Props> = ({
    commentable,
    onCommentableUpdate,
    className = '',
    ...props
}) => {
    const [isLoading, setIsLoading] = useState(false)
    const onSubmit = useCallback(
        (body: string) => {
            // tslint:disable-next-line: no-floating-promises
            ;(async () => {
                setIsLoading(true)
                try {
                    await createComment({ input: { node: commentable.id, body } })
                    onCommentableUpdate()
                    setIsLoading(false)
                    setIsLoading(false)
                } catch (err) {
                    setIsLoading(false)
                    props.extensionsController.services.notifications.showMessages.next({
                        message: `Error adding comment: ${err.message}`,
                        type: NotificationType.Error,
                    })
                }
            })()
        },
        [commentable.id, onCommentableUpdate, props.extensionsController.services.notifications.showMessages]
    )

    return (
        <CommentForm
            {...props}
            submitLabel="Comment"
            placeholder="Leave a comment"
            onSubmit={onSubmit}
            disabled={isLoading}
            className={className}
        />
    )
}
