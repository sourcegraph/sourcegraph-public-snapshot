import { FunctionComponent, HTMLAttributes } from 'react'

import classNames from 'classnames'
import ViewDashboardIcon from 'mdi-react/ViewDashboardIcon'

import { Button, Icon, Link, Text } from '@sourcegraph/wildcard'

import { ALL_INSIGHTS_DASHBOARD, InsightDashboardReference } from '../../../../../core'

import styles from './StandaloneInsightDashboardPills.module.scss'

interface StandaloneInsightDashboardPillsProps extends HTMLAttributes<HTMLDivElement> {
    dashboards: InsightDashboardReference[]
    insightId: string
}

export const StandaloneInsightDashboardPills: FunctionComponent<StandaloneInsightDashboardPillsProps> = props => {
    const { dashboards, insightId, className, ...attributes } = props

    return (
        <div {...attributes} className={classNames(className, styles.list)}>
            <Text size="small" className={styles.title}>
                Insight added to:
            </Text>

            {[ALL_INSIGHTS_DASHBOARD, ...dashboards].map(dashboard => (
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
                >
                    <Icon as={ViewDashboardIcon} />
                    {dashboard.title}
                </Button>
            ))}
        </div>
    )
}
