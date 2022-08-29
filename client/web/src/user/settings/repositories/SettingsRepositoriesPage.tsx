import React, { useCallback, useEffect, useMemo, useState } from 'react'

import { mdiPlus } from '@mdi/js'
import { EMPTY, Observable } from 'rxjs'
import { catchError, tap } from 'rxjs/operators'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { asError, ErrorLike, isErrorLike, repeatUntil } from '@sourcegraph/common'
import { gql } from '@sourcegraph/http-client'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import {
    Container,
    PageHeader,
    ProductStatusBadge,
    LoadingSpinner,
    useObservable,
    Button,
    Alert,
    Link,
    Icon,
    H3,
    H4,
} from '@sourcegraph/wildcard'

import { requestGraphQL } from '../../../backend/graphql'
import { queryExternalServices } from '../../../components/externalServices/backend'
import {
    FilteredConnectionFilter,
    FilteredConnectionQueryArguments,
    Connection,
} from '../../../components/FilteredConnection'
import { PageTitle } from '../../../components/PageTitle'
import { SelfHostedCtaLink } from '../../../components/SelfHostedCtaLink'
import {
    SiteAdminRepositoryFields,
    ExternalServicesResult,
    CodeHostSyncDueResult,
    CodeHostSyncDueVariables,
    RepositoriesResult,
    OrgAreaOrganizationFields,
} from '../../../graphql-operations'
import { listUserRepositories, fetchUserRepositoriesCount } from '../../../site-admin/backend'
import { eventLogger } from '../../../tracking/eventLogger'
import { Owner } from '../cloud-ga'

import { UserSettingReposContainer } from './components'
import { defaultFilters, RepositoriesList } from './RepositoriesList'

import styles from './SettingsRepositoriesPage.module.scss'

interface Props extends TelemetryProps {
    owner: Owner
    routingPrefix: string
    onOrgGetStartedRefresh?: () => void
    org?: OrgAreaOrganizationFields
}

type SyncStatusOrError = undefined | 'scheduled' | 'schedule-complete' | ErrorLike

/**
 * A page displaying the repositories for this user.
 */
