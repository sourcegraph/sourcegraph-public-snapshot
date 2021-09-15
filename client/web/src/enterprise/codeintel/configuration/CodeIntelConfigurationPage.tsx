import * as H from 'history'
import React, { FunctionComponent, useEffect } from 'react'
import { RouteComponentProps } from 'react-router'

import { TelemetryProps, TelemetryService } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { PageTitle } from '@sourcegraph/web/src/components/PageTitle'
import { PageHeader } from '@sourcegraph/wildcard'

import { RepositoryConfiguration } from './RepositoryConfiguration'
import { RepositoryPolicies } from './RepositoryPolicies'

export interface CodeIntelConfigurationPageProps extends RouteComponentProps<{}>, ThemeProps, TelemetryProps {
    repo?: { id: string }
    indexingEnabled?: boolean
    isLightTheme: boolean
    telemetryService: TelemetryService
    history: H.History
}

export const CodeIntelConfigurationPage: FunctionComponent<CodeIntelConfigurationPageProps> = ({
    repo,
    indexingEnabled = window.context?.codeIntelAutoIndexingEnabled,
    isLightTheme,
    telemetryService,
    history,
}) => {
    useEffect(() => telemetryService.logViewEvent('CodeIntelConfigurationPage'), [telemetryService])

    return (
        <>
            <PageTitle title="Precise code intelligence configuration" />
            <PageHeader
                headingElement="h2"
                path={[
                    {
                        text: <>Precise code intelligence configuration</>,
                    },
                ]}
                description={`Rules that define configuration for precise code intelligence ${
                    repo ? 'in this repository' : 'over all repositories'
                }.`}
                className="mb-3"
            />

            {repo ? (
                <RepositoryConfiguration
                    repo={repo}
                    indexingEnabled={indexingEnabled}
                    isLightTheme={isLightTheme}
                    telemetryService={telemetryService}
                    history={history}
                />
            ) : (
                <RepositoryPolicies repo={repo} isGlobal={true} indexingEnabled={indexingEnabled} history={history} />
            )}
        </>
    )
}
