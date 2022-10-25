import React, { useEffect } from 'react'

import { RouteComponentProps } from 'react-router'
import { WebhookFields, WebhooksListResult, WebhooksListVariables } from 'src/graphql-operations'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Text, Container, PageHeader } from '@sourcegraph/wildcard'

import { useConnection } from '../components/FilteredConnection/hooks/useConnection'
import { PageTitle } from '../components/PageTitle'

import { WEBHOOKS } from './backend'

interface Props extends RouteComponentProps<{}>, TelemetryProps {}

export const SiteAdminWebhooksPage: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    history,
    location,
    telemetryService,
}) => {
    useEffect(() => {
        telemetryService.logPageView('SiteAdminWebhooks')
    }, [telemetryService])

    const { connection } = useConnection<WebhooksListResult, WebhooksListVariables, WebhookFields>({
        query: WEBHOOKS,
        variables: {},
        getConnection: result => {
            if (!result) {
                return
            }
            if (!result.data?.webhooks) {
                console.error('no')
                return
            }
            return result.data.webhooks
        },
    })

    console.log('connection', connection)

    // See: https://github.com/sourcegraph/sourcegraph/pull/39134/files
    return (
        <div className="site-admin-webhooks-page">
            <PageTitle title="Webhooks - Admin" />
            <PageHeader
                path={[{ text: 'Webhooks' }]}
                headingElement="h2"
                description="All your webhooks, yo"
                className="mb-3"
            />

            <Container className="mb-3">
                <Text>Horsegraph</Text>
            </Container>
        </div>
    )
}
