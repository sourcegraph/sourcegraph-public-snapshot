import { FC, useCallback, useEffect, useMemo, useState } from 'react'

import { mdiCog, mdiDelete, mdiOpenInNew, mdiPlus } from '@mdi/js'
import classNames from 'classnames'
import { useNavigate, useParams } from 'react-router-dom'
import { DeleteGitHubAppResult, DeleteGitHubAppVariables } from 'src/graphql-operations'

import { Timestamp } from '@sourcegraph/branded/src/components/Timestamp'
import { ErrorLike } from '@sourcegraph/common'
import { useMutation, useQuery } from '@sourcegraph/http-client'
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
    H3,
    Link,
    Text,
    Tooltip,
    Grid,
    AnchorLink,
} from '@sourcegraph/wildcard'

import { GitHubAppByIDResult, GitHubAppByIDVariables } from '../../graphql-operations'
import { ExternalServiceNode } from '../externalServices/ExternalServiceNode'
import { ConnectionList, SummaryContainer, ConnectionSummary } from '../FilteredConnection/ui'
import { PageTitle } from '../PageTitle'

import { AuthProviderMessage } from './AuthProviderMessage'
import { GITHUB_APP_BY_ID_QUERY, DELETE_GITHUB_APP_BY_ID_QUERY } from './backend'

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

    const [deleteGitHubApp, { loading: deleteLoading }] = useMutation<DeleteGitHubAppResult, DeleteGitHubAppVariables>(
        DELETE_GITHUB_APP_BY_ID_QUERY
    )

    const { data, loading, error } = useQuery<GitHubAppByIDResult, GitHubAppByIDVariables>(GITHUB_APP_BY_ID_QUERY, {
        variables: { id: appID ?? '' },
    })

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
            window.open(
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
                <>
                    <PageHeader
                        path={[
                            { icon: mdiCog },
                            { to: '/site-admin/github-apps', text: 'GitHub Apps' },
                            {
                                text: (
                                    <span className="d-flex align-items-center">
                                        <img
                                            className={classNames(styles.logo, 'mr-2')}
                                            src={app.logo}
                                            alt="App logo"
                                        />
                                        <span>{app.name}</span>
                                    </span>
                                ),
                            },
                        ]}
                        className="mb-3"
                        headingElement="h2"
                    />
                    <div className="d-flex align-items-center">
                        <span className="timestamps text-muted">
                            Created <Timestamp date={app.createdAt} /> | Updated <Timestamp date={app.updatedAt} />
                        </span>
                        <span className="ml-auto">
                            <AnchorLink to={app.appURL} target="_blank" className="mr-3">
                                View In GitHub <Icon inline={true} svgPath={mdiOpenInNew} aria-hidden={true} />
                            </AnchorLink>
                            <Button onClick={() => navigate(-1)} variant="secondary">
                                Cancel
                            </Button>
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
                        </span>
                    </div>
                    <Container className="mt-3 mb-3">
                        <Grid columnCount={2} templateColumns="auto 1fr" spacing={[0.6, 2]}>
                            <span className="font-weight-bold">GitHub App Name</span>
                            <span>{app.name}</span>
                            <span className="font-weight-bold">URL</span>
                            <AnchorLink to={app.appURL} target="_blank" className="text-decoration-none">
                                {app.appURL}
                            </AnchorLink>
                            <span className="font-weight-bold">AppID</span>
                            <span>{app.appID}</span>
                        </Grid>
                        <AuthProviderMessage app={app} id={appID} />

                        <hr className="mt-4 mb-4" />

                        <div>
                            <H2 className="d-flex align-items-center">
                                Installations
                                <Button className="ml-auto" onClick={() => onAddInstallation(app)} variant="primary">
                                    <Icon svgPath={mdiPlus} aria-hidden={true} /> Add installation
                                </Button>
                            </H2>
                            <div className="list-group mb-3" aria-label="GitHub App Installations">
                                {app.installations?.length === 0 ? (
                                    <Text>
                                        This GitHub App does not have any installations. Install the App to create a new
                                        connection.
                                    </Text>
                                ) : (
                                    app.installations?.map(installation => (
                                        <Container
                                            className={classNames(styles.installation, 'p-3')}
                                            key={installation.id}
                                        >
                                            <div className="d-flex align-items-center">
                                                <Link
                                                    to={installation.account.url}
                                                    className="d-flex align-items-center"
                                                >
                                                    <img
                                                        className={styles.logo}
                                                        src={installation.account.avatarURL}
                                                        alt="account avatar"
                                                    />
                                                    <div className="d-flex flex-column ml-3">
                                                        {installation.account.login}
                                                        <span className="text-muted">
                                                            ID: {installation.id} | Type: {installation.account.type}
                                                        </span>
                                                    </div>
                                                </Link>
                                                <AnchorLink to={installation.url} target="_blank" className="ml-auto">
                                                    <small>
                                                        View In GitHub{' '}
                                                        <Icon inline={true} svgPath={mdiOpenInNew} aria-hidden={true} />
                                                    </small>
                                                </AnchorLink>
                                            </div>
                                            <div className="mt-4">
                                                <H3 className="d-flex align-items-center mb-0">
                                                    Code host connections
                                                    <ButtonLink
                                                        variant="primary"
                                                        className="ml-auto"
                                                        to={`/site-admin/external-services/new?id=github&appID=${
                                                            app.appID
                                                        }&installationID=${installation.id}&url=${encodeURI(
                                                            app.baseURL
                                                        )}&org=${installation.account.login}`}
                                                        size="sm"
                                                    >
                                                        <Icon svgPath={mdiPlus} aria-hidden={true} /> Add connection
                                                    </ButtonLink>
                                                </H3>
                                                {installation.externalServices?.nodes?.length > 0 ? (
                                                    <>
                                                        <ConnectionList
                                                            as="ul"
                                                            className={styles.listGroup}
                                                            aria-label="Code Host Connections"
                                                        >
                                                            {installation.externalServices?.nodes?.map(node => (
                                                                <ExternalServiceNode
                                                                    key={node.id}
                                                                    node={node}
                                                                    editingDisabled={false}
                                                                />
                                                            ))}
                                                        </ConnectionList>
                                                        {installation.externalServices && (
                                                            <SummaryContainer className="mt-2" centered={true}>
                                                                <ConnectionSummary
                                                                    noSummaryIfAllNodesVisible={false}
                                                                    first={100}
                                                                    centered={true}
                                                                    connection={installation.externalServices}
                                                                    noun="code host connection"
                                                                    pluralNoun="code host connections"
                                                                    hasNextPage={false}
                                                                />
                                                            </SummaryContainer>
                                                        )}
                                                    </>
                                                ) : (
                                                    <Text className="text-center mt-4">
                                                        You haven't added any code host connections yet.
                                                    </Text>
                                                )}
                                            </div>
                                        </Container>
                                    ))
                                )}
                                <SummaryContainer className="mt-3" centered={true}>
                                    <ConnectionSummary
                                        noSummaryIfAllNodesVisible={false}
                                        first={app?.installations?.length ?? 0}
                                        centered={true}
                                        connection={{
                                            nodes: app?.installations ?? [],
                                            totalCount: app?.installations?.length ?? 0,
                                        }}
                                        noun="installation"
                                        pluralNoun="installations"
                                        hasNextPage={false}
                                    />
                                </SummaryContainer>
                            </div>
                        </div>
                    </Container>
                </>
            )}
        </div>
    )
}
