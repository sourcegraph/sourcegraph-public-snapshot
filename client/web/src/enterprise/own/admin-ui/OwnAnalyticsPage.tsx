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
    percent: number
    absoluteCount: number
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
    const { data, loading, error } = useQuery<GetInstanceOwnStatsResult>(GET_INSTANCE_OWN_STATS, {})

    const totalFilesNonZero: number = data?.instanceOwnershipStats?.totalFiles || 1
    const ratioPerCent = (count: number): number => Math.round((10000 * count) / totalFilesNonZero) / 100
    const ownSignalsData: OwnCoverageDatum[] = [
        {
            name: 'CODEOWNERS coverage',
            percent: ratioPerCent(data?.instanceOwnershipStats?.totalCodeownedFiles || 0),
            absoluteCount: data?.instanceOwnershipStats?.totalCodeownedFiles || 0,
            fill: 'var(--info-2)',
            tooltip: 'Files with owners in CODEOWNERS',
        },
        {
            name: 'Assigned Ownership coverage',
            percent: ratioPerCent(data?.instanceOwnershipStats?.totalAssignedOwnershipFiles || 0),
            absoluteCount: data?.instanceOwnershipStats?.totalAssignedOwnershipFiles || 0,
            fill: 'var(--merged)',
            tooltip: 'Files with Assigned Ownership',
        },
        {
            name: 'Any ownership coverage',
            percent: ratioPerCent(data?.instanceOwnershipStats?.totalOwnedFiles || 0),
            absoluteCount: data?.instanceOwnershipStats?.totalOwnedFiles || 0,
            fill: 'var(--info-3)',
            tooltip: 'Files that have an owner',
        },
        {
            name: 'All files',
            percent: ratioPerCent(data?.instanceOwnershipStats?.totalFiles || 0),
            absoluteCount: data?.instanceOwnershipStats?.totalFiles || 0,
            fill: 'var(--text-muted)',
            tooltip: 'All files',
        },
    ]

    return (
        <>
            {loading && <LoadingSpinner />}
            {error && <ErrorAlert prefix="Error getting own analytics" error={error} />}
            {!loading && !error && (
                <>
                    {/* TODO(#52826): If only partial data is available - make that clear to the user. */}
                    <Card className="p-3 position-relative">
                        {ownSignalsData && (
                            <div>
                                <ChartContainer title="Ownership coverage" labelX="Status" labelY="% files">
                                    {width => (
                                        <BarChart
                                            width={width}
                                            height={300}
                                            data={ownSignalsData}
                                            getDatumName={datum => datum.name}
                                            getDatumValue={datum => datum.percent}
                                            getDatumHoverValueLabel={datum => `${datum.absoluteCount}`}
                                            getDatumColor={datum => datum.fill}
                                            getDatumLink={() => ''}
                                            getDatumHover={datum => `${datum.percent}% ${datum.tooltip}`}
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
