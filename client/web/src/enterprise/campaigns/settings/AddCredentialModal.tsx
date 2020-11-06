import React, { useCallback, useState } from 'react'
import * as H from 'history'
import Dialog from '@reach/dialog'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import LockIcon from 'mdi-react/LockIcon'
import { Form } from '../../../../../branded/src/components/Form'
import { asError, isErrorLike } from '../../../../../shared/src/util/errors'
import { ErrorAlert } from '../../../components/alerts'
import { createCampaignsCredential } from './backend'
import { ExternalServiceKind } from '../../../graphql-operations'

export interface AddCredentialModalProps {
    onCancel: () => void
    afterCreate: () => void
    history: H.History
    externalServiceKind: ExternalServiceKind
    externalServiceURL: string
}

export const AddCredentialModal: React.FunctionComponent<AddCredentialModalProps> = ({
    onCancel,
    afterCreate,
    history,
    externalServiceKind,
    externalServiceURL,
}) => {
    const labelId = 'addCredential'
    const [isLoading, setIsLoading] = useState<boolean | Error>(false)
    const [credential, setCredential] = useState<string>('')
    const onChangeCredential = useCallback<React.ChangeEventHandler<HTMLInputElement>>(event => {
        setCredential(event.target.value)
    }, [])
    const onSubmit = useCallback<React.FormEventHandler>(
        async event => {
            event.preventDefault()
            setIsLoading(true)
            try {
                await createCampaignsCredential({ credential, externalServiceKind, externalServiceURL })
                afterCreate()
            } catch (error) {
                setIsLoading(asError(error))
            }
        },
        [afterCreate, credential, externalServiceKind, externalServiceURL]
    )
    return (
        <Dialog
            className="modal-body modal-body--top-third p-4 rounded border"
            onDismiss={onCancel}
            aria-labelledby={labelId}
        >
            <div className="web-content">
                <h3 id={labelId}>GitHub campaigns token</h3>
                {isErrorLike(isLoading) && <ErrorAlert error={isLoading} history={history} />}
                <Form onSubmit={onSubmit}>
                    <div className="form-group">
                        <label htmlFor="token">Personal access token</label>
                        <input
                            id="token"
                            name="token"
                            type="text"
                            className="form-control"
                            required={true}
                            minLength={1}
                            value={credential}
                            onChange={onChangeCredential}
                        />
                        <p className="form-text text-muted">
                            <a href="asdasdas" rel="noreferrer noopener" target="_blank">
                                Create a new access token
                            </a>{' '}
                            with repo or public_repo scope.
                        </p>
                        <p className="form-text">
                            <LockIcon className="icon-inline" /> Access tokens are encrypted before storing.
                        </p>
                    </div>
                    <div className="d-flex justify-content-end pt-5">
                        <button
                            type="button"
                            disabled={isLoading === true}
                            className="btn btn-outline-secondary mr-2"
                            onClick={onCancel}
                        >
                            Cancel
                        </button>
                        <button type="submit" disabled={isLoading === true} className="btn btn-primary">
                            {isLoading === true && <LoadingSpinner className="icon-inline" />}
                            Add token
                        </button>
                    </div>
                </Form>
            </div>
        </Dialog>
    )
}
