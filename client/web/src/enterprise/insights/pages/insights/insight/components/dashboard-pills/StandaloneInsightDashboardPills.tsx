import type { FunctionComponent, HTMLAttributes } from 'react'

import { mdiViewDashboard } from '@mdi/js'
import classNames from 'classnames'

import type { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Button, Icon, Link, Text } from '@sourcegraph/wildcard'

import type { InsightDashboardReference } from '../../../../../core'

import styles from './StandaloneInsightDashboardPills.module.scss'

interface StandaloneInsightDashboardPillsProps extends HTMLAttributes<HTMLDivElement>, TelemetryProps {
    dashboards: InsightDashboardReference[]
    insightId: string
}

export const StandaloneInsightDashboardPills: FunctionComponent<StandaloneInsightDashboardPillsProps> = props => {
    const { dashboards, insightId, className, telemetryService, ...attributes } = props

    const handleDashboardClick = (): void => {
        telemetryService.log('StandaloneInsightDashboardClick')
    }

    return (
        <div {...attributes} className={classNames(className, styles.list)}>
            <Text size="small" className={styles.title}>
                Insight added to:
            </Text>

            <Button
                as={Link}
                to={`/insights/all?focused=${insightId}`}
                variant="secondary"
                outline={true}
                size="sm"
                target="_blank"
                rel="noopener"
                className={styles.pill}
                onClick={handleDashboardClick}
            >
                <Icon aria-hidden={true} svgPath={mdiViewDashboard} />
                All Insights
            </Button>

            {dashboards.map(dashboard => (
                <Button
                    key={dashboard.id}
                    as={Link}
                    to={`/insights/dashboards/${dashboard.id}?focused=${insightId}`}
                    variant="secondary"
                    outline={true}
                    size="sm"
                    target="_blank"
                    rel="noopener"
                    className={styles.pill}
                    onClick={handleDashboardClick}
                >
                    <Icon aria-hidden={true} svgPath={mdiViewDashboard} />
                    {dashboard.title}
                </Button>
            ))}
        </div>
    )
}
