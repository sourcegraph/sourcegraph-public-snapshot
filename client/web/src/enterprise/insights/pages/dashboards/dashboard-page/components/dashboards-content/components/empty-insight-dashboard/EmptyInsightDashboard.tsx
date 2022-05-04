import React from 'react'

import PlusIcon from 'mdi-react/PlusIcon'

import { Button, Link, Card } from '@sourcegraph/wildcard'

import { ALL_INSIGHTS_DASHBOARD, InsightDashboard } from '../../../../../../../core'
import { useUiFeatures } from '../../../../../../../hooks/use-ui-features'
import { isDashboardConfigurable } from '../../utils/is-dashboard-configurable'

import styles from './EmptyInsightDashboard.module.scss'

interface EmptyInsightDashboardProps {
    dashboard: InsightDashboard
    onAddInsight: () => void
}

export const EmptyInsightDashboard: React.FunctionComponent<
    React.PropsWithChildren<EmptyInsightDashboardProps>
> = props => {
    const { onAddInsight, dashboard } = props

    return isDashboardConfigurable(dashboard) ? (
        <EmptySettingsBasedDashboard dashboard={dashboard} onAddInsight={onAddInsight} />
    ) : (
        <EmptyBuiltInDashboard dashboard={dashboard} />
    )
}

/**
 * Built-in empty dashboard state provides link to create a new code insight via creation UI.
 * Since all insights within built-in dashboards are calculated there's no ability to add insight to
 * this type of dashboard.
 */
export const EmptyBuiltInDashboard: React.FunctionComponent<
    React.PropsWithChildren<{ dashboard: InsightDashboard }>
> = props => (
    <section className={styles.emptySection}>
        <Card as={Link} to={`/insights/create?dashboardId=${props.dashboard.id}`} className={styles.itemCard}>
            <PlusIcon size="2rem" />
            <span>Create an insight</span>
        </Card>
        {props.dashboard.id !== ALL_INSIGHTS_DASHBOARD.id && (
            <span className="d-flex justify-content-center mt-3">
                <span>
                    or, add existing insights from <Link to="/insights/dashboards/all">All Insights</Link>
                </span>
            </span>
        )}
    </section>
)

/**
 * Settings based empty dashboard state provides button for adding existing insights to the dashboard.
 * Since it is possible with settings based dashboard to add existing insights to it.
 */
export const EmptySettingsBasedDashboard: React.FunctionComponent<
    React.PropsWithChildren<EmptyInsightDashboardProps>
> = props => {
    const { onAddInsight, dashboard } = props
    const {
        dashboard: { getAddRemoveInsightsPermission },
    } = useUiFeatures()
    const addRemoveInsightPermissions = getAddRemoveInsightsPermission(dashboard)

    return (
        <section className={styles.emptySection}>
            <Button
                type="button"
                disabled={addRemoveInsightPermissions.disabled}
                onClick={onAddInsight}
                variant="secondary"
                className="p-0 w-100 border-0"
            >
                <Card
                    data-tooltip={addRemoveInsightPermissions.tooltip}
                    data-placement="right"
                    className={styles.itemCard}
                >
                    <PlusIcon size="2rem" />
                    <span>Add insights</span>
                </Card>
            </Button>
            <span className="d-flex justify-content-center mt-3">
                <Link to={`/insights/create?dashboardId=${dashboard.id}`}>or, create new insight</Link>
            </span>
        </section>
    )
}
