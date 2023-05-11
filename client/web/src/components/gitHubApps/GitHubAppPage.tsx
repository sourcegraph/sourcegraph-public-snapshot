import { FC, useCallback, useEffect, useMemo, useState } from 'react'

import { mdiCog, mdiDelete, mdiGithub, mdiPlus } from '@mdi/js'
import classNames from 'classnames'
import { useNavigate, useParams } from 'react-router-dom'
import {
    ConnectWebhookToGitHubAppResult,
    ConnectWebhookToGitHubAppVariables,
    DeleteGitHubAppResult,
    DeleteGitHubAppVariables,
    WebhooksListResult,
    WebhooksListVariables,
} from 'src/graphql-operations'

import { Timestamp } from '@sourcegraph/branded/src/components/Timestamp'
import { ErrorLike } from '@sourcegraph/common'
import { useMutation, useQuery } from '@sourcegraph/http-client'
import { UserAvatar } from '@sourcegraph/shared/src/components/UserAvatar'
import { ExternalServiceKind } from '@sourcegraph/shared/src/graphql-operations'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import {
    Container,
    ErrorAlert,
    PageHeader,
    ButtonLink,
    Icon,
    LoadingSpinner,
    Button,
    H2,
    Card,
    Link,
    Text,
    Tooltip,
    H3,
    Select,
} from '@sourcegraph/wildcard'

import { GitHubAppByIDResult, GitHubAppByIDVariables } from '../../graphql-operations'
import { WEBHOOKS } from '../../site-admin/backend'
import { ExternalServiceNode } from '../externalServices/ExternalServiceNode'
import { ConnectionList, SummaryContainer, ConnectionSummary } from '../FilteredConnection/ui'
import { PageTitle } from '../PageTitle'

import { AuthProviderMessage } from './AuthProviderMessage'
import { GITHUB_APP_BY_ID_QUERY, DELETE_GITHUB_APP_BY_ID_QUERY, CONNECT_WEBHOOK_TO_GITHUB_APP_QUERY } from './backend'

import styles from './GitHubAppCard.module.scss'

interface Props extends TelemetryProps {
    externalServicesFromFile: boolean
    allowEditExternalServicesWithFile: boolean
}

