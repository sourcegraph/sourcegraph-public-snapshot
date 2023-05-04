import { FC, useEffect } from 'react'

import { mdiCog, mdiChevronLeft } from '@mdi/js'
import { useParams } from 'react-router-dom'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Container, ErrorAlert, PageHeader, ButtonLink, Icon } from '@sourcegraph/wildcard'

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

    if (!appID) {
        return null
    }

    const app: any = {
        id: atob(appID).replace('GitHubApp:', ''),
        name: appID,
        createdAt: '2021-07-01T00:00:00Z',
        updatedAt: '2023-04-04T12:35:21Z',
    }
    const error = null

    return (
        <div>
            {app ? <PageTitle title={`GitHub App - ${app.name}`} /> : <PageTitle title="GitHub App" />}
            {error && <ErrorAlert className="mb-3" error={error} />}
            {app && (
                <Container className="mb-3">
                    <PageHeader
                        path={[
                            { icon: mdiCog },
                            { to: '/site-admin/github-apps', text: 'GitHub Apps' },
                            {
                                to: `/site-admin/github-apps/${appID}`,
                                text: app.name,
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
                            <ButtonLink to="/site-admin/github-apps/" variant="secondary">
                                <Icon aria-hidden={true} className="mr-1" svgPath={mdiChevronLeft} />
                                Back
                            </ButtonLink>
                        }
                    />
                    <hr />
                    ID: {app.id}
                </Container>
            )}
        </div>
    )
}
