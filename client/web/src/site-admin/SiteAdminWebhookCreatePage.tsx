import { type FC, useEffect } from 'react'

import { mdiWebhook } from '@mdi/js'

import type { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Container, PageHeader } from '@sourcegraph/wildcard'

import { PageTitle } from '../components/PageTitle'

import { WebhookCreateUpdatePage } from './WebhookCreateUpdatePage'

export interface SiteAdminWebhookCreatePageProps extends TelemetryProps {}

export const SiteAdminWebhookCreatePage: FC<SiteAdminWebhookCreatePageProps> = ({
    telemetryService,
    telemetryRecorder,
}) => {
    useEffect(() => {
        telemetryService.logPageView('SiteAdminWebhookCreatePage')
        telemetryRecorder.recordEvent('siteAdminWebhookCreatePage', 'viewed')
    }, [telemetryService, telemetryRecorder])

    return (
        <Container>
            <PageTitle title="Create incoming webhook" />
            <PageHeader
                path={[
                    { icon: mdiWebhook },
                    { to: '/site-admin/webhooks/incoming', text: 'Incoming webhooks' },
                    { text: 'Create' },
                ]}
                headingElement="h2"
                description="Create a new incoming webhook"
                className="mb-3"
            />
            <WebhookCreateUpdatePage />
        </Container>
    )
}
