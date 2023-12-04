import React from 'react'

import { H3, Text } from '@sourcegraph/wildcard'

import type { Insight, InsightDashboard } from '../../../../core'
import { useRemoveInsightFromDashboard } from '../../../../hooks/use-remove-insight'
import { ConfirmationModal, type ConfirmationModalProps } from '../../../modals/ConfirmationModal'

interface ConfirmRemoveModalProps extends Pick<ConfirmationModalProps, 'showModal' | 'onCancel'> {
    insight: Insight
    dashboard: InsightDashboard | null
}

export const ConfirmRemoveModal: React.FunctionComponent<React.PropsWithChildren<ConfirmRemoveModalProps>> = ({
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
            <H3 className="text-danger mb-4">Remove Insight?</H3>
            <Text className="mb-4">
                Are you sure you want to remove the insight <strong>{insight.title}</strong> from the dashboard{' '}
                <strong>{dashboard?.title}</strong>?
            </Text>
        </ConfirmationModal>
    )
}