export const SettingsRepositoriesPage: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    owner,
    routingPrefix,
    telemetryService,
    onOrgGetStartedRefresh,
}) => {
    const [hasRepos, setHasRepos] = useState(false)
    const [externalServices, setExternalServices] = useState<ExternalServicesResult['externalServices']['nodes']>()
    const [repoFilters, setRepoFilters] = useState<FilteredConnectionFilter[]>([])
    const [status, setStatus] = useState<SyncStatusOrError>()
    const [updateReposList, setUpdateReposList] = useState(false)

    const isUserOwner = owner.type === 'user'

    const NoAddedReposBanner = (
        <Container className="text-center">
            <H4>{owner.name ? `${owner.name} has` : 'You have'} not added any repositories to Sourcegraph</H4>

            {externalServices?.length !== 0 ? (
                <span className="text-muted">
                    <Link to={`${routingPrefix}/repositories/manage`}>Add repositories</Link> to start searching code
                    with Sourcegraph.
                </span>
            ) : (
                <span className="text-muted">
                    <Link to={`${routingPrefix}/code-hosts`}>Connect a code host</Link> to add your code to Sourcegraph.{' '}
                    <span>
                        You can also{' '}
                        <Link to={`${routingPrefix}/repositories/manage`}>add individual public repositories</Link> from
                        GitHub.com or GitLab.com.
                    </span>
                </span>
            )}
        </Container>
    )

    const fetchExternalServices = useCallback(
        async (): Promise<ExternalServicesResult['externalServices']['nodes']> =>
            queryExternalServices({
                first: null,
                after: null,
                namespace: owner.id,
            })
                .toPromise()
                .then(({ nodes }) => nodes),

        [owner.id]
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
        const result = await fetchUserRepositoriesCount({
            id: owner.id,
        })
        const repoCount = result.node.repositories.totalCount || 0

        if (repoCount) {
            setHasRepos(true)
        }

        // configure filters
        const specificCodeHostFilters = services.map(service => ({
            tooltip: '',
            value: service.id,
            label: service.displayName.split(' ')[0],
            args: { externalServiceID: service.id },
        }))

        const [statusFilter, codeHostFilter] = defaultFilters

        // update default code host filter by adding GitLab and/or GitHub filters
        const updatedCodeHostFilter = {
            ...codeHostFilter,
            values: [...codeHostFilter.values, ...specificCodeHostFilters],
        }

        setRepoFilters([statusFilter, updatedCodeHostFilter])
    }, [fetchExternalServices, owner.id])

    const TWO_SECONDS = 2

    const queryRepos = useCallback(
        (args: FilteredConnectionQueryArguments): Observable<NonNullable<RepositoriesResult>['repositories']> =>
            listUserRepositories({ ...args, id: owner.id }).pipe(
                tap(() => {
                    if (status === 'schedule-complete') {
                        setUpdateReposList(!updateReposList)
                        setStatus(undefined)
                    }

                    // TODO: @artem - fix the context banner
                    // if (repos.nodes.length !== 0) {
                    //     if (status === 'schedule-complete') {
                    //         setShouldDisplayContextBanner(true)
                    //     }
                    // } else {
                    //     setShouldDisplayContextBanner(false)
                    // }
                })
            ),
        [owner.id, status, updateReposList]
    )

    useObservable(
        useMemo(() => {
            if (externalServices && externalServices.length !== 0) {
                // get serviceIds and check if services will sync in the next 2 seconds
                const serviceIds = externalServices.map(service => service.id)

                return fetchCodeHostSyncDueStatus(serviceIds, TWO_SECONDS).pipe(
                    repeatUntil(
                        result => {
                            const isScheduledToSync = result.data?.codeHostSyncDue === true
                            // if all existing code hosts were just added -
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

    const onRepoQueryUpdate = useCallback(
        (value: Connection<SiteAdminRepositoryFields> | ErrorLike | undefined, query: string): void => {
            if (value as Connection<SiteAdminRepositoryFields>) {
                const conn = value as Connection<SiteAdminRepositoryFields>

                if (onOrgGetStartedRefresh) {
                    onOrgGetStartedRefresh()
                }
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
        [onOrgGetStartedRefresh]
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

            return `Syncing with ${names.join(', ')}.`
        }
        return 'Syncing.'
    }

    return (
        <UserSettingReposContainer>
            <SelfHostedCtaLink
                className={styles.selfHostedCta}
                telemetryService={telemetryService}
                page="settings/repositories"
            />
            {status === 'scheduled' && (
                <Alert variant="info">
                    <span className="font-weight-bold">{getCodeHostsSyncMessage()}</span> Repositories may not be
                    up-to-date and will refresh once sync is finished.
                </Alert>
            )}
            {isErrorLike(status) && <ErrorAlert error={status} icon={true} />}
            <PageTitle title="Your repositories" />
            <PageHeader
                headingElement="h2"
                path={[
                    {
                        text: (
                            <div className="d-flex">
                                Your repositories{' '}
                                <ProductStatusBadge status="beta" className="ml-2" linkToDocs={true} />
                            </div>
                        ),
                    },
                ]}
                description={
                    <span className="text-muted">
                        All repositories synced with Sourcegraph from {owner.name ? owner.name + "'s" : 'your'}{' '}
                        <Link to={`${routingPrefix}/code-hosts`}>connected code hosts</Link>.
                    </span>
                }
                actions={
                    <span>
                        {hasRepos ? (
                            <Button
                                to={`${routingPrefix}/repositories/manage`}
                                onClick={logManageRepositoriesClick}
                                variant="primary"
                                as={Link}
                            >
                                Manage repositories
                            </Button>
                        ) : isUserOwner ? (
                            <Button
                                to={`${routingPrefix}/repositories/manage`}
                                onClick={logManageRepositoriesClick}
                                variant="primary"
                                as={Link}
                            >
                                <Icon aria-hidden={true} svgPath={mdiPlus} /> Add repositories
                            </Button>
                        ) : externalServices && externalServices.length !== 0 ? (
                            <Button
                                to={`${routingPrefix}/repositories/manage`}
                                onClick={logManageRepositoriesClick}
                                variant="primary"
                                as={Link}
                            >
                                <Icon aria-hidden={true} svgPath={mdiPlus} /> Add repositories
                            </Button>
                        ) : (
                            <Button
                                to={`${routingPrefix}/code-hosts`}
                                onClick={logManageRepositoriesClick}
                                variant="primary"
                                as={Link}
                            >
                                <Icon aria-hidden={true} svgPath={mdiPlus} /> Connect code hosts
                            </Button>
                        )}
                    </span>
                }
                className="mb-3"
            />
            {isErrorLike(status) ? (
                <H3 className="text-muted">Sorry, we couldn’t fetch your repositories. Try again?</H3>
            ) : !externalServices ? (
                <div className="d-flex justify-content-center mt-4">
                    <LoadingSpinner />
                </div>
            ) : hasRepos ? (
                <RepositoriesList
                    queryRepos={queryRepos}
                    updateReposList={updateReposList}
                    onRepoQueryUpdate={onRepoQueryUpdate}
                    repoFilters={repoFilters}
                />
            ) : (
                NoAddedReposBanner
            )}
        </UserSettingReposContainer>
    )
}
