import { FC, useContext, useMemo } from 'react'

import { useQuery } from '@apollo/client'
import { mdiClose } from '@mdi/js'
import { VisuallyHidden } from '@reach/visually-hidden'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { isDefined } from '@sourcegraph/common'
import { Button, Modal, H2, Icon, LoadingSpinner } from '@sourcegraph/wildcard'

import {
    AccessibleInsight,
    GetDashboardAccessibleInsightsResult,
    GetDashboardAccessibleInsightsVariables,
} from '../../../../../../../graphql-operations'
import { SubmissionErrors } from '../../../../../components'
import { CodeInsightsBackendContext, CustomInsightDashboard } from '../../../../../core'

import {
    AddInsightFormValues,
    AddInsightModalContent,
} from './components/add-insight-modal-content/AddInsightModalContent'
import { GET_ACCESSIBLE_INSIGHTS_LIST } from './query'

import styles from './AddInsightModal.module.scss'

export interface AddInsightModalProps {
    dashboard: CustomInsightDashboard
    onClose: () => void
}

export const AddInsightModal: FC<AddInsightModalProps> = props => {
    const { dashboard, onClose } = props
    const { assignInsightsToDashboard } = useContext(CodeInsightsBackendContext)

    const { data, loading, error } = useQuery<
        GetDashboardAccessibleInsightsResult,
        GetDashboardAccessibleInsightsVariables
    >(GET_ACCESSIBLE_INSIGHTS_LIST, {
        variables: { id: dashboard.id },
        errorPolicy: 'all',
    })

    const insights = getAvailableInsights(data)
    const dashboardInsightIds = getDashboardInsightIds(data)

    const initialValues = useMemo<AddInsightFormValues>(
        () => ({
            searchInput: '',
            insightIds: dashboardInsightIds,
        }),
        [dashboardInsightIds]
    )

    const handleSubmit = async (values: AddInsightFormValues): Promise<void | SubmissionErrors> => {
        const { insightIds } = values

        await assignInsightsToDashboard({
            id: dashboard.id,
            prevInsightIds: dashboardInsightIds,
            nextInsightIds: insightIds,
        }).toPromise()

        onClose()
    }

    return (
        <Modal className={styles.modal} onDismiss={onClose} aria-label="Add insights to dashboard modal">
            <Button variant="icon" className={styles.closeButton} onClick={onClose}>
                <VisuallyHidden>Close</VisuallyHidden>
                <Icon svgPath={mdiClose} inline={false} aria-hidden={true} />
            </Button>

            {loading && !data && <LoadingSpinner inline={false} />}
            {error && <ErrorAlert error={error} />}
            {data && (
                <>
                    <H2 className="mb-3">
                        Add insight to <q>{dashboard.title}</q>
                    </H2>

                    {!insights.length && <span>There are no insights for this dashboard.</span>}

                    {insights.length > 0 && (
                        <AddInsightModalContent
                            initialValues={initialValues}
                            insights={insights}
                            dashboardID={dashboard.id}
                            onCancel={onClose}
                            onSubmit={handleSubmit}
                        />
                    )}
                </>
            )}
        </Modal>
    )
}

function getDashboardInsightIds(data?: GetDashboardAccessibleInsightsResult): string[] {
    if (!data || !data.dashboardInsightsIds.nodes[0]?.views) {
        return []
    }

    return data.dashboardInsightsIds.nodes[0].views.nodes.filter(isDefined).map(view => view.id)
}

function getAvailableInsights(data?: GetDashboardAccessibleInsightsResult): AccessibleInsight[] {
    return data?.accessibleInsights?.nodes.filter(isDefined) ?? []
}
