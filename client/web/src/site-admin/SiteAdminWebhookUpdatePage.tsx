import { type FC, useEffect } from 'react'

import { mdiWebhook } from '@mdi/js'
import { useParams } from 'react-router-dom'

import { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import type { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { PageHeader } from '@sourcegraph/wildcard'

import { CreatedByAndUpdatedByInfoByline } from '../components/Byline/CreatedByAndUpdatedByInfoByline'
import { ConnectionLoading } from '../components/FilteredConnection/ui'
import { PageTitle } from '../components/PageTitle'

import { useWebhookQuery } from './backend'
import { WebhookCreateUpdatePage } from './WebhookCreateUpdatePage'

export interface SiteAdminWebhookUpdatePageProps extends TelemetryProps, TelemetryV2Props {}

export const SiteAdminWebhookUpdatePage: FC<SiteAdminWebhookUpdatePageProps> = ({
    telemetryService,
    telemetryRecorder,
}) => {
    useEffect(() => {
        telemetryService.logPageView('SiteAdminWebhookUpdatePage')
        telemetryRecorder.recordEvent('admin.webhook.update', 'view')
    }, [telemetryService, telemetryRecorder])

    const { id = '' } = useParams<{ id: string }>()

    const { loading, data } = useWebhookQuery(id)

    const webhook = data?.node && data.node.__typename === 'Webhook' ? data.node : undefined

    return (
        <>
            <PageTitle title="Edit incoming webhook" />
            {loading && !data && <ConnectionLoading />}
            {webhook && (
                <>
                    <PageHeader
                        path={[
                            { icon: mdiWebhook },
                            { to: '/site-admin/webhooks/incoming', text: 'Incoming webhooks' },
                            { text: webhook.name },
                        ]}
                        byline={
                            <CreatedByAndUpdatedByInfoByline
                                createdAt={webhook.createdAt}
                                createdBy={webhook.createdBy}
                                updatedAt={webhook.updatedAt}
                                updatedBy={webhook.updatedBy}
                            />
                        }
                        className="mb-3"
                        headingElement="h2"
                    />
                    <WebhookCreateUpdatePage existingWebhook={webhook} />
                </>
            )}
        </>
    )
}
