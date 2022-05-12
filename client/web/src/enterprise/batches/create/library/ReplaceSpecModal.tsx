import React from 'react'

import { Button, Modal, Typography } from '@sourcegraph/wildcard'

export interface ReplaceSpecModalProps {
    libraryItemName: string
    onCancel: () => void
    onConfirm: () => void
}

export const ReplaceSpecModal: React.FunctionComponent<React.PropsWithChildren<ReplaceSpecModalProps>> = ({
    libraryItemName,
    onCancel,
    onConfirm,
}) => (
    <Modal onDismiss={onCancel} aria-labelledby={MODAL_LABEL_ID}>
        <Typography.H3 id={MODAL_LABEL_ID}>Replace batch spec?</Typography.H3>
        <p className="mb-4">
            Are you sure you want to replace your current batch spec with the template for{' '}
            <strong>{libraryItemName}</strong>?
        </p>
        <div className="d-flex justify-content-end">
            <Button className="mr-2" onClick={onCancel} outline={true} variant="secondary">
                Cancel
            </Button>
            <Button onClick={onConfirm} variant="primary">
                Confirm
            </Button>
        </div>
    </Modal>
)

const MODAL_LABEL_ID = 'replace-batch-spec-modal-title'
