import { type FC, useEffect, useState } from 'react'

import { mdiWebhook, mdiDelete, mdiPencil } from '@mdi/js'
import { noop } from 'lodash'
import { useNavigate, useParams } from 'react-router-dom'

import { useMutation } from '@sourcegraph/http-client'
import type { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import {
    Button,
    ButtonLink,
    Container,
    H2,
    H5,
    Link,
    LoadingSpinner,
    PageHeader,
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
import type { DeleteWebhookResult, DeleteWebhookVariables, WebhookFields } from '../graphql-operations'

import { DELETE_WEBHOOK, useWebhookLogsConnection, useWebhookQuery } from './backend'
import { WebhookInfoLogPageHeader } from './WebhookInfoLogPageHeader'
import { WebhookInformation } from './WebhookInformation'
import { WebhookLogNode } from './webhooks/WebhookLogNode'

import styles from './SiteAdminWebhookPage.module.scss'

export interface WebhookPageProps extends TelemetryProps {}

export const SiteAdminWebhookPage: FC<WebhookPageProps> = props => {
    const { telemetryService, telemetryRecorder } = props

    const { id = '' } = useParams<{ id: string }>()
    const navigate = useNavigate()

    const [onlyErrors, setOnlyErrors] = useState(false)
    const { loading, hasNextPage, fetchMore, connection, error } = useWebhookLogsConnection(id, 20, onlyErrors)
    const { loading: webhookLoading, data: webhookData } = useWebhookQuery(id)

    useEffect(() => {
        telemetryService.logPageView('SiteAdminWebhook')
        telemetryRecorder.recordEvent('siteAdminWebhook', 'viewed')
    }, [telemetryService, telemetryRecorder])

    const [deleteWebhook, { error: deleteError, loading: isDeleting }] = useMutation<
        DeleteWebhookResult,
        DeleteWebhookVariables
    >(DELETE_WEBHOOK, { variables: { hookID: id }, onCompleted: () => navigate('/site-admin/webhooks/incoming') })

    return (
        <Container>
            <PageTitle title="Incoming webhooks" />
            {webhookLoading && !webhookData && <ConnectionLoading />}
            {webhookData?.node && webhookData.node.__typename === 'Webhook' && (
                <PageHeader
                    path={[
                        { icon: mdiWebhook },
                        { to: '/site-admin/webhooks/incoming', text: 'Incoming webhooks' },
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
                        <div className="d-flex flex-row align-items-center">
                            <ButtonLink
                                to={`/site-admin/webhooks/incoming/${id}/edit`}
                                className="test-edit-webhook mr-2"
                                variant="secondary"
                                display="inline"
                            >
                                <Icon aria-hidden={true} svgPath={mdiPencil} />
                                {' Edit'}
                            </ButtonLink>
                            <Button
                                aria-label="Delete"
                                className="test-delete-webhook"
                                variant="danger"
                                disabled={isDeleting}
                                onClick={event => {
                                    event.preventDefault()
                                    if (
                                        !window.confirm(
                                            'Delete this webhook? Any external webhooks configured to point at this endpoint will no longer be received.'
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
                                {isDeleting && <LoadingSpinner />}
                                <Icon aria-hidden={true} svgPath={mdiDelete} />
                                {' Delete'}
                            </Button>
                        </div>
                    }
                />
            )}

            {deleteError && <ErrorAlert className="mt-2" prefix="Error during webhook deletion" error={deleteError} />}

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
                    <SiteAdminWebhookPageHeader timeLabel="Received at" />
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
                            emptyElement={<EmptyList onlyErrors={onlyErrors} />}
                        />
                        {hasNextPage && <ShowMoreButton centered={true} onClick={fetchMore} />}
                    </SummaryContainer>
                )}
            </ConnectionContainer>
        </Container>
    )
}

export interface SiteAdminWebhookPageHeaderProps {
    middleColumnLabel?: string
    timeLabel: string
}

export const SiteAdminWebhookPageHeader: FC<SiteAdminWebhookPageHeaderProps> = ({ middleColumnLabel, timeLabel }) => (
    <>
        {/* Render an empty element here to fill in available space for the first column*/}
        {/* element in the header row*/}
        <span className="d-md-block" />
        <H5 className="text-uppercase text-center text-nowrap">Status code</H5>
        <H5 className="text-uppercase text-nowrap">{middleColumnLabel}</H5>
        <H5 className="text-uppercase text-nowrap">{timeLabel}</H5>
    </>
)

const EmptyList: FC<{ onlyErrors: boolean }> = ({ onlyErrors }) => (
    <div className="m-4 w-100 text-center text-muted">
        {onlyErrors ? (
            'No errors have been received from this webhook recently.'
        ) : (
            <>
                No requests received yet. Be sure to{' '}
                <Link to="/help/admin/config/webhooks/incoming#configuring-webhooks-on-the-code-host">
                    configure the webhook on the code host
                </Link>
                .
            </>
        )}
    </div>
)
