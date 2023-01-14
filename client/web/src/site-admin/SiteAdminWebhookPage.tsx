import { FC, useEffect, useState } from 'react'

import { mdiCog, mdiDelete } from '@mdi/js'
import { noop } from 'lodash'
import { RouteComponentProps } from 'react-router'

import { useMutation } from '@sourcegraph/http-client'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import {
    Button,
    ButtonLink,
    Container,
    H2,
    H5,
    Link,
    LoadingSpinner,
    PageHeader,
    Tooltip,
    ErrorAlert,
    Icon,
} from '@sourcegraph/wildcard'

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
import { DeleteWebhookResult, DeleteWebhookVariables, WebhookFields } from '../graphql-operations'

import { DELETE_WEBHOOK, useWebhookLogsConnection, useWebhookQuery } from './backend'
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
        history,
    } = props

    const [onlyErrors, setOnlyErrors] = useState(false)
    const { loading, hasNextPage, fetchMore, connection, error } = useWebhookLogsConnection(id, 20, onlyErrors)
    const { loading: webhookLoading, data: webhookData } = useWebhookQuery(id)

    useEffect(() => {
        telemetryService.logPageView('SiteAdminWebhook')
    }, [telemetryService])

    const [deleteWebhook, { error: deleteError, loading: isDeleting }] = useMutation<
        DeleteWebhookResult,
        DeleteWebhookVariables
    >(DELETE_WEBHOOK, { variables: { hookID: id }, onCompleted: () => history.push('/site-admin/webhooks') })

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
                                <Tooltip content="Edit webhook">
                                    <ButtonLink
                                        to={`/site-admin/webhooks/${id}/edit`}
                                        className="test-edit-webhook"
                                        size="sm"
                                        variant="primary"
                                        display="inline"
                                    >
                                        <Icon aria-hidden={true} svgPath={mdiCog} />
                                        {' Edit'}
                                    </ButtonLink>
                                </Tooltip>
                            </div>
                            <div className="mr-1">
                                <Tooltip content="Delete webhook">
                                    <Button
                                        aria-label="Delete"
                                        className="test-delete-webhook"
                                        variant="danger"
                                        size="sm"
                                        disabled={isDeleting}
                                        onClick={event => {
                                            event.preventDefault()
                                            if (
                                                !window.confirm(
                                                    'Delete this webhook? Any external webhooks configured to point at this webhook will no longer be received.'
                                                )
                                            ) {
                                                return
                                            }
                                            deleteWebhook().catch(
                                                // noop here is used because creation error is handled directly when useMutation is called
                                                noop
                                            )
                                        }}
                                    >
                                        <>
                                            {isDeleting && <LoadingSpinner />}
                                            <Icon aria-hidden={true} svgPath={mdiDelete} />
                                            {' Delete'}
                                        </>
                                    </Button>
                                </Tooltip>
                                {deleteError && (
                                    <ErrorAlert
                                        className="mt-2"
                                        prefix="Error during webhook deletion"
                                        error={deleteError}
                                    />
                                )}
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
                        <WebhookLogNode doNotShowExternalService={true} key={node.id} node={node} />
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
        {/* Another empty element to fill in the empty middle of the grid */}
        <span className="d-md-block" />
        <H5 className="text-uppercase text-nowrap">Received at</H5>
    </>
)

const EmptyList: FC = () => (
    <div className="m-4 w-100 text-center text-muted">
        No requests received yet. Be sure to{' '}
        <Link to="/help/admin/config/webhooks#configuring-webhooks-on-the-code-host">
            configure the webhook on the code host
        </Link>
        .
    </div>
)
