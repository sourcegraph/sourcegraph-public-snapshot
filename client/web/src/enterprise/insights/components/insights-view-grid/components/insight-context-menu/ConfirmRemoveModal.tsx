import React from 'react'

import { ConfirmationModal, ConfirmationModalProps } from '@sourcegraph/shared/src/components/ConfirmationModal'

import { Insight, InsightDashboard } from '../../../../core/types'
import { useRemoveInsightFromDashboard } from '../../../../hooks/use-remove-insight'

interface ConfirmRemoveModalProps extends Pick<ConfirmationModalProps, 'showModal' | 'handleCancel'> {
    insight: Insight
    dashboard: InsightDashboard | null
}

export const ConfirmRemoveModal: React.FunctionComponent<ConfirmRemoveModalProps> = ({
    insight,
    dashboard,
    showModal,
    handleCancel,
}) => {
    const { remove: handleRemove } = useRemoveInsightFromDashboard()

    return (
        <ConfirmationModal
            showModal={showModal}
            handleCancel={handleCancel}
            handleConfirmation={() => dashboard && handleRemove({ insight, dashboard })}
            header="Remove Insight?"
            message={
                <>
                    Are you sure you want to remove the insight <strong>{insight.title}</strong> from the dashboard{' '}
                    <strong>{dashboard?.title}</strong>?
                </>
            }
        />
    )
}
