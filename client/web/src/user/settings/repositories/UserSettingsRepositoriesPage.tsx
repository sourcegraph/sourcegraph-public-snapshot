import React, { useCallback, useEffect, useMemo, useState } from 'react'
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
import { EMPTY, Observable } from 'rxjs'
import { listUserRepositories } from '../../../site-admin/backend'
import { queryExternalServices } from '../../../components/externalServices/backend'
import { RouteComponentProps } from 'react-router'
import { Link } from '../../../../../shared/src/components/Link'
import { RepositoryNode } from './RepositoryNode'
import AddIcon from 'mdi-react/AddIcon'
import { asError, ErrorLike, isErrorLike } from '../../../../../shared/src/util/errors'
import { repeatUntil } from '../../../../../shared/src/util/rxjs/repeatUntil'
import { ErrorAlert } from '../../../components/alerts'
import { useObservable } from '../../../../../shared/src/util/useObservable'
import { catchError, map } from 'rxjs/operators'

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
    const [hasRepos, setHasRepos] = useState(false)
    const [pendingOrError, setPendingOrError] = useState<Status>()

    const filters =
        useObservable<FilteredConnectionFilter[]>(
            useMemo(
                () =>
                    queryExternalServices({ namespace: userID, first: null, after: null }).pipe(
                        repeatUntil(
                            result => {
                                let pending: Status
                                const now = new Date().getTime()
                                for (const node of result.nodes) {
                                    // if the next sync time is not blank, or in the future we must not be syncing
                                    if (node.nextSyncAt !== '' && now - new Date(node.nextSyncAt).getTime() > 0) {
                                        pending = 'pending'
                                    }
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
                                    label: node.displayName,
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
    if (filters[1] && filters[1].values.length === 1) {
        body = (
            <div className="card p-3 m-2">
                <h3 className="mb-1">You have not added any repositories to Sourcegraph</h3>
                <p className="text-muted mb-0">
                    <Link className="text-primary" to={`${routingPrefix}/code-hosts`}>
                        Connect a code host
                    </Link>{' '}
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
                filters={filters}
                history={history}
                location={location}
                totalCountSummaryComponent={TotalCountSummary}
                inputClassName="user-settings-repos__filter-input"
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
                {filters[1] && filters[1].values.length !== 1 && (
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
