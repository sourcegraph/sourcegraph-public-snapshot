import React, { useCallback, useState } from 'react'

import { asError, isErrorLike } from '@sourcegraph/common'
import { Button, TextArea, Modal, H3, Text, ErrorAlert, Form } from '@sourcegraph/wildcard'

import { LoaderButton } from '../../../../components/LoaderButton'
import type { Scalars } from '../../../../graphql-operations'
import { createChangesetComments as _createChangesetComments } from '../backend'

export interface CreateCommentModalProps {
    onCancel: () => void
    afterCreate: () => void
    batchChangeID: Scalars['ID']
    changesetIDs: Scalars['ID'][]

    /** For testing only. */
    createChangesetComments?: typeof _createChangesetComments
}

export const CreateCommentModal: React.FunctionComponent<React.PropsWithChildren<CreateCommentModalProps>> = ({
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
                await createChangesetComments(batchChangeID, changesetIDs, commentBody)
                afterCreate()
            } catch (error) {
                setIsLoading(asError(error))
            }
        },
        [afterCreate, batchChangeID, changesetIDs, commentBody, createChangesetComments]
    )

    return (
        <Modal onDismiss={onCancel} aria-labelledby={LABEL_ID}>
            <H3 id={LABEL_ID}>Post a bulk comment on changesets</H3>
            <Text className="mb-4">Use this feature to create a bulk comment on all the selected code hosts.</Text>
            {isErrorLike(isLoading) && <ErrorAlert error={isLoading} />}
            <Form onSubmit={onSubmit}>
                <div className="form-group">
                    <TextArea
                        id="token"
                        name="token"
                        placeholder={PLACEHOLDER_COMMENT}
                        required={true}
                        rows={8}
                        minLength={1}
                        value={commentBody}
                        onChange={onChangeInput}
                        label="Comment text"
                    />
                </div>
                <div className="d-flex justify-content-end">
                    <Button
                        disabled={isLoading === true}
                        className="mr-2"
                        onClick={onCancel}
                        outline={true}
                        variant="secondary"
                    >
                        Cancel
                    </Button>
                    <LoaderButton
                        type="submit"
                        disabled={isLoading === true || commentBody.length === 0}
                        variant="primary"
                        loading={isLoading === true}
                        alwaysShowLabel={true}
                        label="Post comments"
                    />
                </div>
            </Form>
        </Modal>
    )
}

const LABEL_ID = 'create-comment-modal-id'

const PLACEHOLDER_COMMENT = `A comment that will be posted to all selected changesets.

You can use whatever formatting is available on your code host, such as _Markdown_!

Use this to request a review, provide an update, or post your favorite emoji, like 🦡.`
