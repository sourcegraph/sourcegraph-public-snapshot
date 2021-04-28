import { isAfter, parseISO } from 'date-fns'
import AddIcon from 'mdi-react/AddIcon'
import React, { useCallback, useEffect, useMemo, useState } from 'react'
import { RouteComponentProps } from 'react-router'
import { EMPTY, Observable } from 'rxjs'
import { catchError, map } from 'rxjs/operators'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { Link } from '@sourcegraph/shared/src/components/Link'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { asError, ErrorLike, isErrorLike } from '@sourcegraph/shared/src/util/errors'
import { repeatUntil } from '@sourcegraph/shared/src/util/rxjs/repeatUntil'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'

import { ErrorAlert } from '../../../components/alerts'
import { queryExternalServices } from '../../../components/externalServices/backend'
import {
    Connection,
    FilteredConnection,
    FilteredConnectionFilter,
    FilteredConnectionQueryArguments,
    FilterValue,
} from '../../../components/FilteredConnection'
import { PageTitle } from '../../../components/PageTitle'
import {
    RepositoriesResult,
    SiteAdminRepositoryFields,
    UserRepositoriesResult,
    ListExternalServiceFields,
} from '../../../graphql-operations'
import { listUserRepositories } from '../../../site-admin/backend'
import { eventLogger } from '../../../tracking/eventLogger'

import { RepositoryNode } from './RepositoryNode'

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
const emptyFilters: FilteredConnectionFilter[] = []

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
    const [hasRepos, setHasRepos] = useState<boolean | null>(null)
    const [externalServices, setExternalServices] = useState<ListExternalServiceFields[]>()
    const [pendingOrError, setPendingOrError] = useState<Status>()

    const noReposState = (
        <div className="border rounded p-3">
            <h3>You have not added any repositories to Sourcegraph</h3>

            {externalServices?.length === 0 ? (
                <small>
                    <Link className="text-primary" to={`${routingPrefix}/code-hosts`}>
                        Connect code hosts
                    </Link>{' '}
                    to start searching your own repositories, or{' '}
                    <Link className="text-primary" to={`${routingPrefix}/repositories/manage`}>
                        add public repositories
                    </Link>{' '}
                    from GitHub or GitLab.
                </small>
            ) : (
                <small>
                    <Link className="text-primary" to={`${routingPrefix}/repositories/manage`}>
                        Add repositories
                    </Link>{' '}
                    to start searching your code with Sourcegraph.
                </small>
            )}
        </div>
    )
    const showResults = (): JSX.Element => {
        const emptyState = (
            <div className="border rounded p-3">
                <small>No repositories matched.</small>
            </div>
        )
        return (
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
                filters={filters}
                history={history}
                location={location}
                emptyElement={emptyState}
                totalCountSummaryComponent={TotalCountSummary}
                inputClassName="user-settings-repos__filter-input"
            />
        )
    }

    const filters =
        useObservable<FilteredConnectionFilter[]>(
            useMemo(
                () =>
                    queryExternalServices({ namespace: userID, first: null, after: null }).pipe(
                        repeatUntil(
                            result => {
                                let pending: Status
                                const now = new Date()

                                setExternalServices(result.nodes)

                                let repoCount = 0

                                for (const node of result.nodes) {
                                    repoCount += node.repoCount

                                    // when the service was just added both
                                    // createdAt and updatedAt will have the same timestamp
                                    if (node.createdAt === node.updatedAt) {
                                        continue
                                    }

                                    // if the next sync is in the future we must not be syncing
                                    if (node.nextSyncAt) {
                                        if (isAfter(now, parseISO(node.nextSyncAt))) {
                                            pending = 'pending'
                                        }
                                    }
                                }

                                if (repoCount > 0) {
                                    setHasRepos(true)
                                }

                                setPendingOrError(pending)
                                return pending !== 'pending'
                            },
                            { delay: 2000 }
                        ),
                        map(result => {
                            const services: FilterValue[] = [
                                {
                                    value: 'all',
                                    label: 'All',
                                    args: {},
                                },
                                ...result.nodes.map(node => ({
                                    value: node.id,
                                    label: node.displayName.split(' ')[0],
                                    tooltip: '',
                                    args: { externalServiceID: node.id },
                                })),
                            ]

                            return [
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
                        }),
                        catchError(error => {
                            setPendingOrError(asError(error))
                            return EMPTY
                        })
                    ),
                [userID]
            )
        ) || emptyFilters

    const queryRepositories = useCallback(
        (args: FilteredConnectionQueryArguments): Observable<RepositoriesResult['repositories']> =>
            listUserRepositories({ ...args, id: userID }).pipe(
                repeatUntil(
                    (result): boolean => {
                        // don't repeat the query when user doesn't have repos
                        if (result.nodes && result.nodes.length === 0) {
                            return true
                        }

                        return (
                            result.nodes &&
                            result.nodes.length > 0 &&
                            result.nodes.every(nodes => !nodes.mirrorInfo.cloneInProgress && nodes.mirrorInfo.cloned) &&
                            !(pendingOrError === 'pending')
                        )
                    },

                    { delay: 2000 }
                )
            ),
        [pendingOrError, userID]
    )

    const updated = useCallback(
        (value: Connection<SiteAdminRepositoryFields> | ErrorLike | undefined, query: string): void => {
            if (value as Connection<SiteAdminRepositoryFields>) {
                const conn = value as Connection<SiteAdminRepositoryFields>

                // hasRepos is only useful when query is not set since user may
                // still have repos that don't match given query
                if (query === '') {
                    if (conn.totalCount !== 0) {
                        setHasRepos(true)
                    } else {
                        setHasRepos(false)
                    }
                }
            }
        },
        []
    )
    const logManageRepositoriesClick = useCallback(() => {
        eventLogger.log('UserSettingsRepositoriesManageRepositoriesClick')
    }, [])

    useEffect(() => {
        telemetryService.logViewEvent('UserSettingsRepositories')
    }, [telemetryService])

    return (
        <div className="user-settings-repos">
            {pendingOrError === 'pending' && (
                <div className="alert alert-info">
                    <span className="font-weight-bold">Some repositories are still being updated.</span> These
                    repositories may not appear up-to-date in the list of repositories.
                </div>
            )}
            {isErrorLike(pendingOrError) && <ErrorAlert error={pendingOrError} icon={true} />}
            <PageTitle title="Repositories" />
            <div className="d-flex justify-content-between align-items-center">
                <h2 className="mb-2">Repositories</h2>
                <Link
                    className="btn btn-primary"
                    to={`${routingPrefix}/repositories/manage`}
                    onClick={logManageRepositoriesClick}
                >
                    {(hasRepos && <>Manage Repositories</>) || (
                        <>
                            <AddIcon className="icon-inline" /> Add repositories
                        </>
                    )}
                </Link>
            </div>
            <p className="text-muted pb-2">
                All repositories synced with Sourcegraph from{' '}
                <Link className="text-primary" to={`${routingPrefix}/code-hosts`}>
                    connected code hosts
                </Link>
            </p>
            {externalServices ? (
                <>
                    {hasRepos === false && noReposState}
                    {(hasRepos || hasRepos === null) && showResults()}
                </>
            ) : (
                <div className="d-flex justify-content-center mt-4">
                    <LoadingSpinner className="icon-inline" />
                </div>
            )}
        </div>
    )
}

const TotalCountSummary: React.FunctionComponent<{ totalCount: number }> = ({ totalCount }) => (
    <div className="d-inline-block mt-4 mr-2">
        <small>
            {totalCount} {totalCount === 1 ? 'repository' : 'repositories'} total
        </small>
    </div>
)
