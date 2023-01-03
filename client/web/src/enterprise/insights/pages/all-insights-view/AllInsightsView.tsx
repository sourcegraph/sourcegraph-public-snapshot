import { FC } from 'react'

import { useQuery } from '@apollo/client'
import { mdiPlus } from '@mdi/js'

import { isDefined } from '@sourcegraph/common'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Card, ErrorAlert, Icon, Link, LoadingSpinner } from '@sourcegraph/wildcard'

import { GetAllInsightConfigurationsResult } from '../../../../graphql-operations'
import { SmartInsightsViewGrid } from '../../components'
import { createInsightView } from '../../core/backend/gql-backend'

import { GET_ALL_INSIGHT_CONFIGURATIONS } from './query'

import styles from './AllInsightsView.module.scss'

interface AllInsightsViewProps extends TelemetryProps {}

export const AllInsightsView: FC<AllInsightsViewProps> = props => {
    const { data, error } = useQuery<GetAllInsightConfigurationsResult>(GET_ALL_INSIGHT_CONFIGURATIONS, {
        nextFetchPolicy: 'cache-first',
        errorPolicy: 'all',
    })

    if (data === undefined) {
        return <LoadingSpinner aria-hidden={true} inline={false} />
    }

    if (error) {
        return <ErrorAlert error={error} />
    }

    const insightConfigurations = data.insightViews.nodes.filter(isDefined).map(createInsightView)

    return insightConfigurations.length > 0 ? (
        <SmartInsightsViewGrid insights={insightConfigurations} telemetryService={props.telemetryService} />
    ) : (
        <EmptyVirtualDashboard />
    )
}

/**
 * Virtual empty dashboard state provides link to create a new code insight via creation UI.
 * Since all insights within virtual dashboards are calculated there's no ability to add insight to
 * this type of dashboard manually.
 */
export const EmptyVirtualDashboard: FC = () => (
    <Card as={Link} to="/insights/create" className={styles.emptyCard}>
        <Icon svgPath={mdiPlus} inline={false} aria-hidden={true} height="2rem" width="2rem" />
        <span>It seems that you don't have any insights yet, you can create your first insight from here.</span>
    </Card>
)
