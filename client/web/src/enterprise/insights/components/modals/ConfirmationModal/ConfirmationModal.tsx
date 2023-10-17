import React from 'react'

import { Button, type ButtonProps, Modal } from '@sourcegraph/wildcard'

export interface ConfirmationModalProps {
    showModal: boolean
    onCancel: () => void
    onConfirm: () => void
    ariaLabel: string
    disabled?: boolean
    variant?: ButtonProps['variant']
    cancelText?: string
    confirmText?: string
}

export const ConfirmationModal: React.FunctionComponent<
    React.PropsWithChildren<React.PropsWithChildren<ConfirmationModalProps>>
> = props => {
    const {
        showModal,
        onCancel,
        onConfirm,
        children,
        ariaLabel,
        disabled,
        variant = 'primary',
        cancelText = 'Cancel',
        confirmText = 'Confirm',
    } = props

    return (
        <Modal isOpen={showModal} position="center" aria-label={ariaLabel}>
            {children}
            <div className="d-flex justify-content-end">
                <Button className="mr-2" onClick={onCancel} outline={true} variant="secondary" disabled={disabled}>
                    {cancelText}
                </Button>
                <Button onClick={onConfirm} variant={variant} disabled={disabled}>
                    {confirmText}
                </Button>
            </div>
        </Modal>
    )
}
