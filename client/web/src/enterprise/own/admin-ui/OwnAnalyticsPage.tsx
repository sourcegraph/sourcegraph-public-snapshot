import { FC } from 'react'

import { useQuery } from '@sourcegraph/http-client'
import { Alert, BarChart, Card, ErrorAlert, Link, LoadingSpinner, Text } from '@sourcegraph/wildcard'

import {
    GetInstanceOwnStatsResult,
    GetOwnSignalConfigurationsResult,
    OwnSignalConfig,
} from '../../../graphql-operations'
import { AnalyticsPageTitle } from '../../../site-admin/analytics/components/AnalyticsPageTitle'
import { ChartContainer } from '../../../site-admin/analytics/components/ChartContainer'

import { GET_INSTANCE_OWN_STATS, GET_OWN_JOB_CONFIGURATIONS } from './query'

interface OwnCoverageDatum {
    name: string
    count: number
    fill: string
    tooltip: string
}

export const OwnAnalyticsPage: FC = () => {
    // TODO(#52826): Error handling and loading
    const { data } = useQuery<GetOwnSignalConfigurationsResult>(GET_OWN_JOB_CONFIGURATIONS, {})
    const enabled =
        data?.ownSignalConfigurations.some(
            (config: OwnSignalConfig) => config.name === 'analytics' && config.isEnabled
        ) || false
    return (
        <>
            <AnalyticsPageTitle>Own</AnalyticsPageTitle>
            {enabled ? <OwnAnalyticsPanel /> : <OwnEnableAnalytics />}
        </>
    )
}

const OwnAnalyticsPanel: FC = () => {
    const { data, loading, error } = useQuery<GetInstanceOwnStatsResult>(GET_INSTANCE_OWN_STATS, {})

    const ownSignalsData: OwnCoverageDatum[] = [
        {
            name: 'CODEOWNERS',
            count: data?.instanceOwnershipStats?.totalCodeownedFiles || 0,
            fill: 'var(--info)',
            tooltip: 'Total number of files owned through CODEOWNERS',
        },
        {
            name: 'All files',
            count: data?.instanceOwnershipStats?.totalFiles || 0,
            fill: 'var(--text-muted)',
            tooltip: 'Total number of files',
        },
    ]

    return (
        <>
            {loading && <LoadingSpinner />}
            {error && <ErrorAlert prefix="Error finding out if own analytics are enabled" error={error} />}
            {!loading && !error && (
                <>
                    {/* TODO(#52826): If only partial data is available - make that clear to the user. */}
                    <Card className="p-3 position-relative">
                        {ownSignalsData && (
                            <div>
                                <ChartContainer title="Ownership coverage" labelX="Status" labelY="File count">
                                    {width => (
                                        <BarChart
                                            width={width}
                                            height={300}
                                            data={ownSignalsData}
                                            getDatumName={datum => datum.name}
                                            getDatumValue={datum => datum.count}
                                            getDatumColor={datum => datum.fill}
                                            getDatumLink={datum => ''}
                                            getDatumHover={datum => datum.tooltip}
                                        />
                                    )}
                                </ChartContainer>
                            </div>
                        )}
                    </Card>
                    <Text className="font-italic text-center mt-2">
                        {/* TODO(#52826): Provide more precise information about how stale data is, and how often it refreshes. */}
                        Data is generated periodically from CODEOWNERS files and repository contents.
                    </Text>
                </>
            )}
        </>
    )
}

const OwnEnableAnalytics: FC = () => (
    <Alert variant="info">
        Analytics is not enabled, please <Link to="/site-admin/own-signal-page">enable Own analytics</Link> job first to
        see Own stats.
    </Alert>
)
