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

interface OwnUsageDatum {
    ownershipReasonType: string
    entriesCount: number
    fill: string
}

export const OwnAnalyticsPage: FC = () => {
    // TODO: Error handling and loading
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

    const ownSignalsData: OwnUsageDatum[] = [
        {
            ownershipReasonType: 'Codeowned files',
            entriesCount: data?.instanceOwnershipStats?.totalCodeownedFiles || 0,
            fill: 'var(--info)',
        },
        {
            ownershipReasonType: 'Total files',
            entriesCount: data?.instanceOwnershipStats?.totalFiles || 0,
            fill: 'var(--text-muted)',
        },
    ]
    const getValue = (datum: OwnUsageDatum) => datum.entriesCount
    const getColor = (datum: OwnUsageDatum) => datum.fill
    const getLink = (datum: OwnUsageDatum) => ''
    const getName = (datum: OwnUsageDatum) => datum.ownershipReasonType

    return (
        <>
            {loading && <LoadingSpinner />}
            {error && <ErrorAlert prefix="Error finding out if own analytics are enabled" error={error} />}
            {!loading && !error && (
                <>
                    <Card className="p-3 position-relative">
                        {ownSignalsData && (
                            <div>
                                <ChartContainer title="Title" labelX="Time" labelY="LabelY">
                                    {width => (
                                        <BarChart
                                            width={width}
                                            height={300}
                                            data={ownSignalsData}
                                            getDatumName={getName}
                                            getDatumValue={getValue}
                                            getDatumColor={getColor}
                                            getDatumLink={getLink}
                                            getDatumHover={datum => `custom text for ${datum.ownershipReasonType}`}
                                        />
                                    )}
                                </ChartContainer>
                            </div>
                        )}
                    </Card>
                    <Text className="font-italic text-center mt-2">
                        All events are generated from entries in the event logs table and are updated every 24 hours.
                    </Text>
                </>
            )}
        </>
    )
}

const OwnEnableAnalytics: FC = () => (
    <Alert variant="info">
        Analytics is not enabled, please <Link to={'/site-admin/own-signal-page'}>enable Own analytics</Link> first to
        see own stats.
    </Alert>
)
