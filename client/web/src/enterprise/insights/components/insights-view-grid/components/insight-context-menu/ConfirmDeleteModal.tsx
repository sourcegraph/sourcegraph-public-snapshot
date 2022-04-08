import React from 'react'

import { ConfirmationModal, ConfirmationModalProps } from '@sourcegraph/shared/src/components/ConfirmationModal'

import { Insight } from '../../../../core/types'
import { useDeleteInsight } from '../../../../hooks/use-delete-insight'

interface ConfirmDeleteModalProps extends Pick<ConfirmationModalProps, 'showModal' | 'handleCancel'> {
    insight: Insight
}

export const ConfirmDeleteModal: React.FunctionComponent<ConfirmDeleteModalProps> = ({
    insight,
    showModal,
    handleCancel,
}) => {
    const { delete: handleDelete } = useDeleteInsight()

    return (
        <ConfirmationModal
            showModal={showModal}
            handleCancel={handleCancel}
            handleConfirmation={() => handleDelete(insight)}
            header="Delete Insight?"
            message={
                <>
                    Are you sure you want to delete the insight <strong>{insight.title}</strong>?
                </>
            }
        />
    )
}
