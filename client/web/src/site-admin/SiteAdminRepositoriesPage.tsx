import React, { useEffect, useCallback } from 'react'

import CloudDownloadIcon from 'mdi-react/CloudDownloadIcon'
import CloudOutlineIcon from 'mdi-react/CloudOutlineIcon'
import SettingsIcon from 'mdi-react/SettingsIcon'
import { RouteComponentProps } from 'react-router'
import { Observable } from 'rxjs'

import { RepoLink } from '@sourcegraph/shared/src/components/RepoLink'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { LoadingSpinner, Button, Link, Alert, Icon, Typography } from '@sourcegraph/wildcard'

import {
    FilteredConnection,
    FilteredConnectionFilter,
    FilteredConnectionQueryArguments,
} from '../components/FilteredConnection'
import { PageTitle } from '../components/PageTitle'
import { RepositoriesResult, SiteAdminRepositoryFields } from '../graphql-operations'
import { refreshSiteFlags } from '../site/backend'

import { fetchAllRepositoriesAndPollIfEmptyOrAnyCloning } from './backend'

interface RepositoryNodeProps {
    node: SiteAdminRepositoryFields
}

const RepositoryNode: React.FunctionComponent<React.PropsWithChildren<RepositoryNodeProps>> = ({ node }) => (
    <li
        className="repository-node list-group-item py-2"
        data-test-repository={node.name}
        data-test-cloned={node.mirrorInfo.cloned}
    >
        <div className="d-flex align-items-center justify-content-between">
            <div>
                <RepoLink repoName={node.name} to={node.url} />
                {node.mirrorInfo.cloneInProgress && (
                    <small className="ml-2 text-success">
                        <LoadingSpinner /> Cloning
                    </small>
                )}
                {!node.mirrorInfo.cloneInProgress && !node.mirrorInfo.cloned && (
                    <small
                        className="ml-2 text-muted"
                        data-tooltip="Visit the repository to clone it. See its mirroring settings for diagnostics."
                    >
                        <Icon as={CloudOutlineIcon} /> Not yet cloned
                    </small>
                )}
            </div>
            <div className="repository-node__actions">
                {!node.mirrorInfo.cloneInProgress && !node.mirrorInfo.cloned && (
                    <Button to={node.url} variant="secondary" size="sm" as={Link}>
                        <Icon as={CloudDownloadIcon} /> Clone now
                    </Button>
                )}{' '}
                {
                    <Button
                        to={`/${node.name}/-/settings`}
                        data-tooltip="Repository settings"
                        variant="secondary"
                        size="sm"
                        as={Link}
                    >
                        <Icon as={SettingsIcon} /> Settings
                    </Button>
                }{' '}
            </div>
        </div>
    </li>
)

interface Props extends RouteComponentProps<{}>, TelemetryProps {}

const FILTERS: FilteredConnectionFilter[] = [
    {
        id: 'status',
        label: 'Status',
        type: 'radio',
        values: [
            {
                label: 'All',
                value: 'all',
                tooltip: 'Show all repositories',
                args: {},
            },
            {
                label: 'Cloned',
                value: 'cloned',
                tooltip: 'Show cloned repositories only',
                args: { cloned: true, notCloned: false },
            },
            {
                label: 'Not cloned',
                value: 'not-cloned',
                tooltip: 'Show only repositories that have not been cloned yet',
                args: { cloned: false, notCloned: true },
            },
            {
                label: 'Needs index',
                value: 'needs-index',
                tooltip: 'Show only repositories that need to be indexed',
                args: { indexed: false },
            },
            {
                label: 'Failed fetch/clone',
                value: 'failed-fetch',
                tooltip: 'Show only repositories that have failed to fetch or clone',
                args: { failedFetch: true },
            },
        ],
    },
]

/**
 * A page displaying the repositories on this site.
 */
export const SiteAdminRepositoriesPage: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    history,
    location,
    telemetryService,
}) => {
    useEffect(() => {
        telemetryService.logViewEvent('SiteAdminRepos')
    }, [telemetryService])

    // Refresh global alert about enabling repositories when the user visits & navigates away from this page.
    useEffect(() => {
        refreshSiteFlags()
            .toPromise()
            .then(null, error => console.error(error))
        return () => {
            refreshSiteFlags()
                .toPromise()
                .then(null, error => console.error(error))
        }
    }, [])
    const queryRepositories = useCallback(
        (args: FilteredConnectionQueryArguments): Observable<RepositoriesResult['repositories']> =>
            fetchAllRepositoriesAndPollIfEmptyOrAnyCloning(args),
        []
    )
    const showRepositoriesAddedBanner = new URLSearchParams(location.search).has('repositoriesUpdated')

    return (
        <div className="site-admin-repositories-page">
            <PageTitle title="Repositories - Admin" />
            {showRepositoriesAddedBanner && (
                <Alert variant="success" as="p">
                    Updating repositories. It may take a few moments to clone and index each repository. Repository
                    statuses are displayed below.
                </Alert>
            )}
            <Typography.H2>Repositories</Typography.H2>
            <p>
                Repositories are synced from connected{' '}
                <Link to="/site-admin/external-services" data-testid="test-repositories-code-host-connections-link">
                    code host connections
                </Link>
                .
            </p>
            <FilteredConnection<SiteAdminRepositoryFields, Omit<RepositoryNodeProps, 'node'>>
                className="list-group list-group-flush mt-3"
                noun="repository"
                pluralNoun="repositories"
                queryConnection={queryRepositories}
                nodeComponent={RepositoryNode}
                filters={FILTERS}
                history={history}
                location={location}
            />
        </div>
    )
}
