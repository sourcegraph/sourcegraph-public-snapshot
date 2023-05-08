import { FC, useEffect, useMemo } from 'react'

import { mdiCog, mdiGithub, mdiRefresh, mdiPlus } from '@mdi/js'
import classNames from 'classnames'
import { useParams } from 'react-router-dom'

import { Timestamp } from '@sourcegraph/branded/src/components/Timestamp'
import { useQuery } from '@sourcegraph/http-client'
import { UserAvatar } from '@sourcegraph/shared/src/components/UserAvatar'
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
} from '@sourcegraph/wildcard'

import { GitHubAppByIDResult, GitHubAppByIDVariables } from '../../graphql-operations'
import { ExternalServiceNode } from '../externalServices/ExternalServiceNode'
import { ConnectionList, SummaryContainer, ConnectionSummary } from '../FilteredConnection/ui'
import { PageTitle } from '../PageTitle'

import { GITHUB_APP_BY_ID_QUERY } from './backend'

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

    useEffect(() => {
        telemetryService.logPageView('SiteAdminGitHubApp')
    }, [telemetryService])

    const { data, loading, error } = useQuery<GitHubAppByIDResult, GitHubAppByIDVariables>(GITHUB_APP_BY_ID_QUERY, {
        variables: { id: appID ?? '' },
    })

    const app = useMemo(() => data?.gitHubApp, [data])

    // TODO - make an actual GraphQL request to do it here...
    const refreshFromGH = (): void => {}

    if (!appID) {
        return null
    }

    return (
        <div>
            {app ? <PageTitle title={`GitHub App - ${app.name}`} /> : <PageTitle title="GitHub App" />}
            {error && <ErrorAlert className="mb-3" error={error} />}
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
                                <Button onClick={refreshFromGH} variant="info" className="ml-auto">
                                    <Icon inline={true} svgPath={mdiRefresh} aria-hidden={true} /> Refresh from GitHub
                                </Button>
                                <ButtonLink to={app.appURL} variant="info" className="ml-2">
                                    <Icon inline={true} svgPath={mdiGithub} aria-hidden={true} /> Edit
                                </ButtonLink>
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
                    <hr />

                    <div className="mt-4">
                        <H2>App installations</H2>
                        <div className="list-group mb-3" aria-label="GitHub App Installations">
                            {app.installations?.map(installation => (
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
                                    <ButtonLink to={installation.url} variant="secondary" className="ml-auto" size="sm">
                                        <Icon inline={true} svgPath={mdiGithub} aria-hidden={true} /> Edit
                                    </ButtonLink>
                                </Card>
                            ))}
                        </div>
                        <ButtonLink
                            to={
                                app.appURL.endsWith('/')
                                    ? app.appURL + 'installations/new'
                                    : app.appURL + '/installations/new'
                            }
                            variant="success"
                        >
                            <Icon svgPath={mdiPlus} aria-hidden={true} /> Add installation
                        </ButtonLink>
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
                        <Button variant="success">
                            <Icon svgPath={mdiPlus} aria-hidden={true} /> Add connection
                        </Button>
                    </div>
                </Container>
            )}
        </div>
    )
}
