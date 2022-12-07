import React, { useEffect } from 'react'

import { mdiCog, mdiMapSearch, mdiPlus } from '@mdi/js'
import classNames from 'classnames'
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

import { useWebhooksConnection } from './backend'
import { WebhookNode } from './WebhookNode'

import styles from './SiteAdminWebhooksPage.module.scss'

interface Props extends RouteComponentProps<{}>, TelemetryProps {}

export const SiteAdminWebhooksPage: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    history,
    location,
    telemetryService,
}) => {
    useEffect(() => {
        telemetryService.logPageView('SiteAdminWebhooks')
    }, [telemetryService])

    const { loading, hasNextPage, fetchMore, connection, error } = useWebhooksConnection()
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
                <ConnectionContainer>
                    {error && <ConnectionError errors={[error.message]} />}
                    {loading && !connection && <ConnectionLoading />}
                    {connection && connection.nodes?.length > 0 && <Header />}
                    <ConnectionList
                        as="ul"
                        className={classNames(styles.webhooksGrid, 'list-group')}
                        aria-label="Webhooks"
                    >
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
                        <SummaryContainer className="mt-2">
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
    <div className={styles.webhooksGrid}>
        <H5 className="p-2 d-none d-md-block text-uppercase text-left text-nowrap">Webhook</H5>
        <H5 className="d-none d-md-block text-uppercase text-center text-nowrap">Actions</H5>
    </div>
)

const EmptyList: React.FunctionComponent<React.PropsWithChildren<{}>> = () => (
    <div className="text-muted text-center mb-3 w-100">
        <Icon className="icon" svgPath={mdiMapSearch} inline={false} aria-hidden={true} />
        <div className="pt-2">No webhooks have been created so far.</div>
    </div>
)
