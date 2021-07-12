import Dialog from '@reach/dialog'
import React, { useCallback, useState } from 'react'

import { Form } from '@sourcegraph/branded/src/components/Form'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { asError, isErrorLike } from '@sourcegraph/shared/src/util/errors'

import { ErrorAlert } from '../../../../components/alerts'
import { Scalars } from '../../../../graphql-operations'
import { createChangesetComments as _createChangesetComments } from '../backend'

export interface CreateCommentModalProps {
    onCancel: () => void
    afterCreate: () => void
    batchChangeID: Scalars['ID']
    changesetIDs: () => Promise<Scalars['ID'][]>

    /** For testing only. */
    createChangesetComments?: typeof _createChangesetComments
}

export const CreateCommentModal: React.FunctionComponent<CreateCommentModalProps> = ({
    onCancel,
    afterCreate,
    batchChangeID,
    changesetIDs,
    createChangesetComments = _createChangesetComments,
}) => {
    const [isLoading, setIsLoading] = useState<boolean | Error>(false)
    const [commentBody, setCommentBody] = useState<string>('')

    const onChangeInput = useCallback<React.ChangeEventHandler<HTMLTextAreaElement>>(event => {
        setCommentBody(event.target.value)
    }, [])

    const onSubmit = useCallback<React.FormEventHandler>(
        async event => {
            event.preventDefault()
            setIsLoading(true)
            try {
                const ids = await changesetIDs()
                await createChangesetComments(batchChangeID, ids, commentBody)
                afterCreate()
            } catch (error) {
                setIsLoading(asError(error))
            }
        },
        [afterCreate, batchChangeID, changesetIDs, commentBody, createChangesetComments]
    )

    return (
        <Dialog
            className="modal-body modal-body--top-third p-4 rounded border"
            onDismiss={onCancel}
            aria-labelledby={LABEL_ID}
        >
            <h3 id={LABEL_ID}>Post a bulk comment on changesets</h3>
            <p className="mb-4">Use this feature to create a bulk comment on all the selected code hosts.</p>
            {isErrorLike(isLoading) && <ErrorAlert error={isLoading} />}
            <Form onSubmit={onSubmit}>
                <div className="form-group">
                    <label htmlFor="token">Comment text</label>
                    <textarea
                        id="token"
                        name="token"
                        className="form-control"
                        placeholder={PLACEHOLDER_COMMENT}
                        required={true}
                        rows={8}
                        minLength={1}
                        value={commentBody}
                        onChange={onChangeInput}
                    />
                </div>
                <div className="d-flex justify-content-end">
                    <button
                        type="button"
                        disabled={isLoading === true}
                        className="btn btn-outline-secondary mr-2"
                        onClick={onCancel}
                    >
                        Cancel
                    </button>
                    <button
                        type="submit"
                        disabled={isLoading === true || commentBody.length === 0}
                        className="btn btn-primary"
                    >
                        {isLoading === true && <LoadingSpinner className="icon-inline" />}
                        Post comments
                    </button>
                </div>
            </Form>
        </Dialog>
    )
}

const LABEL_ID = 'create-comment-modal-id'

const PLACEHOLDER_COMMENT = `A comment that will be posted to all selected changesets.

You can use whatever formatting is available on your code host, such as _Markdown_!

Use this to request a review, provide an update, or post your favorite emoji, like 🦡.`
