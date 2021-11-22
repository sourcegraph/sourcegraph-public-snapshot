import * as H from 'history'
import React, { FunctionComponent, useEffect } from 'react'

import { Link } from '@sourcegraph/shared/src/components/Link'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { PageTitle } from '@sourcegraph/web/src/components/PageTitle'
import { PageHeader } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../../../auth'
import { CodeIntelConfigurationPageHeader } from '../components/CodeIntelConfigurationPageHeader'
import { ConfigurationEditor } from '../components/ConfigurationEditor'

export interface RepositoryIndexConfigurationPageProps extends ThemeProps, TelemetryProps {
    repo: { id: string }
    authenticatedUser: AuthenticatedUser | null
    history: H.History
}

export const RepositoryIndexConfigurationPage: FunctionComponent<RepositoryIndexConfigurationPageProps> = ({
    repo,
    authenticatedUser,
    history,
    telemetryService,
    ...props
}) => {
    useEffect(() => telemetryService.logViewEvent('CodeIntelConfigurationPage'), [telemetryService])

    return (
        <>
            <PageTitle title="Precise code intelligence repository index configuration" />
            <CodeIntelConfigurationPageHeader>
                <PageHeader
                    headingElement="h2"
                    path={[
                        {
                            text: <>Precise code intelligence repository index configuration</>,
                        },
                    ]}
                    description={
                        <>
                            Provide explicit index job configuration to customize how this repository is indexed. See
                            the{' '}
                            <Link to="https://docs.sourcegraph.com/code_intelligence/references/auto_indexing_configuration">
                                reference guide
                            </Link>{' '}
                            for more information.
                        </>
                    }
                    className="mb-3"
                />
            </CodeIntelConfigurationPageHeader>

            <ConfigurationEditor
                repoId={repo.id}
                authenticatedUser={authenticatedUser}
                history={history}
                telemetryService={telemetryService}
                {...props}
            />
        </>
    )
}
