import React, { useEffect } from 'react'

import { mdiCog, mdiMapSearch, mdiPlus } from '@mdi/js'
import { RouteComponentProps } from 'react-router'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ButtonLink, Container, H5, Icon, PageHeader } from '@sourcegraph/wildcard'

import {
    ConnectionContainer,
    ConnectionError,
    ConnectionList,
    ConnectionLoading,
    ConnectionSummary,
    ShowMoreButton,
    SummaryContainer,
} from '../components/FilteredConnection/ui'
import { PageTitle } from '../components/PageTitle'

import { useWebhooksConnection, useWebhookPageHeader } from './backend'
import { WebhookNode } from './WebhookNode'
import { PerformanceGauge } from './webhooks/PerformanceGauge'

import styles from './SiteAdminWebhooksPage.module.scss'

interface Props extends RouteComponentProps<{}>, TelemetryProps {}

export const SiteAdminWebhooksPage: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    telemetryService,
}) => {
    useEffect(() => {
        telemetryService.logPageView('SiteAdminWebhooks')
    }, [telemetryService])

    const { loading, hasNextPage, fetchMore, connection, error } = useWebhooksConnection()
    const headerTotals = useWebhookPageHeader()
    return (
        <div className="site-admin-webhooks-page">
            <PageTitle title="Incoming webhooks" />
            <PageHeader
                path={[{ icon: mdiCog }, { to: '/site-admin/webhooks', text: 'Incoming webhooks' }]}
                headingElement="h2"
                description="All configured incoming webhooks"
                className="mb-3"
                actions={
                    <ButtonLink to="/site-admin/webhooks/create" className="test-create-webhook" variant="primary">
                        <Icon aria-hidden={true} svgPath={mdiPlus} /> Add webhook
                    </ButtonLink>
                }
            />

            <Container>
                {!headerTotals.loading && (
                    <div className={styles.grid}>
                        <PerformanceGauge
                            count={headerTotals.totalErrors}
                            countClassName={headerTotals.totalErrors > 0 ? 'text-danger' : ''}
                            label="error"
                        />
                        <PerformanceGauge
                            count={headerTotals.totalNoEvents}
                            countClassName={headerTotals.totalNoEvents > 0 ? 'text-warning' : ''}
                            label="no event"
                        />
                    </div>
                )}
                <ConnectionContainer>
                    {error && <ConnectionError errors={[error.message]} />}
                    {loading && !connection && <ConnectionLoading />}
                    {connection && connection.nodes?.length > 0 && <Header />}
                    <ConnectionList as="ul" className="list-group" aria-label="Webhooks">
                        {connection?.nodes?.map(node => (
                            <WebhookNode
                                key={node.id}
                                name={node.name}
                                id={node.id}
                                codeHostKind={node.codeHostKind}
                                codeHostURN={node.codeHostURN}
                            />
                        ))}
                    </ConnectionList>
                    {connection && (
                        <SummaryContainer className="mt-2" centered={true}>
                            <ConnectionSummary
                                noSummaryIfAllNodesVisible={false}
                                first={connection.totalCount ?? 0}
                                centered={true}
                                connection={connection}
                                noun="webhook"
                                pluralNoun="webhooks"
                                hasNextPage={hasNextPage}
                                emptyElement={<EmptyList />}
                            />
                            {hasNextPage && <ShowMoreButton centered={true} onClick={fetchMore} />}
                        </SummaryContainer>
                    )}
                </ConnectionContainer>
            </Container>
        </div>
    )
}

const Header: React.FunctionComponent<React.PropsWithChildren<{}>> = () => (
    <H5 className="p-2 d-none d-md-block text-uppercase text-left text-nowrap">Webhook</H5>
)

const EmptyList: React.FunctionComponent<React.PropsWithChildren<{}>> = () => (
    <div className="text-muted text-center mb-3 w-100">
        <Icon className="icon" svgPath={mdiMapSearch} inline={false} aria-hidden={true} />
        <div className="pt-2">No webhooks have been created so far.</div>
    </div>
)
