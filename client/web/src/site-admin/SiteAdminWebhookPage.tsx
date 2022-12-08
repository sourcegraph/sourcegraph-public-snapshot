import { FC, useEffect, useState } from 'react'

import { mdiCog, mdiPlus } from '@mdi/js'
import { RouteComponentProps } from 'react-router'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ButtonLink, Container, H2, H5, Icon, PageHeader } from '@sourcegraph/wildcard'

import { CreatedByAndUpdatedByInfoByline } from '../components/Byline/CreatedByAndUpdatedByInfoByline'
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
import { WebhookFields } from '../graphql-operations'

import { useWebhookLogsConnection, useWebhookQuery } from './backend'
import { WebhookInfoLogPageHeader } from './WebhookInfoLogPageHeader'
import { WebhookInformation } from './WebhookInformation'
import { WebhookLogNode } from './webhooks/WebhookLogNode'

import styles from './SiteAdminWebhookPage.module.scss'

export interface WebhookPageProps extends TelemetryProps, RouteComponentProps<{ id: string }> {}

export const SiteAdminWebhookPage: FC<WebhookPageProps> = props => {
    const {
        match: {
            params: { id },
        },
        telemetryService,
    } = props

    const [onlyErrors, setOnlyErrors] = useState(false)
    const { loading, hasNextPage, fetchMore, connection, error } = useWebhookLogsConnection(id, 20, onlyErrors)
    const { loading: webhookLoading, data: webhookData } = useWebhookQuery(id)

    useEffect(() => {
        telemetryService.logPageView('SiteAdminWebhook')
    }, [telemetryService])

    return (
        <Container>
            <PageTitle title="Incoming webhook" />
            {webhookLoading && !webhookData && <ConnectionLoading />}
            {webhookData?.node && webhookData.node.__typename === 'Webhook' && (
                <PageHeader
                    path={[
                        { icon: mdiCog },
                        { to: '/site-admin/webhooks', text: 'Incoming webhooks' },
                        { text: webhookData.node.name },
                    ]}
                    byline={
                        <CreatedByAndUpdatedByInfoByline
                            createdAt={webhookData.node.createdAt}
                            createdBy={webhookData.node.createdBy}
                            updatedAt={webhookData.node.updatedAt}
                            updatedBy={webhookData.node.updatedBy}
                        />
                    }
                    className="mb-3"
                    headingElement="h2"
                    actions={
                        <div className="d-flex flex-row-reverse align-items-center">
                            <div className="flex-grow mr-2">
                                <ButtonLink
                                    to="/site-admin/webhooks/create"
                                    className="test-create-webhook"
                                    size="sm"
                                    variant="primary"
                                    display="inline"
                                >
                                    <Icon aria-hidden={true} svgPath={mdiPlus} /> Create
                                </ButtonLink>
                            </div>
                            <div className="mr-1">
                                <ButtonLink
                                    className="test-edit-webhook"
                                    size="sm"
                                    variant="secondary"
                                    display="inline"
                                >
                                    Edit
                                </ButtonLink>
                            </div>
                        </div>
                    }
                />
            )}

            <H2>Information</H2>
            {webhookData?.node && webhookData.node.__typename === 'Webhook' && (
                <WebhookInformation webhook={webhookData.node as WebhookFields} />
            )}

            <H2>Logs</H2>
            <WebhookInfoLogPageHeader webhookID={id} onlyErrors={onlyErrors} onSetOnlyErrors={setOnlyErrors} />

            <ConnectionContainer className="mt-5">
                {error && <ConnectionError errors={[error.message]} />}
                {loading && !connection && <ConnectionLoading />}

                <ConnectionList aria-label="WebhookLogs" className={styles.logs}>
                    <Header />
                    {connection?.nodes?.map(node => (
                        <WebhookLogNode key={node.id} node={node} />
                    ))}
                </ConnectionList>

                {connection && (
                    <SummaryContainer className="mt-2">
                        <ConnectionSummary
                            noSummaryIfAllNodesVisible={false}
                            first={connection.totalCount ?? 0}
                            centered={true}
                            connection={connection}
                            noun="webhook log"
                            pluralNoun="webhook logs"
                            hasNextPage={hasNextPage}
                            emptyElement={<EmptyList />}
                        />
                        {hasNextPage && <ShowMoreButton centered={true} onClick={fetchMore} />}
                    </SummaryContainer>
                )}
            </ConnectionContainer>
        </Container>
    )
}

const Header: FC = () => (
    <>
        {/* Render an empty element here to fill in available space for the first column*/}
        {/* element in the header row*/}
        <span className="d-md-block" />
        <H5 className="text-uppercase text-center text-nowrap">Status code</H5>
        <H5 className="text-uppercase text-nowrap">External service</H5>
        <H5 className="text-uppercase text-nowrap">Received at</H5>
    </>
)

const EmptyList: FC = () => <div className="m-4 w-100 text-center">No webhook logs found</div>
