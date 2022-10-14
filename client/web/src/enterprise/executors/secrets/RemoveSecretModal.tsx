import React, { useCallback } from 'react'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { Button, Modal, Text } from '@sourcegraph/wildcard'

import { LoaderButton } from '../../../components/LoaderButton'
import { BatchChangesCodeHostFields, BatchChangesCredentialFields } from '../../../graphql-operations'

import { useDeleteExecutorSecret } from './backend'
import { ModalHeader } from './ModalHeader'

export interface RemoveSecretModalProps {
    secret: ExecutorSecretFields

    onCancel: () => void
    afterDelete: () => void
}

export const RemoveSecretModal: React.FunctionComponent<React.PropsWithChildren<RemoveSecretModalProps>> = ({
    secret,
    onCancel,
    afterDelete,
}) => {
    const labelId = 'removeCredential'
    const [deleteExecutorSecret, { loading, error }] = useDeleteExecutorSecret()
    const onDelete = useCallback<React.MouseEventHandler>(async () => {
        await deleteExecutorSecret({ variables: { id: secret.id, scope: secret.scope } })
        afterDelete()
    }, [afterDelete, secret.id, secret.scope, deleteExecutorSecret])
    return (
        <Modal onDismiss={onCancel} aria-labelledby={labelId}>
            <H3 id={labelId}>Remove secret {secret.key}?</H3>

            <strong className="d-block text-danger my-3">Removing secrets is irreversible.</strong>

            {error && <ErrorAlert error={error} />}

            <div className="d-flex justify-content-end pt-1">
                <Button disabled={loading} className="mr-2" onClick={onCancel} outline={true} variant="secondary">
                    Cancel
                </Button>
                <LoaderButton
                    disabled={loading}
                    onClick={onDelete}
                    variant="danger"
                    loading={loading}
                    alwaysShowLabel={true}
                    label="Remove secret"
                />
            </div>
        </Modal>
    )
}
