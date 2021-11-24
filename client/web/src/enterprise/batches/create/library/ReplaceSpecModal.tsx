import Dialog from '@reach/dialog'
import React from 'react'

export interface ReplaceSpecModalProps {
    libraryItemName: string
    onCancel: () => void
    onConfirm: () => void
}

export const ReplaceSpecModal: React.FunctionComponent<ReplaceSpecModalProps> = ({
    libraryItemName,
    onCancel,
    onConfirm,
}) => (
    <Dialog
        className="modal-body modal-body--top-third p-4 rounded border"
        onDismiss={onCancel}
        aria-labelledby={MODAL_LABEL_ID}
    >
        <h3 id={MODAL_LABEL_ID}>Replace batch spec?</h3>
        <p className="mb-4">
            Are you sure you want to replace your current batch spec with the template for{' '}
            <strong>{libraryItemName}</strong>?
        </p>
        <div className="d-flex justify-content-end">
            <button type="button" className="btn btn-outline-secondary mr-2" onClick={onCancel}>
                Cancel
            </button>
            <button type="button" onClick={onConfirm} className="btn btn-primary">
                Confirm
            </button>
        </div>
    </Dialog>
)

const MODAL_LABEL_ID = 'replace-batch-spec-modal-title'
