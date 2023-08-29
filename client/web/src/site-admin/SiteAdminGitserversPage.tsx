import { type FC, useEffect } from 'react'

import { mdiServer } from '@mdi/js'

import type { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Container, PageHeader } from '@sourcegraph/wildcard'

import { PageTitle } from '../components/PageTitle'

import styles from './SiteAdminGitserversPage.module.scss'

export interface GitserversPageProps extends TelemetryProps {}

export const SiteAdminGitserversPage: FC<GitserversPageProps> = ({ telemetryService }) => {
    useEffect(() => {
        telemetryService.logPageView('SiteAdminWebhook')
    }, [telemetryService])
    return (
        <Container>
            <PageTitle title="Gitservers" />
            <PageHeader
                path={[{ icon: mdiServer }, { text: 'Gitservers' }]}
                className="mb-3"
                headingElement="h2"
                description="Manage your Gitservers"
            />

            <div>Git servers content will go here</div>
        </Container>
    )
}
