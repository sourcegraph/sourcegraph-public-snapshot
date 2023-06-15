import { FC } from 'react'

import { useQuery } from '@sourcegraph/http-client'
import { Alert, BarChart, Card, ErrorAlert, H3, Link, LoadingSpinner, Text } from '@sourcegraph/wildcard'

import {
    AnalyticsDateRange,
    GetInstanceOwnStatsResult,
    GetInstanceOwnStatsVariables,
    GetOwnSignalConfigurationsResult,
    OwnSignalConfig,
} from '../../../graphql-operations'
import { AnalyticsPageTitle } from '../../../site-admin/analytics/components/AnalyticsPageTitle'
import { ChartContainer } from '../../../site-admin/analytics/components/ChartContainer'
import { HorizontalSelect } from '../../../site-admin/analytics/components/HorizontalSelect'
import { ValueLegendList, ValueLegendListProps } from '../../../site-admin/analytics/components/ValueLegendList'
import { useChartFilters } from '../../../site-admin/analytics/useChartFilters'

import { GET_INSTANCE_OWN_STATS, GET_OWN_JOB_CONFIGURATIONS } from './query'

interface OwnCoverageDatum {
    name: string
    count: number
    fill: string
    tooltip: string
}

export const OwnAnalyticsPage: FC = () => {
    const { data, loading, error } = useQuery<GetOwnSignalConfigurationsResult>(GET_OWN_JOB_CONFIGURATIONS, {})
    const enabled =
        data?.ownSignalConfigurations.some(
            (config: OwnSignalConfig) => config.name === 'analytics' && config.isEnabled
        ) || false
    return (
        <>
            {loading && <LoadingSpinner />}
            {error && <ErrorAlert prefix="Error finding out if own analytics are enabled" error={error} />}
            <AnalyticsPageTitle>Own</AnalyticsPageTitle>
            {enabled ? <OwnAnalyticsPanel /> : <OwnEnableAnalytics />}
        </>
    )
}

const OwnAnalyticsPanel: FC = () => {
    const { dateRange } = useChartFilters({ name: 'Own' })
    const { data, loading, error } = useQuery<GetInstanceOwnStatsResult, GetInstanceOwnStatsVariables>(
        GET_INSTANCE_OWN_STATS,
        {
            variables: {
                dateRange: dateRange.value,
            },
        }
    )

    const legends: ValueLegendListProps['items'] = [
        {
            value: data?.ownershipUsageStats.fileHasOwnerSearches,
            description: 'Filter by owner searches',
            color: 'var(--cyan)',
            tooltip: 'The number of times search with file:has.owner(...) was issued.',
        },
        {
            value: data?.ownershipUsageStats.selectFileOwnersSearches,
            description: 'Select owner searches',
            color: 'var(--orange)',
            tooltip: 'The number of times search with select:file.owners was issued.',
        },
        {
            value: data?.ownershipUsageStats.ownershipPanelViewed,
            description: 'Ownership panel views',
            color: 'var(--body-color)',
            position: 'right',
            tooltip: 'The number of times ownership panel was opened',
        },
    ]

    const ownSignalsData: OwnCoverageDatum[] = [
        {
            name: 'CODEOWNERS',
            count: data?.instanceOwnershipStats?.totalCodeownedFiles || 0,
            fill: 'var(--info-2)',
            tooltip: 'Total number of files owned through CODEOWNERS',
        },
        {
            name: 'Assigned ownership',
            count: data?.instanceOwnershipStats?.totalAssignedOwnershipFiles || 0,
            fill: 'var(--info)',
            tooltip: 'Total number of files with assigned owners',
        },
        {
            name: 'All owned files',
            count: data?.instanceOwnershipStats?.totalOwnedFiles || 0,
            fill: 'var(--info-3)',
            tooltip: 'Total number of owned files',
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
            {error && <ErrorAlert prefix="Error getting own analytics" error={error} />}
            {!loading && !error && (
                <>
                    {/* TODO(#52826): If only partial data is available - make that clear to the user. */}
                    {legends && (
                        <>
                            <H3>Usage</H3>
                            <Card className="p-3 position-relative mb-3">
                                <div className="d-flex justify-content-end align-items-stretch mb-2 text-nowrap">
                                    <HorizontalSelect<typeof dateRange.value> {...dateRange} />
                                </div>
                                <ValueLegendList className="mb-3" items={legends} />
                            </Card>
                        </>
                    )}

                    {ownSignalsData && (
                        <>
                            <H3>Coverage</H3>
                            <Card className="p-3 position-relative">
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
                            </Card>
                        </>
                    )}
                    <Text className="font-italic text-center mt-2">
                        {/* TODO(#52826): Provide more precise information about how stale data is, and how often it refreshes. */}
                        Data is generated periodically from CODEOWNERS files, assigned ownership data, repository
                        contents and event logs.
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
