import React from 'react'

import { Button, Modal } from '@sourcegraph/wildcard'

export interface ConfirmationModalProps {
    showModal: boolean
    handleCancel: () => void
    handleConfirmation: () => void
    header: React.ReactElement | string
    message: React.ReactElement | string
}

export const ConfirmationModal: React.FunctionComponent<ConfirmationModalProps> = (props: ConfirmationModalProps) => {
    const { showModal, handleCancel, handleConfirmation, header, message } = props

    return (
        <Modal isOpen={showModal} position="center" aria-label="Delete confirmation modal">
            <h3>{header}</h3>
            <p className="mb-4">{message}</p>
            <div className="d-flex justify-content-end">
                <Button className="mr-2" onClick={handleCancel} outline={true} variant="secondary">
                    Cancel
                </Button>
                <Button onClick={handleConfirmation} variant="primary">
                    Confirm
                </Button>
            </div>
        </Modal>
    )
}
