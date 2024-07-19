import { type FC, useEffect } from 'react'

import { mdiWebhook } from '@mdi/js'

import { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import type { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { PageHeader } from '@sourcegraph/wildcard'

import { PageTitle } from '../components/PageTitle'

import { WebhookCreateUpdatePage } from './WebhookCreateUpdatePage'

export interface SiteAdminWebhookCreatePageProps extends TelemetryProps, TelemetryV2Props {}

export const SiteAdminWebhookCreatePage: FC<SiteAdminWebhookCreatePageProps> = ({
    telemetryService,
    telemetryRecorder,
}) => {
    useEffect(() => {
        telemetryService.logPageView('SiteAdminWebhookCreatePage')
        telemetryRecorder.recordEvent('admin.webhook.create', 'view')
    }, [telemetryService, telemetryRecorder])

    return (
        <>
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
        </>
    )
}
