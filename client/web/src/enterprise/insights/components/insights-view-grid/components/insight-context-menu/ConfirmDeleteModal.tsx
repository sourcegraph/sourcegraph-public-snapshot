import React from 'react'

import { ConfirmationModal, ConfirmationModalProps } from '@sourcegraph/wildcard'

import { Insight } from '../../../../core/types'
import { useDeleteInsight } from '../../../../hooks/use-delete-insight'

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
        >
            <h3>Delete Insight?</h3>
            <p className="mb-4">
                Are you sure you want to delete the insight <strong>{insight.title}</strong>?
            </p>
        </ConfirmationModal>
    )
}
