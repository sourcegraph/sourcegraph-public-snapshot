import React, { useCallback, useState } from 'react'
import * as H from 'history'
import Dialog from '@reach/dialog'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { asError, isErrorLike } from '../../../../../shared/src/util/errors'
import { ErrorAlert } from '../../../components/alerts'
import { deleteCampaignsCredential } from './backend'
import { Scalars } from '../../../graphql-operations'

export interface RemoveCredentialModalProps {
    credentialID: Scalars['ID']

    onCancel: () => void
    afterDelete: () => void

    history: H.History
}

export const RemoveCredentialModal: React.FunctionComponent<RemoveCredentialModalProps> = ({
    credentialID,
    onCancel,
    afterDelete,
    history,
}) => {
    const labelId = 'removeCredential'
    const [isLoading, setIsLoading] = useState<boolean | Error>(false)
    const onDelete = useCallback<React.MouseEventHandler>(async () => {
        setIsLoading(true)
        try {
            await deleteCampaignsCredential(credentialID)
            afterDelete()
        } catch (error) {
            setIsLoading(asError(error))
        }
    }, [afterDelete, credentialID])
    return (
        <Dialog
            className="modal-body modal-body--top-third p-4 rounded border"
            onDismiss={onCancel}
            aria-labelledby={labelId}
        >
            <div className="web-content test-remove-credential-modal">
                <h3 className="text-danger" id={labelId}>
                    Remove campaigns token?
                </h3>

                {isErrorLike(isLoading) && <ErrorAlert error={isLoading} history={history} />}

                <p>You will not be able to create changesets on this code host if this token is removed.</p>

                <div className="d-flex justify-content-end pt-5">
                    <button
                        type="button"
                        disabled={isLoading === true}
                        className="btn btn-outline-secondary mr-2"
                        onClick={onCancel}
                    >
                        Cancel
                    </button>
                    <button
                        type="button"
                        disabled={isLoading === true}
                        className="btn btn-danger test-remove-credential-modal-submit"
                        onClick={onDelete}
                    >
                        {isLoading === true && <LoadingSpinner className="icon-inline" />}
                        Yes, remove connection
                    </button>
                </div>
            </div>
        </Dialog>
    )
}
