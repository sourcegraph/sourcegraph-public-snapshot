import Dialog from '@reach/dialog'
import React, { useCallback, useState } from 'react'

import { asError, isErrorLike } from '@sourcegraph/common'
import { Button, LoadingSpinner } from '@sourcegraph/wildcard'

import { ErrorAlert } from '../../../components/alerts'
import { BatchChangesCodeHostFields, BatchChangesCredentialFields } from '../../../graphql-operations'

import { deleteBatchChangesCredential } from './backend'
import { CodeHostSshPublicKey } from './CodeHostSshPublicKey'
import { ModalHeader } from './ModalHeader'

export interface RemoveCredentialModalProps {
    codeHost: BatchChangesCodeHostFields
    credential: BatchChangesCredentialFields

    onCancel: () => void
    afterDelete: () => void
}

export const RemoveCredentialModal: React.FunctionComponent<RemoveCredentialModalProps> = ({
    codeHost,
    credential,
    onCancel,
    afterDelete,
}) => {
    const labelId = 'removeCredential'
    const [isLoading, setIsLoading] = useState<boolean | Error>(false)
    const onDelete = useCallback<React.MouseEventHandler>(async () => {
        setIsLoading(true)
        try {
            await deleteBatchChangesCredential(credential.id)
            afterDelete()
        } catch (error) {
            setIsLoading(asError(error))
        }
    }, [afterDelete, credential.id])
    return (
        <Dialog
            className="modal-body modal-body--top-third p-4 rounded border"
            onDismiss={onCancel}
            aria-labelledby={labelId}
        >
            <div className="test-remove-credential-modal">
                <ModalHeader
                    id={labelId}
                    externalServiceKind={codeHost.externalServiceKind}
                    externalServiceURL={codeHost.externalServiceURL}
                />

                <h3 className="text-danger mb-4">Removing credentials is irreversible</h3>

                {isErrorLike(isLoading) && <ErrorAlert error={isLoading} />}

                <p>
                    To create changesets on this code host after removing credentials, you will need to repeat the 'Add
                    credentials' process.
                </p>

                {codeHost.requiresSSH && (
                    <CodeHostSshPublicKey
                        externalServiceKind={codeHost.externalServiceKind}
                        sshPublicKey={credential.sshPublicKey!}
                        showInstructionsLink={false}
                        showCopyButton={false}
                        label="Public key to remove"
                    />
                )}

                <div className="d-flex justify-content-end pt-1">
                    <Button
                        disabled={isLoading === true}
                        className="mr-2"
                        onClick={onCancel}
                        outline={true}
                        variant="secondary"
                    >
                        Cancel
                    </Button>
                    <Button
                        disabled={isLoading === true}
                        className="test-remove-credential-modal-submit"
                        onClick={onDelete}
                        variant="danger"
                    >
                        {isLoading === true && <LoadingSpinner />}
                        Remove credentials
                    </Button>
                </div>
            </div>
        </Dialog>
    )
}
