import * as H from 'history'
import React, { FunctionComponent, useState, useEffect } from 'react'
import { RouteComponentProps } from 'react-router'

import { TelemetryProps, TelemetryService } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { PageTitle } from '@sourcegraph/web/src/components/PageTitle'
import { PageHeader } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../../auth'

import { CodeIntelConfigurationPageHeader } from './CodeIntelConfigurationPageHeader'
import { FlashMessage } from './FlashMessage'
import { PolicyListActions } from './PolicyListActions'
import { RepositoryConfiguration } from './RepositoryConfiguration'
import { RepositoryPolicies } from './RepositoryPolicies'

export interface CodeIntelConfigurationPageProps extends RouteComponentProps<{}>, ThemeProps, TelemetryProps {
    authenticatedUser: AuthenticatedUser | null
    repo?: { id: string }
    indexingEnabled?: boolean
    isLightTheme: boolean
    telemetryService: TelemetryService
    history: H.History
}

export const CodeIntelConfigurationPage: FunctionComponent<CodeIntelConfigurationPageProps> = ({
    authenticatedUser,
    repo,
    indexingEnabled = window.context?.codeIntelAutoIndexingEnabled,
    isLightTheme,
    telemetryService,
    history,
}) => {
    useEffect(() => telemetryService.logViewEvent('CodeIntelConfigurationPage'), [telemetryService])
    const [displayActions, setDisplayAction] = useState(true)
    const [isDeleting, setIsDeleting] = useState(false)
    const [isLoading, setIsLoading] = useState(false)

    return (
        <>
            <PageTitle title="Precise code intelligence configuration" />
            <CodeIntelConfigurationPageHeader>
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
                {displayActions && authenticatedUser?.siteAdmin && (
                    <>
                        <PolicyListActions disabled={isLoading} deleting={isDeleting} history={history} />
                    </>
                )}
            </CodeIntelConfigurationPageHeader>

            {history.location.state && (
                <FlashMessage state={history.location.state.modal} message={history.location.state.message} />
            )}

            {repo ? (
                <RepositoryConfiguration
                    authenticatedUser={authenticatedUser}
                    repo={repo}
                    indexingEnabled={indexingEnabled}
                    isLightTheme={isLightTheme}
                    telemetryService={telemetryService}
                    history={history}
                    onHandleDisplayAction={setDisplayAction}
                    onHandleIsDeleting={setIsDeleting}
                    onHandleIsLoading={setIsLoading}
                />
            ) : (
                <RepositoryPolicies
                    authenticatedUser={authenticatedUser}
                    repo={repo}
                    isGlobal={true}
                    indexingEnabled={indexingEnabled}
                    history={history}
                    onHandleDisplayAction={setDisplayAction}
                    onHandleIsDeleting={setIsDeleting}
                    onHandleIsLoading={setIsLoading}
                />
            )}
        </>
    )
}
