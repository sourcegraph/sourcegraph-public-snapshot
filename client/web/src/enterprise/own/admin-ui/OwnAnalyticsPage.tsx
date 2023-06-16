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

    if (!data?.instanceOwnershipStats?.totalFiles) {
        return <>{error && <ErrorAlert prefix="Error getting own analytics" error={error} />}</>
    }

    const totalFiles = data.instanceOwnershipStats.totalFiles
    const totalCodeownedFiles = data.instanceOwnershipStats.totalCodeownedFiles
    const totalAssignedOwnershipFiles = data.instanceOwnershipStats.totalAssignedOwnershipFiles
    const totalOwnedFiles = data.instanceOwnershipStats.totalOwnedFiles

    const totalCodeownedFilesPercent = Math.round((totalCodeownedFiles / totalFiles) * 100)
    const totalAssignedOwnershipFilesPercent = Math.round((totalAssignedOwnershipFiles / totalFiles) * 100)
    const totalOwnedFilesPercent = Math.round((totalOwnedFiles / totalFiles) * 100)

    const ownSignalsData: OwnCoverageDatum[] = [
        {
            name: 'CODEOWNERS',
            count: totalCodeownedFilesPercent,
            fill: 'var(--info-2)',
            tooltip: `Files owned through CODEOWNERS:${totalCodeownedFiles}/${totalFiles}`,
        },
        {
            name: 'Assigned ownership',
            count: totalAssignedOwnershipFilesPercent,
            fill: 'var(--info)',
            tooltip: `Files with assigned owners: ${totalAssignedOwnershipFiles}/${totalFiles}`,
        },
        {
            name: 'All owned files',
            count: totalOwnedFilesPercent,
            fill: 'var(--info-3)',
            tooltip: `Owned files: ${totalOwnedFiles}/${totalFiles}`,
        },
        // TODO decide whether we remove or keep all files
        {
            name: 'All files',
            count: 100,
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
                    <Card className="p-3 position-relative">
                        {ownSignalsData && (
                            <div>
                                <ChartContainer
                                    title="Ownership coverage in %"
                                    labelX="Ownership type"
                                    labelY="Files percentage"
                                >
                                    {width => (
                                        <BarChart
                                            width={width}
                                            height={300}
                                            data={ownSignalsData}
                                            getDatumName={datum => datum.name}
                                            getDatumValue={datum => datum.count}
                                            getDatumColor={datum => datum.fill}
                                            getDatumFadeColor={() => 'var(--gray-04)'}
                                            getDatumHover={datum => datum.tooltip}
                                            getDatumHoverValueLabel={datum => `${datum.count}%`}
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
