import { useCallback, useEffect, useState, type FC } from 'react'

import { mdiDelete, mdiPencil, mdiWebhook } from '@mdi/js'
import { useNavigate, useParams } from 'react-router-dom'

import { type TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import type { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import {
    Alert,
    Button,
    ButtonLink,
    Code,
    Container,
    ErrorAlert,
    H2,
    H5,
    Icon,
    Link,
    PageHeader,
    Text,
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
import { ExternalServiceKind, type WebhookFields } from '../graphql-operations'

import { useWebhookLogsConnection, useWebhookQuery } from './backend'
import { WebhookConfirmDeleteModal } from './WebhookConfirmDeleteModal'
import { WebhookInfoLogPageHeader } from './WebhookInfoLogPageHeader'
import { WebhookInformation } from './WebhookInformation'
import { WebhookLogNode } from './webhooks/WebhookLogNode'

import styles from './SiteAdminWebhookPage.module.scss'

export interface WebhookPageProps extends TelemetryProps, TelemetryV2Props {}

export const SiteAdminWebhookPage: FC<WebhookPageProps> = props => {
    const { telemetryService, telemetryRecorder } = props

    const { id = '' } = useParams<{ id: string }>()
    const navigate = useNavigate()

    const [onlyErrors, setOnlyErrors] = useState(false)
    const {
        loading,
        hasNextPage,
        fetchMore,
        connection,
        error: webhookLogsError,
    } = useWebhookLogsConnection(id, onlyErrors)
    const { loading: webhookLoading, data: webhookData, error: webhookError } = useWebhookQuery(id)

    useEffect(() => {
        telemetryService.logPageView('SiteAdminWebhook')
        telemetryRecorder.recordEvent('admin.webhook', 'view')
    }, [telemetryService, telemetryRecorder])

    const [showDeleteModal, setShowDeleteModal] = useState(false)
    const deleteWebhook = useCallback(() => {
        setShowDeleteModal(true)
    }, [])

    return (
        <>
            <PageTitle title="Incoming webhooks" />
            {webhookLoading && !webhookData && <ConnectionLoading />}
            {!webhookLoading && !webhookData && webhookError && <ErrorAlert error={webhookError} />}
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
                                <Icon aria-hidden={true} svgPath={mdiPencil} /> Edit
                            </ButtonLink>
                            <Button
                                aria-label="Delete"
                                className="test-delete-webhook"
                                variant="danger"
                                disabled={showDeleteModal}
                                onClick={deleteWebhook}
                            >
                                <Icon aria-hidden={true} svgPath={mdiDelete} /> Delete
                            </Button>
                        </div>
                    }
                />
            )}
            <Container className="mb-3">
                <H2>Information</H2>
                {webhookData?.node && webhookData.node.__typename === 'Webhook' && (
                    <WebhookInformation webhook={webhookData.node as WebhookFields} />
                )}

                <H2>Logs</H2>
                <WebhookInfoLogPageHeader webhookID={id} onlyErrors={onlyErrors} onSetOnlyErrors={setOnlyErrors} />

                <ConnectionContainer className="mt-5">
                    {webhookLogsError && <ConnectionError errors={[webhookLogsError.message]} />}
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

            {webhookData?.node && webhookData.node.__typename === 'Webhook' && (
                <>
                    <H2>Setup instructions</H2>
                    <Container>
                        <WebhookSetupInstructions webhook={webhookData.node} />
                    </Container>
                </>
            )}

            {showDeleteModal && webhookData?.node && webhookData.node.__typename === 'Webhook' && (
                <WebhookConfirmDeleteModal
                    webhook={webhookData.node}
                    onCancel={() => setShowDeleteModal(false)}
                    afterDelete={() => navigate('/site-admin/webhooks/incoming')}
                    telemetryRecorder={telemetryRecorder}
                />
            )}
        </>
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

interface WebhookSetupInstructionsProps {
    webhook: WebhookFields
}

const WebhookSetupInstructions: React.FunctionComponent<WebhookSetupInstructionsProps> = ({ webhook }) => {
    if (webhook.codeHostKind === ExternalServiceKind.GITHUB) {
        return (
            <>
                <Text>
                    To set up a GitHub webhook, follow the instructions below, or see more in the{' '}
                    <Link to="/help/admin/config/webhooks/incoming#github">GitHub webooks documentation</Link>.
                </Text>
                <Alert variant="info">
                    Note: For GitHub App integrations, webhooks are created automatically. You do not need to create
                    them manually.
                </Alert>
                <Text className="mb-0">
                    <ol className="mb-0">
                        <li>
                            Copy the webhook URL <strong>{webhook.url}</strong>
                        </li>
                        <li>
                            On GitHub, go to the settings page of your organization. From there, click{' '}
                            <strong>Settings</strong>, then
                            <strong>Webhooks</strong>, then <strong>Add webhook</strong>.
                        </li>
                        <li>
                            Fill in the webhook form:
                            <ul>
                                <li>Payload URL: the URL you just copied above.</li>
                                <li>
                                    Content type: this must be set to <strong>application/json</strong>.
                                </li>
                                <li>Secret: the secret token you can find above.</li>
                                <li>Active: ensure this is enabled.</li>
                                <li>
                                    Which events: select <strong>Let me select individual events</strong>, and then
                                    enable:
                                    <table className="table ml-3">
                                        <thead>
                                            <tr>
                                                <th className="px-2">Repo updates</th>
                                                <th className="px-2">Batch Changes</th>
                                                <th className="px-2">Repo permissions</th>
                                            </tr>
                                        </thead>
                                        <tbody>
                                            <tr>
                                                <td>
                                                    <ul>
                                                        <li>
                                                            <Code>push</Code>
                                                        </li>
                                                    </ul>
                                                </td>
                                                <td>
                                                    <ul>
                                                        <li>
                                                            <Code>Issue comments</Code>
                                                        </li>
                                                        <li>
                                                            <Code>Pull requests</Code>
                                                        </li>
                                                        <li>
                                                            <Code>Pull request reviews</Code>
                                                        </li>
                                                        <li>
                                                            <Code>Pull request review comments</Code>
                                                        </li>
                                                        <li>
                                                            <Code>Check runs</Code>
                                                        </li>
                                                        <li>
                                                            <Code>Check suites</Code>
                                                        </li>
                                                        <li>
                                                            <Code>Statuses</Code>
                                                        </li>
                                                    </ul>
                                                </td>
                                                <td>
                                                    <ul>
                                                        <li>
                                                            <Code>Collaborator add, remove, or changed</Code>
                                                        </li>
                                                        <li>
                                                            <Code>Memberships</Code>
                                                        </li>
                                                        <li>
                                                            <Code>Organizations</Code>
                                                        </li>
                                                        <li>
                                                            <Code>Repositories</Code>
                                                        </li>
                                                        <li>
                                                            <Code>Teams</Code>
                                                        </li>
                                                    </ul>
                                                </td>
                                            </tr>
                                        </tbody>
                                    </table>
                                </li>
                            </ul>
                        </li>
                        <li>
                            Click <strong>Add webhook</strong>.
                        </li>
                        <li>Confirm that the new webhook is listed.</li>
                        <li>You should see an initial ping event sent from GitHub in the webhook logs above.</li>
                    </ol>
                </Text>
            </>
        )
    }
    if (webhook.codeHostKind === ExternalServiceKind.GITLAB) {
        return (
            <>
                <Text className="mb-0">
                    To set up a GitLab webhook, follow the instructions in the{' '}
                    <Link to="/help/admin/config/webhooks/incoming#gitlab">GitLab integration documentation</Link>.
                </Text>
            </>
        )
    }
    if (webhook.codeHostKind === ExternalServiceKind.BITBUCKETSERVER) {
        return (
            <>
                <Text className="mb-0">
                    To set up a Bitbucket Server webhook, follow the instructions in the{' '}
                    <Link to="/help/admin/config/webhooks/incoming#bitbucket-server">
                        Bitbucket Server integration documentation
                    </Link>
                    .
                </Text>
            </>
        )
    }
    if (webhook.codeHostKind === ExternalServiceKind.BITBUCKETCLOUD) {
        return (
            <>
                <Text className="mb-0">
                    To set up a Bitbucket Cloud webhook, follow the instructions in the{' '}
                    <Link to="/help/admin/config/webhooks/incoming#bitbucket-cloud">
                        Bitbucket Cloud integration documentation
                    </Link>
                    .
                </Text>
            </>
        )
    }
    if (webhook.codeHostKind === ExternalServiceKind.AZUREDEVOPS) {
        return (
            <>
                <Text className="mb-0">
                    To set up an Azure DevOps webhook, follow the instructions in the{' '}
                    <Link to="/help/admin/config/webhooks/incoming#azure-devops">
                        Azure DevOps integration documentation
                    </Link>
                    .
                </Text>
            </>
        )
    }
    return null
}
