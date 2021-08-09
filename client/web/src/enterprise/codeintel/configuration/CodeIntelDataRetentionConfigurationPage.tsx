import * as H from 'history'
import React, { FunctionComponent, useEffect } from 'react'
import { RouteComponentProps } from 'react-router'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'

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

    return <h1>Empty</h1>
}
