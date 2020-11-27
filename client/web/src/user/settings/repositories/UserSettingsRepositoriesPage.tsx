import React, { useCallback, useEffect, useState } from 'react'
import { PageTitle } from '../../../components/PageTitle'
import { RepositoriesResult, SiteAdminRepositoryFields } from '../../../graphql-operations'
import { TelemetryProps } from '../../../../../shared/src/telemetry/telemetryService'
import {
    FilteredConnection,
    FilteredConnectionFilter,
    FilteredConnectionQueryArguments,
    FilterValue,
} from '../../../components/FilteredConnection'
import { Observable } from 'rxjs'
import { listUserRepositories } from '../../../site-admin/backend'
import { queryExternalServices } from '../../../components/externalServices/backend'
import { RouteComponentProps } from 'react-router'
import { RepoLink } from '../../../../../shared/src/components/RepoLink'
import { Link } from '../../../../../shared/src/components/Link'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import GithubIcon from 'mdi-react/GithubIcon'
import CloudOutlineIcon from 'mdi-react/CloudOutlineIcon'
import TickIcon from 'mdi-react/TickIcon'
import GitlabIcon from 'mdi-react/GitlabIcon'
import BitbucketIcon from 'mdi-react/BitbucketIcon'
import AddIcon from 'mdi-react/AddIcon'
import SourceRepositoryIcon from 'mdi-react/SourceRepositoryIcon'
import ChevronRightIcon from 'mdi-react/ChevronRightIcon';

interface UserSettingsRepositoryNodeProps {
    node: SiteAdminRepositoryFields
}

interface StatusIconProps {
    node: SiteAdminRepositoryFields
}

const StatusIcon: React.FunctionComponent<StatusIconProps> = ({ node }) => {
    if (node.mirrorInfo.cloneInProgress) {
        return (
            <small data-tooltip="Clone in progress." className="mr-2 text-success">
                <LoadingSpinner className="icon-inline" />
            </small>
        )
    }
    if (!node.mirrorInfo.cloneInProgress && !node.mirrorInfo.cloned) {
        return (
            <small
                className="mr-2 text-muted"
                data-tooltip="Visit the repository to clone it. See its mirroring settings for diagnostics."
            >
                <CloudOutlineIcon className="icon-inline" />
            </small>
        )
    }
    return (
        <small className="mr-2">
            <TickIcon className="icon-inline" />
        </small>
    )
}

interface CodeHostIconProps {
    hostType: string
}

const CodeHostIcon: React.FunctionComponent<CodeHostIconProps> = ({ hostType }) => {
    switch (hostType) {
        case 'github':
            return (
                <small className="mr-2">
                    <GithubIcon className="icon-inline" />
                </small>
            )
        case 'gitlab':
            return (
                <small className="mr-2">
                    <GitlabIcon className="icon-inline" />
                </small>
            )
        case 'bitbucket':
            return (
                <small className="mr-2">
                    <BitbucketIcon className="icon-inline" />
                </small>
            )
    }
    return (
        <small className="mr-2">
            <SourceRepositoryIcon className="icon-inline" />
        </small>
    )
}

const UserSettingsRepositoryNode: React.FunctionComponent<UserSettingsRepositoryNodeProps> = ({ node }) => (
    <tr
        className="w-100 repository-node d-flex align-items-center justify-content-between"
        data-test-repository={node.name}
        data-test-cloned={node.mirrorInfo.cloned}
    >
        <td className="w-100 d-flex justify-content-between align-items-baseline">
            <div>
                <StatusIcon node={node} />
                <CodeHostIcon hostType={node.externalRepository.serviceType} />
                <RepoLink className="text-muted" repoClassName="text-primary" repoName={node.name} to={node.url} />
            </div>
            <div>
                {node.isPrivate && <div className="badge badge-secondary text-muted">Private</div>}
                <ChevronRightIcon className="icon-inline ml-2"/>
            </div>
        </td>
    </tr>
)

interface Props extends RouteComponentProps, TelemetryProps {
    userID: string
    routingPrefix: string
}

/**
 * A page displaying the repositories for this user.
 */
export const UserSettingsRepositoriesPage: React.FunctionComponent<Props> = ({
    history,
    location,
    userID,
    routingPrefix,
    telemetryService,
}) => {
    const emptyFilters: FilteredConnectionFilter[] = []
    const [state, setState] = useState({ filters: emptyFilters, fetched: false })

    if (!state.fetched) {
        queryExternalServices({ namespace: userID, first: null, after: null })
            .toPromise()
            .then(result => {
                const services: FilterValue[] = [
                    {
                        value: 'all',
                        label: 'All',
                        args: {},
                    },
                ]
                result.nodes.map(node => {
                    services.push({
                        value: node.id,
                        label: node.displayName,
                        tooltip: '',
                        args: { externalServiceID: node.id },
                    })
                })
                const newFilters: FilteredConnectionFilter[] = [
                    {
                        label: 'Status',
                        type: 'select',
                        id: 'status',
                        tooltip: 'Repository status',
                        values: [
                            {
                                value: 'all',
                                label: 'All',
                                args: {},
                            },
                            {
                                value: 'cloned',
                                label: 'Cloned',
                                args: { cloned: true, notCloned: false },
                            },
                            {
                                value: 'not-cloned',
                                label: 'Not Cloned',
                                args: { cloned: false, notCloned: true },
                            },
                        ],
                    },
                    {
                        label: 'Code host',
                        type: 'select',
                        id: 'code-host',
                        tooltip: 'Code host',
                        values: services,
                    },
                ]
                setState({ filters: newFilters, fetched: true })
            })
            .catch(error => {
                console.log('ERROR', error)
            })
    }

    const queryRepositories = useCallback(
        (args: FilteredConnectionQueryArguments): Observable<RepositoriesResult['repositories']> =>
            listUserRepositories({ ...args, id: userID }),
        [userID]
    )

    useEffect(() => {
        telemetryService.logViewEvent('UserSettingsRepositories')
    }, [telemetryService])

    let body: JSX.Element
    if (state.filters[1] && state.filters[1].values.length === 1) {
        body = <div className="card p-3 m-2">
            <h3 className="mb-1">You have not added any repositories to Sourcegraph</h3>
            <p className="text-muted mb-0"><a className="text-primary" href={ routingPrefix +'/external-services' }>Connect a code host</a> to start adding your repositories to Sourcegraph.</p>
        </div>
    } else {
        body = <FilteredConnection<SiteAdminRepositoryFields, Omit<UserSettingsRepositoryNodeProps, 'node'>>
            className="table mt-3"
            defaultFirst={15}
            compact={false}
            noun="repository"
            pluralNoun="repositories"
            queryConnection={queryRepositories}
            nodeComponent={UserSettingsRepositoryNode}
            listComponent="table"
            listClassName="w-100"
            filters={state.filters}
            history={history}
            location={location}
        />
    }

    return (
        <div className="user-settings-repositories-page">
                <PageTitle title="Repositories" />
            <div className="d-flex justify-content-between align-items-center">
                <h2 className="mb-2">Repositories</h2>
                {state.filters[1] && state.filters[1].values.length !== 1 && (
                    <Link
                        className="btn btn-primary test-goto-add-external-service-page"
                        to={`${routingPrefix}/external-services`}
                    >
                        <AddIcon className="icon-inline" /> Manage repositories
                    </Link>
                )}
            </div>
            <p className="text-muted">
                All repositories synced with Sourcegraph from <a className="text-primary" href={routingPrefix+'/external-services'}>connected code hosts</a>
            </p>
            {body}
        </div>
    )
}
