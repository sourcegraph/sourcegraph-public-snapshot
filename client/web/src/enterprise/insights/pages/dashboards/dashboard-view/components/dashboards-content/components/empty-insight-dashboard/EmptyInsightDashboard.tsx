import type { FC } from 'react'

import { mdiPlus } from '@mdi/js'

import { Button, Link, Card, Tooltip, Icon } from '@sourcegraph/wildcard'

import type { CustomInsightDashboard } from '../../../../../../../core'
import { useUiFeatures } from '../../../../../../../hooks'
import { encodeDashboardIdQueryParam } from '../../../../../../../routers.constant'

import styles from './EmptyInsightDashboard.module.scss'

interface EmptyCustomDashboardProps {
    dashboard: CustomInsightDashboard
    onAddInsightRequest?: () => void
}

/**
 * Custom empty dashboard state provides ability to add existing insights to the dashboard.
 */
export const EmptyCustomDashboard: FC<EmptyCustomDashboardProps> = props => {
    const { dashboard, onAddInsightRequest } = props

    const {
        dashboard: { getAddRemoveInsightsPermission },
    } = useUiFeatures()
    const permissions = getAddRemoveInsightsPermission(dashboard)

    return (
        <section className={styles.emptySection}>
            <Button
                type="button"
                disabled={permissions.disabled}
                variant="secondary"
                className="p-0 w-100 border-0"
                data-testid="add-insights-button-card"
                onClick={onAddInsightRequest}
            >
                <Tooltip content={permissions.tooltip} placement="right">
                    <Card className={styles.itemCard}>
                        <Icon svgPath={mdiPlus} inline={false} aria-hidden={true} height="2rem" width="2rem" />
                        <span>Add insights</span>
                    </Card>
                </Tooltip>
            </Button>
            <span className="d-flex justify-content-center mt-3">
                <Link to={encodeDashboardIdQueryParam('/insights/create', dashboard.id)}>or, create new insight</Link>
            </span>
        </section>
    )
}
