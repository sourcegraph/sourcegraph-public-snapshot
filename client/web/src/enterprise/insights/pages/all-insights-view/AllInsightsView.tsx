import { FC } from 'react'

import { useQuery } from '@apollo/client'
import { mdiPlus } from '@mdi/js'

import { isDefined } from '@sourcegraph/common'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Card, ErrorAlert, Icon, Link, LoadingSpinner } from '@sourcegraph/wildcard'

import { useShowMorePagination } from '../../../../components/FilteredConnection/hooks/useShowMorePagination';
import { GetAllInsightConfigurationsResult, GetAllInsightConfigurationsVariables } from '../../../../graphql-operations'
import { SmartInsightsViewGrid } from '../../components'
import { createInsightView } from '../../core/backend/gql-backend'

import { GET_ALL_INSIGHT_CONFIGURATIONS } from './query'

import styles from './AllInsightsView.module.scss'

interface AllInsightsViewProps extends TelemetryProps {}

export const AllInsightsView: FC<AllInsightsViewProps> = props => {
    const {} = useShowMorePagination<
        GetAllInsightConfigurationsResult,
        GetAllInsightConfigurationsVariables,
        any
    >({
        query: GET_ALL_INSIGHT_CONFIGURATIONS,
        variables: { first: 15, after: null, },
        getConnection: result => {},
        options: { fetchPolicy: 'cache-first' }
    })

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
        <div>
            <SmartInsightsViewGrid insights={insightConfigurations} telemetryService={props.telemetryService} />
        </div>
    ) : (
        <Card as={Link} to="/insights/create" className={styles.emptyCard}>
            <Icon svgPath={mdiPlus} inline={false} aria-hidden={true} height="2rem" width="2rem" />
            <span>It seems that you don't have any insights yet, you can create your first insight from here.</span>
        </Card>
    )
}
