import * as H from 'history'
import React, { FunctionComponent, useEffect } from 'react'
import { RouteComponentProps } from 'react-router'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { PageTitle } from '@sourcegraph/web/src/components/PageTitle'
import { Container, PageHeader } from '@sourcegraph/wildcard'

export interface CodeIntelDataRetentionConfigurationPageProps
    extends RouteComponentProps<{}>,
        ThemeProps,
        TelemetryProps {
    repo: { id: string }
    history: H.History
}

export const CodeIntelDataRetentionConfigurationPage: FunctionComponent<CodeIntelDataRetentionConfigurationPageProps> = ({
    telemetryService,
}) => {
    useEffect(() => telemetryService.logViewEvent('CodeIntelDataRetentionConfigurationPage'), [telemetryService])

    return (
        <div className="code-intel-index-configuration">
            <PageTitle title="Code intelligence data retention configuration" />

            <PageHeader
                headingElement="h2"
                path={[
                    {
                        text: <>Code intelligence data retention configuration</>,
                    },
                ]}
                description="TODO"
                className="mb-3"
            />

            <Container>TODO</Container>
        </div>
    )
}
