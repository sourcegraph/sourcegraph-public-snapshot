import React, { useCallback, useEffect, useState } from 'react'
import { PageTitle } from '../../../components/PageTitle'
import { RepositoriesResult, SiteAdminRepositoryFields, UserRepositoriesResult } from '../../../graphql-operations'
import { TelemetryProps } from '../../../../../shared/src/telemetry/telemetryService'
import {
    Connection,
    FilteredConnection,
    FilteredConnectionFilter,
    FilteredConnectionQueryArguments,
    FilterValue,
} from '../../../components/FilteredConnection'
import { Observable } from 'rxjs'
import { listUserRepositories } from '../../../site-admin/backend'
import { queryExternalServices } from '../../../components/externalServices/backend'
import { RouteComponentProps } from 'react-router'
import { Link } from '../../../../../shared/src/components/Link'
import { RepositoryNode } from './RepositoryNode'
import AddIcon from 'mdi-react/AddIcon'
import { asError, ErrorLike, isErrorLike } from '../../../../../shared/src/util/errors'
import { repeatUntil } from '../../../../../shared/src/util/rxjs/repeatUntil'
import { ErrorAlert } from '../../../components/alerts'

interface Props extends RouteComponentProps, TelemetryProps {
    userID: string
    routingPrefix: string
}

interface RowProps {
    node: SiteAdminRepositoryFields
}

const Row: React.FunctionComponent<RowProps> = props => (
    <RepositoryNode
        name={props.node.name}
        url={props.node.url}
        serviceType={props.node.externalRepository.serviceType.toUpperCase()}
        mirrorInfo={props.node.mirrorInfo}
        isPrivate={props.node.isPrivate}
    />
)

type Status = undefined | 'pending' | ErrorLike

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
    const [hasRepos, setHasRepos] = useState(false)
    const [pendingOrError, setPendingOrError] = useState<Status>()

    if (!state.fetched) {
        queryExternalServices({ namespace: userID, first: null, after: null })
            .pipe(
                repeatUntil(
                    (result): boolean => {
                        const services: FilterValue[] = [
                            {
                                value: 'all',
                                label: 'All',
                                args: {},
                            },
                        ]
                        let pending = true
                        const now = new Date().getTime()
                        result.nodes.map(node => {
                            // if the next sync time is not blank, or in the future we must not be syncing
                            if (node.nextSyncAt !== '' && now - new Date(node.nextSyncAt).getTime() < 0) {
                                pending = false
                            }
                            services.push({
                                value: node.id,
                                label: node.displayName,
                                tooltip: '',
                                args: { externalServiceID: node.id },
                            })
                        })
                        setPendingOrError(pending && 'pending')
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
                        return !pending
                    },
                    { delay: 2000 }
                )
            )
            .toPromise()
            .catch(error => {
                setPendingOrError(asError(error))
            })
    }

    const queryRepositories = useCallback(
        (args: FilteredConnectionQueryArguments): Observable<RepositoriesResult['repositories']> =>
            listUserRepositories({ ...args, id: userID }).pipe(
                repeatUntil(
                    (result): boolean =>
                        result.nodes &&
                        result.nodes.length > 0 &&
                        result.nodes.every(nodes => !nodes.mirrorInfo.cloneInProgress && nodes.mirrorInfo.cloned) &&
                        !(pendingOrError === 'pending'),
                    { delay: 2000 }
                )
            ),
        [pendingOrError, userID]
    )

    const updated = useCallback((value: Connection<SiteAdminRepositoryFields> | ErrorLike | undefined): void => {
        if (value as Connection<SiteAdminRepositoryFields>) {
            const conn = value as Connection<SiteAdminRepositoryFields>
            if (conn.totalCount !== 0) {
                setHasRepos(true)
            }
        }
    }, [])

    useEffect(() => {
        telemetryService.logViewEvent('UserSettingsRepositories')
    }, [telemetryService])

    let body: JSX.Element
    if (state.filters[1] && state.filters[1].values.length === 1) {
        body = (
            <div className="card p-3 m-2">
                <h3 className="mb-1">You have not added any repositories to Sourcegraph</h3>
                <p className="text-muted mb-0">
                    <a className="text-primary" href={routingPrefix + '/external-services'}>
                        Connect a code host
                    </a>{' '}
                    to start adding your repositories to Sourcegraph.
                </p>
            </div>
        )
    } else {
        body = (
            <FilteredConnection<SiteAdminRepositoryFields, Omit<UserRepositoriesResult, 'node'>>
                className="table mt-3"
                defaultFirst={15}
                compact={false}
                noun="repository"
                pluralNoun="repositories"
                queryConnection={queryRepositories}
                nodeComponent={Row}
                listComponent="table"
                listClassName="w-100"
                onUpdate={updated}
                filters={state.filters}
                history={history}
                location={location}
                totalCountSummaryComponent={TotalCountSummary}
            />
        )
    }

    return (
        <div className="user-settings-repositories-page">
            {pendingOrError === 'pending' && (
                <div className="alert alert-info">
                    <span className="font-weight-bold">Some repositories are still being fetched.</span>
                    These repositories may not appear in the list of repositories.
                </div>
            )}
            {isErrorLike(pendingOrError) && <ErrorAlert error={pendingOrError} icon={true} history={history} />}
            <PageTitle title="Repositories" />
            <div className="d-flex justify-content-between align-items-center">
                <h2 className="mb-2">Repositories</h2>
                {state.filters[1] && state.filters[1].values.length !== 1 && (
                    <Link
                        className="btn btn-primary test-goto-add-external-service-page"
                        to={`${routingPrefix}/repositories/manage`}
                    >
                        {(hasRepos && <>Manage Repositories</>) || (
                            <>
                                <AddIcon className="icon-inline" /> Add repositories
                            </>
                        )}
                    </Link>
                )}
            </div>
            <p className="text-muted pb-2">
                All repositories synced with Sourcegraph from{' '}
                <Link className="text-primary" to={`${routingPrefix}/code-hosts`}>
                    connected code hosts
                </Link>
            </p>
            {body}
        </div>
    )
}

const TotalCountSummary: React.FunctionComponent<{ totalCount: number }> = ({ totalCount }) => (
    <small>{totalCount} repositories total</small>
)
