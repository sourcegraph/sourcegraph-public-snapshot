import { FunctionComponent, useEffect } from 'react'

import * as H from 'history'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { PageHeader, Link } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../../../auth'
import { PageTitle } from '../../../../components/PageTitle'
import { CodeIntelConfigurationPageHeader } from '../components/CodeIntelConfigurationPageHeader'
import { ConfigurationEditor } from '../components/ConfigurationEditor'

export interface CodeIntelRepositoryIndexConfigurationPageProps extends ThemeProps, TelemetryProps {
    repo: { id: string }
    authenticatedUser: AuthenticatedUser | null
    history: H.History
}

export const CodeIntelRepositoryIndexConfigurationPage: FunctionComponent<
    React.PropsWithChildren<CodeIntelRepositoryIndexConfigurationPageProps>
> = ({ repo, authenticatedUser, history, telemetryService, ...props }) => {
    useEffect(() => telemetryService.logViewEvent('CodeIntelRepositoryIndexConfiguration'), [telemetryService])

    return (
        <>
            <PageTitle title="Code graph data repository index configuration" />
            <CodeIntelConfigurationPageHeader>
                <PageHeader
                    headingElement="h2"
                    path={[
                        {
                            text: <>Code graph data repository index configuration</>,
                        },
                    ]}
                    description={
                        <>
                            Provide explicit index job configuration to customize how this repository is indexed. See
                            the{' '}
                            <Link to="/help/code_navigation/references/auto_indexing_configuration">
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
