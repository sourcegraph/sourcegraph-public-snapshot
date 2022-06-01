import React from 'react'

import { Button, Modal, H3 } from '@sourcegraph/wildcard'

import { LoaderButton } from '../../../../components/LoaderButton'

interface CancelExecutionModalProps {
    isOpen: boolean
    modalHeader?: string
    modalBody: React.ReactNode
    isLoading?: boolean
    confirmLabel?: string
    onCancel: () => void
    onConfirm: () => void
}

export const CancelExecutionModal: React.FunctionComponent<React.PropsWithChildren<CancelExecutionModalProps>> = ({
    isOpen,
    modalHeader = 'Cancel execution',
    modalBody,
    isLoading,
    confirmLabel = 'Cancel execution',
    onCancel,
    onConfirm,
}) => (
    <Modal isOpen={isOpen} position="center" aria-labelledby="modal-header">
        <H3 id="modal-header">{modalHeader}</H3>
        {modalBody}
        <div className="d-flex justify-content-end">
            <Button className="mr-2" onClick={onCancel} outline={true} variant="secondary">
                Go back
            </Button>
            <LoaderButton
                onClick={onConfirm}
                disabled={isLoading}
                variant="danger"
                loading={isLoading}
                alwaysShowLabel={true}
                label={confirmLabel}
            />
        </div>
    </Modal>
)
