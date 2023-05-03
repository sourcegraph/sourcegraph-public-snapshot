import { FC, useEffect } from 'react'

import { mdiCog } from '@mdi/js'
import { useParams } from 'react-router-dom'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Container, ErrorAlert, PageHeader, ButtonLink } from '@sourcegraph/wildcard'

import { CreatedByAndUpdatedByInfoByline } from '../Byline/CreatedByAndUpdatedByInfoByline'
import { PageTitle } from '../PageTitle'

interface Props extends TelemetryProps {
    externalServicesFromFile: boolean
    allowEditExternalServicesWithFile: boolean
}

export const GitHubAppPage: FC<Props> = ({
    telemetryService,
    externalServicesFromFile,
    allowEditExternalServicesWithFile,
}) => {
    const { appID } = useParams()

    useEffect(() => {
        telemetryService.logPageView('SiteAdminGitHubApp')
    }, [telemetryService])

    const app: any = null
    const error = null

    return (
        <div>
            {app ? <PageTitle title={`GitHub App - ${app.displayName}`} /> : <PageTitle title="GitHub App" />}
            {error && <ErrorAlert className="mb-3" error={error} />}
            <h1>GitHub App</h1>

            {app && (
                <Container className="mb-3">
                    <PageHeader
                        path={[
                            { icon: mdiCog },
                            { to: '/site-admin/github-apps', text: 'GitHub Apps' },
                            {
                                to: `/site-admin/github-apps/${appID}`,
                                text: app.displayName,
                            },
                        ]}
                        byline={
                            <CreatedByAndUpdatedByInfoByline
                                createdAt={app.createdAt}
                                updatedAt={app.updatedAt}
                                noAuthor={true}
                            />
                        }
                        className="mb-3"
                        headingElement="h2"
                        actions={
                            <ButtonLink to={`/site-admin/github-apps/`} variant="secondary">
                                Cancel
                            </ButtonLink>
                        }
                    />
                </Container>
            )}
        </div>
    )
}
