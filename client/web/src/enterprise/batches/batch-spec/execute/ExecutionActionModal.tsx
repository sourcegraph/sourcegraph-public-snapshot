import React from 'react'

import { Button, Modal, H3 } from '@sourcegraph/wildcard'
import { BUTTON_VARIANTS } from '@sourcegraph/wildcard/src/components/Button'

import { LoaderButton } from '../../../../components/LoaderButton'

interface ExecutionActionModalProps {
    isOpen: boolean
    modalHeader: string
    modalBody: React.ReactNode
    isLoading?: boolean
    confirmLabel: string
    confirmVariant?: typeof BUTTON_VARIANTS[number]
    onCancel: () => void
    onConfirm: () => void
}

export const ExecutionActionModal: React.FunctionComponent<React.PropsWithChildren<ExecutionActionModalProps>> = ({
    isOpen,
    modalHeader,
    modalBody,
    isLoading,
    confirmLabel,
    confirmVariant = 'danger',
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
                variant={confirmVariant}
                loading={isLoading}
                alwaysShowLabel={true}
                label={confirmLabel}
            />
        </div>
    </Modal>
)
