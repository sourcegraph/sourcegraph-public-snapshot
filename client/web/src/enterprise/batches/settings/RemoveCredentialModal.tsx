import React, { useCallback } from 'react'

import { Button, Modal, Text, ErrorAlert } from '@sourcegraph/wildcard'

import { LoaderButton } from '../../../components/LoaderButton'
import type { BatchChangesCodeHostFields, BatchChangesCredentialFields } from '../../../graphql-operations'

import { useDeleteBatchChangesCredential } from './backend'
import { CodeHostSshPublicKey } from './CodeHostSshPublicKey'
import { ModalHeader } from './ModalHeader'

export interface RemoveCredentialModalProps {
    codeHost: BatchChangesCodeHostFields
    credential: BatchChangesCredentialFields

    onCancel: () => void
    afterDelete: () => void
}

export const RemoveCredentialModal: React.FunctionComponent<React.PropsWithChildren<RemoveCredentialModalProps>> = ({
    codeHost,
    credential,
    onCancel,
    afterDelete,
}) => {
    const labelId = 'removeCredential'
    const [deleteBatchChangesCredential, { loading, error }] = useDeleteBatchChangesCredential()
    const onDelete = useCallback<React.MouseEventHandler>(async () => {
        await deleteBatchChangesCredential({ variables: { id: credential.id } })
        afterDelete()
    }, [afterDelete, credential.id, deleteBatchChangesCredential])
    return (
        <Modal onDismiss={onCancel} aria-labelledby={labelId}>
            <div className="test-remove-credential-modal">
                <ModalHeader
                    id={labelId}
                    externalServiceKind={codeHost.externalServiceKind}
                    externalServiceURL={codeHost.externalServiceURL}
                />

                <strong className="d-block text-danger my-3">Removing credentials is irreversible.</strong>

                {error && <ErrorAlert error={error} />}

                <Text>
                    To create changesets on this code host after removing credentials, you will need to repeat the 'Add
                    credentials' process.
                </Text>

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
                    <Button disabled={loading} className="mr-2" onClick={onCancel} outline={true} variant="secondary">
                        Cancel
                    </Button>
                    <LoaderButton
                        disabled={loading}
                        className="test-remove-credential-modal-submit"
                        onClick={onDelete}
                        variant="danger"
                        loading={loading}
                        alwaysShowLabel={true}
                        label="Remove credentials"
                    />
                </div>
            </div>
        </Modal>
    )
}
