import { FC, useEffect } from 'react'

import { mdiWebhook } from '@mdi/js'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Container, PageHeader } from '@sourcegraph/wildcard'

import { PageTitle } from '../components/PageTitle'

import { WebhookCreateUpdatePage } from './WebhookCreateUpdatePage'

export interface SiteAdminWebhookCreatePageProps extends TelemetryProps {}

export const SiteAdminWebhookCreatePage: FC<SiteAdminWebhookCreatePageProps> = ({ telemetryService }) => {
    useEffect(() => {
        telemetryService.logPageView('SiteAdminWebhookCreatePage')
    }, [telemetryService])

    return (
        <Container>
            <PageTitle title="Incoming webhook" />
            <PageHeader
                path={[
                    { icon: mdiWebhook },
                    { to: '/site-admin/webhooks/incoming', text: 'Incoming webhooks' },
                    { text: 'Create' },
                ]}
                headingElement="h2"
            />
            <WebhookCreateUpdatePage />
        </Container>
    )
}
