import React from 'react'

import { Insight } from '../../../../core/types'
import { useDeleteInsight } from '../../../../hooks/use-delete-insight'
import { ConfirmationModal, ConfirmationModalProps } from '../ConfirmationModal'

interface ConfirmDeleteModalProps extends Pick<ConfirmationModalProps, 'showModal' | 'onCancel'> {
    insight: Insight
}

export const ConfirmDeleteModal: React.FunctionComponent<ConfirmDeleteModalProps> = ({
    insight,
    showModal,
    onCancel,
}) => {
    const { delete: handleDelete, loading } = useDeleteInsight()

    return (
        <ConfirmationModal
            showModal={showModal}
            onCancel={onCancel}
            onConfirm={() => !loading && handleDelete(insight)}
            ariaLabel="Delete insight modal"
            disabled={loading}
            variant="danger"
            confirmText="Delete forever"
        >
            <h3 className="text-danger mb-4">Delete '{insight.title}'?</h3>
            <p className="mb-4">Are you sure you want to delete insight {insight.title}? This can't be undone.</p>
        </ConfirmationModal>
    )
}
