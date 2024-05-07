import React, { useEffect } from 'react'

import { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import { Card, Link, Text } from '@sourcegraph/wildcard'

import { AnalyticsPageTitle } from '../components/AnalyticsPageTitle'

interface Props extends TelemetryV2Props {}

export const AnalyticsCodyPage: React.FC<Props> = ({ telemetryRecorder }) => {
    useEffect(() => telemetryRecorder.recordEvent('admin.analytics.cody', 'view'), [telemetryRecorder])

    return (
        <>
            <AnalyticsPageTitle>Cody</AnalyticsPageTitle>

            <Card className="p-3">
                <Text>
                    Cody analytics, including active users, completions, chat, and commands can be found at{' '}
                    <Link to="https://cody-analytics.sourcegraph.com" target="_blank" rel="noopener">
                        cody-analytics.sourcegraph.com
                    </Link>
                    .
                </Text>
                <Text>To request access, please contact your account team.</Text>
            </Card>
        </>
    )
}
