import classNames from 'classnames'
import AddIcon from 'mdi-react/AddIcon'
import React, { useCallback, useEffect, useMemo, useState } from 'react'
import { RouteComponentProps } from 'react-router'
import { EMPTY, Observable } from 'rxjs'
import { catchError, tap } from 'rxjs/operators'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { Link } from '@sourcegraph/shared/src/components/Link'
import { dataOrThrowErrors, gql } from '@sourcegraph/shared/src/graphql/graphql'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { asError, ErrorLike, isErrorLike } from '@sourcegraph/shared/src/util/errors'
import { repeatUntil } from '@sourcegraph/shared/src/util/rxjs/repeatUntil'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'
import { useRedesignToggle } from '@sourcegraph/shared/src/util/useRedesignToggle'
import { Container, PageHeader } from '@sourcegraph/wildcard'

import { requestGraphQL } from '../../../backend/graphql'
import { ErrorAlert } from '../../../components/alerts'
import { queryExternalServices } from '../../../components/externalServices/backend'
import {
    FilteredConnection,
    FilteredConnectionFilter,
    FilteredConnectionQueryArguments,
    Connection,
} from '../../../components/FilteredConnection'
import { PageTitle } from '../../../components/PageTitle'
import {
    SiteAdminRepositoryFields,
    UserRepositoriesResult,
    UserRepositoriesVariables,
    ExternalServicesResult,
    CodeHostSyncDueResult,
    CodeHostSyncDueVariables,
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

const DEFAULT_FILTERS: FilteredConnectionFilter[] = [
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
        values: [
            {
                value: 'all',
                label: 'All',
                args: {},
            },
        ],
    },
]

const Row: React.FunctionComponent<RowProps> = props => (
    <RepositoryNode
        name={props.node.name}
        url={props.node.url}
        serviceType={props.node.externalRepository.serviceType.toUpperCase()}
        mirrorInfo={props.node.mirrorInfo}
        isPrivate={props.node.isPrivate}
    />
)

