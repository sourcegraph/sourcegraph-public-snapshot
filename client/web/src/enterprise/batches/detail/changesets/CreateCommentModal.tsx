import Dialog from '@reach/dialog'
import classNames from 'classnames'
import React, { useCallback, useState } from 'react'

import { Form } from '@sourcegraph/branded/src/components/Form'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { asError, isErrorLike } from '@sourcegraph/shared/src/util/errors'

import { ErrorAlert } from '../../../../components/alerts'
import { ExternalServiceKind, Scalars } from '../../../../graphql-operations'

export interface CreateCommentModalProps {
    onCancel: () => void
    afterCreate: () => void
    userID: Scalars['ID'] | null
    // externalServiceKind: ExternalServiceKind
    // externalServiceURL: string
    // requiresSSH: boolean

    /** For testing only. */
    // createBatchChangesCredential?: typeof _createBatchChangesCredential
}

export const CreateCommentModal: React.FunctionComponent<CreateCommentModalProps> = ({
    onCancel,
    afterCreate,
    userID,
    // externalServiceKind,
    // externalServiceURL,
    // requiresSSH,
    // createBatchChangesCredential = _createBatchChangesCredential,
}) => {
    const labelId = 'addCredential'
    const [isLoading, setIsLoading] = useState<boolean | Error>(false)
    const [credential, setCredential] = useState<string>('')

    const onChangeCredential = useCallback<React.ChangeEventHandler<HTMLTextAreaElement>>(event => {
        setCredential(event.target.value)
    }, [])

    const onSubmit = useCallback<React.FormEventHandler>(
        async event => {
            event.preventDefault()
            setIsLoading(true)
            try {
                // const createdCredential = await createBatchChangesCredential({
                //     user: userID,
                //     credential,
                //     externalServiceKind,
                //     externalServiceURL,
                // })
                afterCreate()
            } catch (error) {
                setIsLoading(asError(error))
            }
        },
        [
            afterCreate,
            userID,
            credential,
            // externalServiceKind,
            // externalServiceURL,
            // requiresSSH,
            // createBatchChangesCredential,
        ]
    )

    return (
        <Dialog
            className="modal-body modal-body--top-third p-4 rounded border"
            onDismiss={onCancel}
            aria-labelledby={labelId}
        >
            <div className="web-content test-add-credential-modal">
                <h3>Post a bulk comment on changesets</h3>
                <p className="mb-4">Use this feature to create a bulk comment on all the selected code hosts.</p>
                {isErrorLike(isLoading) && <ErrorAlert error={isLoading} />}
                <Form onSubmit={onSubmit}>
                    <div className="form-group">
                        <label htmlFor="token">Comment text</label>
                        <textarea
                            id="token"
                            name="token"
                            className="form-control test-add-credential-modal-input text-monospace"
                            placeholder={`## Please review this

This change is really important for us so please go review and merge this.`}
                            required={true}
                            rows={8}
                            minLength={1}
                            value={credential}
                            onChange={onChangeCredential}
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
                            disabled={isLoading === true || credential.length === 0}
                            className="btn btn-primary test-add-credential-modal-submit"
                        >
                            {isLoading === true && <LoadingSpinner className="icon-inline" />}
                            Post comments
                        </button>
                    </div>
                </Form>
            </div>
        </Dialog>
    )
}
