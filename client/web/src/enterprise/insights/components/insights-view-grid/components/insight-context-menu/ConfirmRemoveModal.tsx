import React from 'react'

import { Insight, InsightDashboard } from '../../../../core/types'
import { useRemoveInsightFromDashboard } from '../../../../hooks/use-remove-insight'
import { ConfirmationModal, ConfirmationModalProps } from '../ConfirmationModal'

interface ConfirmRemoveModalProps extends Pick<ConfirmationModalProps, 'showModal' | 'onCancel'> {
    insight: Insight
    dashboard: InsightDashboard | null
}

export const ConfirmRemoveModal: React.FunctionComponent<ConfirmRemoveModalProps> = ({
    insight,
    dashboard,
    showModal,
    onCancel,
}) => {
    const { remove, loading } = useRemoveInsightFromDashboard()

    return (
        <ConfirmationModal
            showModal={showModal}
            onCancel={onCancel}
            onConfirm={() => dashboard && !loading && remove({ insight, dashboard })}
            ariaLabel="Remove insight modal"
            disabled={loading}
            variant="danger"
        >
            <h3 className="text-danger mb-4">Remove Insight?</h3>
            <p className="mb-4">
                Are you sure you want to remove the insight <strong>{insight.title}</strong> from the dashboard{' '}
                <strong>{dashboard?.title}</strong>?
            </p>
        </ConfirmationModal>
    )
}