type SyncStatusOrError = undefined | 'scheduled' | 'schedule-complete' | ErrorLike

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
    const [isRedesignEnabled] = useRedesignToggle()
    const [hasRepos, setHasRepos] = useState(false)
    const [externalServices, setExternalServices] = useState<ExternalServicesResult['externalServices']['nodes']>()
    const [repoFilters, setRepoFilters] = useState<FilteredConnectionFilter[]>([])
    const [status, setStatus] = useState<SyncStatusOrError>()
    const [updateReposList, setUpdateReposList] = useState(false)

    const NoAddedReposBanner = (
        <Container
            className={classNames(isRedesignEnabled && 'text-center', !isRedesignEnabled && 'border rounded p-3')}
        >
            <h4>You have not added any repositories to Sourcegraph.</h4>

            {externalServices?.length === 0 ? (
                <small>
                    <Link to={`${routingPrefix}/code-hosts`}>Connect code hosts</Link> to start searching your own
                    repositories, or <Link to={`${routingPrefix}/repositories/manage`}>add public repositories</Link>{' '}
                    from GitHub or GitLab.
                </small>
            ) : (
                <small>
                    <Link to={`${routingPrefix}/repositories/manage`}>Add repositories</Link> to start searching your
                    code with Sourcegraph.
                </small>
            )}
        </Container>
    )

    const fetchUserReposCount = useCallback(
        async (): Promise<UserRepositoriesResult> =>
            dataOrThrowErrors(
                await requestGraphQL<UserRepositoriesResult, UserRepositoriesVariables>(
                    gql`
                        query UserRepositoriesTotalCount(
                            $id: ID!
                            $first: Int
                            $query: String
                            $cloned: Boolean
                            $notCloned: Boolean
                            $indexed: Boolean
                            $notIndexed: Boolean
                            $externalServiceID: ID
                        ) {
                            node(id: $id) {
                                ... on User {
                                    repositories(
                                        first: $first
                                        query: $query
                                        cloned: $cloned
                                        notCloned: $notCloned
                                        indexed: $indexed
                                        notIndexed: $notIndexed
                                        externalServiceID: $externalServiceID
                                    ) {
                                        totalCount(precise: true)
                                    }
                                }
                            }
                        }
                    `,
                    {
                        id: userID,
                        cloned: true,
                        notCloned: true,
                        indexed: true,
                        notIndexed: true,
                        first: null,
                        query: null,
                        externalServiceID: null,
                    }
                ).toPromise()
            ),

        [userID]
    )

    const fetchExternalServices = useCallback(
        async (): Promise<ExternalServicesResult['externalServices']['nodes']> =>
            queryExternalServices({
                first: null,
                after: null,
                namespace: userID,
            })
                .toPromise()
                .then(({ nodes }) => nodes),

        [userID]
    )

    const fetchCodeHostSyncDueStatus = useCallback(
        (ids: string[], seconds: number) =>
            requestGraphQL<CodeHostSyncDueResult, CodeHostSyncDueVariables>(
                gql`
                    query CodeHostSyncDue($ids: [ID!]!, $seconds: Int!) {
                        codeHostSyncDue(ids: $ids, seconds: $seconds)
                    }
                `,
                { ids, seconds }
            ),
        []
    )

    const init = useCallback(async (): Promise<void> => {
        // fetch and set external services
        const services = await fetchExternalServices()
        setExternalServices(services)

        // check if user has any manually added or affiliated repositories
        const result = await fetchUserReposCount()
        if (result?.node?.repositories?.totalCount && result.node.repositories.totalCount > 0) {
            setHasRepos(true)
        }

        // configure filters
        const specificCodeHostFilters = services.map(service => ({
            tooltip: '',
            value: service.id,
            label: service.displayName.split(' ')[0],
            args: { externalServiceID: service.id },
        }))

        const [statusFilter, codeHostFilter] = DEFAULT_FILTERS

        // update default code host filter by adding GitLab and/or GitHub filters
        const updatedCodeHostFilter = {
            ...codeHostFilter,
            values: [...codeHostFilter.values, ...specificCodeHostFilters],
        }

        setRepoFilters([statusFilter, updatedCodeHostFilter])
    }, [fetchExternalServices, fetchUserReposCount])

    const TWO_SECONDS = 2

    useObservable(
        useMemo(() => {
            if (externalServices && externalServices.length !== 0) {
                // get serviceIds and check if services will sync in the next 2 seconds
                const serviceIds = externalServices.map(service => service.id)

                return fetchCodeHostSyncDueStatus(serviceIds, TWO_SECONDS).pipe(
                    repeatUntil(
                        result => {
                            const isScheduledToSync = result.data?.codeHostSyncDue === true
                            // if all existing code hosts were just added
                            // created and updated timestamps are the same
                            const areCodeHostsJustAdded = externalServices.every(
                                ({ updatedAt, createdAt, repoCount }) => updatedAt === createdAt && repoCount === 0
                            )

                            if (isScheduledToSync && !areCodeHostsJustAdded) {
                                setStatus('scheduled')
                            } else {
                                setStatus(previousState => {
                                    if (previousState === 'scheduled') {
                                        return 'schedule-complete'
                                    }

                                    return undefined
                                })
                            }

                            // don't repeat the query if the sync is not scheduled
                            // or code host(s) we just added
                            return !isScheduledToSync || areCodeHostsJustAdded
                        },
                        { delay: 2000 }
                    ),
                    catchError(error => {
                        setStatus(asError(error))
                        return EMPTY
                    })
                )
            }

            return EMPTY
        }, [externalServices, fetchCodeHostSyncDueStatus])
    )

    useEffect(() => {
        // don't re-fetch data when sync is scheduled or we had an error
        // we should fetch only on the page load or once the sync is complete
        if (status === 'scheduled' || isErrorLike(status)) {
            return
        }

        init().catch(error => setStatus(asError(error)))
    }, [init, status])

    const queryRepositories = useCallback(
        (
            args: FilteredConnectionQueryArguments
        ): Observable<NonNullable<UserRepositoriesResult['node']>['repositories']> =>
            listUserRepositories({ ...args, id: userID }).pipe(
                tap(() => {
                    if (status === 'schedule-complete') {
                        setUpdateReposList(!updateReposList)
                        setStatus(undefined)
                    }
                })
            ),
        [userID, status, updateReposList]
    )

    const onRepoQueryUpdate = useCallback(
        (value: Connection<SiteAdminRepositoryFields> | ErrorLike | undefined, query: string): void => {
            if (value as Connection<SiteAdminRepositoryFields>) {
                const conn = value as Connection<SiteAdminRepositoryFields>

                // hasRepos is only useful when query is not set since user may
                // still have repos that don't match given query
                if (query === '') {
                    if (conn.totalCount !== 0 || conn.nodes.length !== 0) {
                        setHasRepos(true)
                    } else {
                        setHasRepos(false)
                    }
                }
            }
        },
        []
    )

    const NoMatchedRepos = (
        <div className="border rounded p-3">
            <small>No repositories matched.</small>
        </div>
    )

    const RepoFilteredConnection = (
        <Container>
            <FilteredConnection<SiteAdminRepositoryFields, Omit<UserRepositoriesResult, 'node'>>
                className="table mb-0"
                defaultFirst={15}
                compact={false}
                noun="repository"
                pluralNoun="repositories"
                queryConnection={queryRepositories}
                updateOnChange={String(updateReposList)}
                nodeComponent={Row}
                listComponent="table"
                listClassName="w-100"
                onUpdate={onRepoQueryUpdate}
                filters={repoFilters}
                history={history}
                location={location}
                emptyElement={NoMatchedRepos}
                totalCountSummaryComponent={TotalCountSummary}
                inputClassName="user-settings-repos__filter-input"
            />
        </Container>
    )

    const logManageRepositoriesClick = useCallback(() => {
        eventLogger.log('UserSettingsRepositoriesManageRepositoriesClick')
    }, [])

    useEffect(() => {
        telemetryService.logViewEvent('UserSettingsRepositories')
    }, [telemetryService])

    const getCodeHostsSyncMessage = (): string => {
        if (Array.isArray(externalServices) && externalServices) {
            const names = externalServices.map(service => {
                const { displayName: name } = service
                const namespaceStartIndex = name.indexOf('(')

                return namespaceStartIndex !== -1 ? name.slice(0, namespaceStartIndex - 1) : name
            })

            return `Synching ${names.join(', ')} code host${names.length > 1 ? 's' : ''}.`
        }
        return 'Synching code hosts.'
    }

    return (
        <div className="user-settings-repos">
            {status === 'scheduled' && (
                <div className="alert alert-info">
                    <span className="font-weight-bold">{getCodeHostsSyncMessage()}</span> Repositories list may not be
                    up-to-date and will refresh once the sync is finished.
                </div>
            )}
            {isErrorLike(status) && <ErrorAlert error={status} icon={true} />}
            <PageTitle title="Repositories" />
            <PageHeader
                headingElement="h2"
                path={[{ text: 'Repositories' }]}
                description={
                    <>
                        All repositories synced with Sourcegraph from{' '}
                        <Link to={`${routingPrefix}/code-hosts`}>connected code hosts</Link>
                    </>
                }
                actions={
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
                }
                className="mb-3"
            />
            {isErrorLike(status) ? (
                <h3 className="text-muted">Sorry, we couldn’t fetch your repositories. Try again?</h3>
            ) : !externalServices ? (
                <div className="d-flex justify-content-center mt-4">
                    <LoadingSpinner className="icon-inline" />
                </div>
            ) : hasRepos ? (
                RepoFilteredConnection
            ) : (
                NoAddedReposBanner
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
