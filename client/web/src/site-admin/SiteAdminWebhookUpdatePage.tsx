import { FC, useEffect } from 'react'

import { mdiCog } from '@mdi/js'
import { RouteComponentProps } from 'react-router'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Container, PageHeader } from '@sourcegraph/wildcard'

import { CreatedByAndUpdatedByInfoByline } from '../components/Byline/CreatedByAndUpdatedByInfoByline'
import { ConnectionLoading } from '../components/FilteredConnection/ui'
import { PageTitle } from '../components/PageTitle'

import { useWebhookQuery } from './backend'
import { WebhookCreateUpdatePage } from './WebhookCreateUpdatePage'

export interface SiteAdminWebhookUpdatePageProps extends TelemetryProps, RouteComponentProps<{ id: string }> {}

export const SiteAdminWebhookUpdatePage: FC<SiteAdminWebhookUpdatePageProps> = ({
    match: {
        params: { id },
    },
    telemetryService,
    history,
}) => {
    useEffect(() => {
        telemetryService.logPageView('SiteAdminWebhookUpdatePage')
    }, [telemetryService])

    const { loading, data } = useWebhookQuery(id)

    return (
        <Container>
            <PageTitle title="Incoming webhook" />
            {loading && !data && <ConnectionLoading />}
            {data?.node && data.node.__typename === 'Webhook' && (
                <PageHeader
                    path={[
                        { icon: mdiCog },
                        { to: '/site-admin/webhooks', text: 'Incoming webhooks' },
                        { text: data.node.name },
                    ]}
                    byline={
                        <CreatedByAndUpdatedByInfoByline
                            createdAt={data.node.createdAt}
                            createdBy={data.node.createdBy}
                            updatedAt={data.node.updatedAt}
                            updatedBy={data.node.updatedBy}
                        />
                    }
                    className="mb-3"
                    headingElement="h2"
                />
            )}
            {data?.node && data.node.__typename === 'Webhook' && (
                <WebhookCreateUpdatePage existingWebhook={data.node} history={history} />
            )}
        </Container>
    )
}