export const GitHubAppPage: FC<Props> = ({
    telemetryService,
    externalServicesFromFile,
    allowEditExternalServicesWithFile,
}) => {
    const { appID } = useParams()
    const navigate = useNavigate()

    useEffect(() => {
        telemetryService.logPageView('SiteAdminGitHubApp')
    }, [telemetryService])
    const [fetchError, setError] = useState<ErrorLike>()
    const [selectedWebhook, setSelectedWebhook] = useState<string>('')

    const [deleteGitHubApp, { loading: deleteLoading }] = useMutation<DeleteGitHubAppResult, DeleteGitHubAppVariables>(
        DELETE_GITHUB_APP_BY_ID_QUERY
    )

    const [connectWebhook, { loading: connectWebhookLoading }] = useMutation<
        ConnectWebhookToGitHubAppResult,
        ConnectWebhookToGitHubAppVariables
    >(CONNECT_WEBHOOK_TO_GITHUB_APP_QUERY)

    const { data, loading, error, refetch } = useQuery<GitHubAppByIDResult, GitHubAppByIDVariables>(
        GITHUB_APP_BY_ID_QUERY,
        {
            variables: { id: appID ?? '' },
        }
    )

    const { data: webhooks, loading: webhooksLoading } = useQuery<WebhooksListResult, WebhooksListVariables>(
        WEBHOOKS,
        {}
    )

    const app = useMemo(() => data?.gitHubApp, [data])

    const onDelete = useCallback<React.MouseEventHandler>(async () => {
        if (!window.confirm(`Delete the GitHub App ${app?.name}?`)) {
            return
        }
        try {
            await deleteGitHubApp({
                variables: { gitHubApp: app?.id ?? '' },
            })
        } finally {
            navigate('/site-admin/github-apps')
        }
    }, [app, deleteGitHubApp, navigate])

    const handleWebhookChange = useCallback(
        (event: React.ChangeEvent<HTMLSelectElement>) => {
            setSelectedWebhook(event.target.value)
        },
        [setSelectedWebhook]
    )

    if (!appID) {
        return null
    }

    const handleError = (error: ErrorLike): [] => {
        setError(error)
        return []
    }

    const onAddInstallation = async (app: NonNullable<GitHubAppByIDResult['gitHubApp']>): Promise<void> => {
        try {
            const req = await fetch(`/.auth/githubapp/state?id=${app?.id}`)
            const state = await req.text()
            window.location.assign(
                app.appURL.endsWith('/')
                    ? app.appURL + 'installations/new?state=' + state
                    : app.appURL + '/installations/new?state=' + state
            )
        } catch (error) {
            handleError(error)
        }
    }

    return (
        <div>
            {app ? <PageTitle title={`GitHub App - ${app.name}`} /> : <PageTitle title="GitHub App" />}
            {(error || fetchError) && <ErrorAlert className="mb-3" error={error ?? fetchError} />}
            {loading && <LoadingSpinner />}
            {app && (
                <Container className="mb-3">
                    <PageHeader
                        path={[
                            { icon: mdiCog },
                            { to: '/site-admin/github-apps', text: 'GitHub Apps' },
                            {
                                to: `/site-admin/github-apps/${appID}`,
                                text: app.name,
                            },
                        ]}
                        className="mb-3"
                        headingElement="h2"
                        actions={
                            <>
                                <ButtonLink to={app.appURL} variant="info" className="ml-2">
                                    <Icon inline={true} svgPath={mdiGithub} aria-hidden={true} /> Edit
                                </ButtonLink>
                                <Tooltip content="Delete GitHub App">
                                    <Button
                                        aria-label="Delete"
                                        className="ml-2"
                                        onClick={onDelete}
                                        disabled={deleteLoading}
                                        variant="danger"
                                    >
                                        <Icon aria-hidden={true} svgPath={mdiDelete} />
                                        {' Delete'}
                                    </Button>
                                </Tooltip>
                            </>
                        }
                    />
                    <span className="d-flex align-items-center mt-2 mb-3">
                        <img className={classNames(styles.logo, 'mr-4')} src={app.logo} alt="App logo" />
                        <div className="d-flex flex-column">
                            <small className="text-muted">AppID: {app.appID}</small>
                            <small className="text-muted">Slug: {app.slug}</small>
                            <small className="text-muted">ClientID: {app.clientID}</small>
                        </div>
                        <span className="ml-auto">
                            <span>
                                Created <Timestamp date={app.createdAt} />
                            </span>
                            <span className="ml-3">
                                Updated <Timestamp date={app.updatedAt} />
                            </span>
                        </span>
                    </span>
                    <AuthProviderMessage app={app} id={appID} />
                    <H3>Webhooks</H3>
                    {app.webhook ? (
                        <Link to={`/site-admin/webhooks/incoming/${app.webhook.id}`}>View webhook logs</Link>
                    ) : (
                        <>
                            <Text>
                                This GitHub App does not have Webhooks set up. Connect a webhook from the list below, or{' '}
                                <Link to="/site-admin/webhooks/incoming">create a new webhook</Link>.
                            </Text>
                            {webhooksLoading && <LoadingSpinner />}
                            {!webhooksLoading && webhooks && webhooks.webhooks.nodes.length > 0 ? (
                                <span className="d-flex mb-3">
                                    <Select
                                        id="webhookSelector"
                                        aria-label="Available webhooks"
                                        value={selectedWebhook}
                                        onChange={handleWebhookChange}
                                    >
                                        <option value="">Select a webhook</option>
                                        {webhooks?.webhooks.nodes
                                            .filter(
                                                item =>
                                                    item.codeHostKind === ExternalServiceKind.GITHUB &&
                                                    item.codeHostURN === app.baseURL + '/'
                                            )
                                            .map(item => (
                                                <option key={item.id} value={item.id}>
                                                    {item.name}
                                                </option>
                                            ))}
                                    </Select>
                                    <div>
                                        <Button
                                            variant="primary"
                                            className="ml-2"
                                            size="sm"
                                            disabled={selectedWebhook === '' || connectWebhookLoading}
                                            onClick={async () => {
                                                await connectWebhook({
                                                    variables: { gitHubApp: app?.id ?? '', webhook: selectedWebhook },
                                                })
                                                await refetch({ id: appID ?? '' })
                                            }}
                                        >
                                            Connect webhook
                                        </Button>
                                    </div>
                                </span>
                            ) : null}
                        </>
                    )}
                    <hr />

                    <div className="mt-4">
                        <H2>App installations</H2>
                        <div className="list-group mb-3" aria-label="GitHub App Installations">
                            {app.installations?.length === 0 ? (
                                <Text>
                                    This GitHub App does not have any installations. Install the App to create a new
                                    connection.
                                </Text>
                            ) : (
                                app.installations?.map(installation => (
                                    <Card
                                        className={classNames(styles.listNode, 'd-flex flex-row align-items-center')}
                                        key={installation.id}
                                    >
                                        <span className="mr-3">
                                            <Link to={installation.account.url} className="mr-3">
                                                <UserAvatar
                                                    size={32}
                                                    user={{ ...installation.account, displayName: null }}
                                                    className="mr-2"
                                                />
                                                {installation.account.login}
                                            </Link>
                                            <span>Type: {installation.account.type}</span>
                                        </span>
                                        <small className="text-muted mr-3">ID: {installation.id}</small>
                                        <ButtonLink
                                            to={installation.url}
                                            variant="secondary"
                                            className="ml-auto mr-1"
                                            size="sm"
                                        >
                                            <Icon inline={true} svgPath={mdiGithub} aria-hidden={true} /> Edit
                                        </ButtonLink>
                                        <ButtonLink
                                            variant="success"
                                            to={`/site-admin/external-services/new?id=github&appID=${
                                                app.appID
                                            }&installationID=${installation.id}&url=${encodeURI(app.baseURL)}&org=${
                                                installation.account.login
                                            }`}
                                            size="sm"
                                        >
                                            <Icon svgPath={mdiPlus} aria-hidden={true} /> New connection
                                        </ButtonLink>
                                    </Card>
                                ))
                            )}
                        </div>
                        <Button
                            onClick={async () => {
                                await onAddInstallation(app)
                            }}
                            variant="success"
                        >
                            <Icon svgPath={mdiPlus} aria-hidden={true} /> Add installation
                        </Button>
                    </div>
                    <hr className="mt-4" />
                    <div className="mt-4">
                        <H2>Code host connections</H2>
                        <ConnectionList as="ul" className="list-group" aria-label="Code Host Connections">
                            {app.externalServices?.nodes?.map(node => (
                                <ExternalServiceNode key={node.id} node={node} editingDisabled={true} />
                            ))}
                        </ConnectionList>
                        {app.externalServices && (
                            <SummaryContainer className="mt-2" centered={true}>
                                <ConnectionSummary
                                    noSummaryIfAllNodesVisible={false}
                                    first={app.externalServices.totalCount ?? 0}
                                    centered={true}
                                    connection={app.externalServices}
                                    noun="code host connection"
                                    pluralNoun="code host connections"
                                    hasNextPage={false}
                                />
                            </SummaryContainer>
                        )}
                    </div>
                </Container>
            )}
        </div>
    )
}
