import React, { useCallback, useState, useEffect } from 'react'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { Link } from '@sourcegraph/shared/src/components/Link'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { asError, ErrorLike, isErrorLike } from '@sourcegraph/shared/src/util/errors'
import { isDefined, keyExistsIn } from '@sourcegraph/shared/src/util/types'
import { SelfHostedCta } from '@sourcegraph/web/src/components/SelfHostedCta'
import { Container, PageHeader } from '@sourcegraph/wildcard'

import { ErrorAlert } from '../../../components/alerts'
import { queryExternalServices } from '../../../components/externalServices/backend'
import { AddExternalServiceOptions } from '../../../components/externalServices/externalServices'
import { PageTitle } from '../../../components/PageTitle'
import { ExternalServiceKind, ListExternalServiceFields } from '../../../graphql-operations'
import { AuthProvider, SourcegraphContext } from '../../../jscontext'
import { useCodeHostScopeContext } from '../../../site/CodeHostScopeAlerts/CodeHostScopeProvider'
import { eventLogger } from '../../../tracking/eventLogger'
import { UserExternalServicesOrRepositoriesUpdateProps } from '../../../util'
import { githubRepoScopeRequired, gitlabAPIScopeRequired, Owner } from '../cloud-ga'

import { CodeHostItem } from './CodeHostItem'

type AuthProvidersByKind = Partial<Record<ExternalServiceKind, AuthProvider>>

export interface UserAddCodeHostsPageProps
    extends Pick<UserExternalServicesOrRepositoriesUpdateProps, 'onUserExternalServicesOrRepositoriesUpdate'>,
        TelemetryProps {
    owner: Owner
    codeHostExternalServices: Record<string, AddExternalServiceOptions>
    routingPrefix: string
    context: Pick<SourcegraphContext, 'authProviders'>
}

type ServicesByKind = Partial<Record<ExternalServiceKind, ListExternalServiceFields>>
type Status = undefined | 'loading' | ServicesByKind | ErrorLike

const isServicesByKind = (status: Status): status is ServicesByKind =>
    typeof status === 'object' && Object.keys(status).every(key => keyExistsIn(key, ExternalServiceKind))

export const ifNotNavigated = (callback: () => void, waitMS: number = 2000): void => {
    let timeoutID = 0
    let willNavigate = false

    const unloadListener = (): void => {
        willNavigate = true
    }

    window.addEventListener('unload', unloadListener)

    timeoutID = window.setTimeout(() => {
        // if we waited waitMS and the navigation didn't happen - run the callback
        if (!willNavigate) {
            // cleanup
            window.removeEventListener('unload', unloadListener)
            window.clearTimeout(timeoutID)

            return callback()
        }
    }, waitMS)
}

export const UserAddCodeHostsPage: React.FunctionComponent<UserAddCodeHostsPageProps> = ({
    owner,
    codeHostExternalServices,
    routingPrefix,
    context,
    onUserExternalServicesOrRepositoriesUpdate,
    telemetryService,
}) => {
    const [statusOrError, setStatusOrError] = useState<Status>()
    const { scopes, setScope } = useCodeHostScopeContext()
    const [isUpdateModalOpen, setIssUpdateModalOpen] = useState(false)
    const toggleUpdateModal = useCallback(() => {
        setIssUpdateModalOpen(!isUpdateModalOpen)
    }, [isUpdateModalOpen])

    // If we have a GitHub or GitLab services, check whether we need to prompt the user to
    // update their scope
    const isGitHubTokenUpdateRequired = scopes.github ? githubRepoScopeRequired(owner.tags, scopes.github) : false
    const isGitLabTokenUpdateRequired = scopes.gitlab ? gitlabAPIScopeRequired(owner.tags, scopes.gitlab) : false

    const isTokenUpdateRequired: Partial<Record<ExternalServiceKind, boolean | undefined>> = {
        [ExternalServiceKind.GITHUB]: githubRepoScopeRequired(owner.tags, scopes.github),
        [ExternalServiceKind.GITLAB]: gitlabAPIScopeRequired(owner.tags, scopes.gitlab),
    }

    useEffect(() => {
        eventLogger.logViewEvent('UserSettingsCodeHostConnections')
    }, [])

    const fetchExternalServices = useCallback(async () => {
        setStatusOrError('loading')

        const { nodes: fetchedServices } = await queryExternalServices({
            namespace: owner.id,
            first: null,
            after: null,
        }).toPromise()

        const services: ServicesByKind = fetchedServices.reduce<ServicesByKind>((accumulator, service) => {
            // backend constraint - non-admin users have only one external service per ExternalServiceKind
            accumulator[service.kind] = service
            return accumulator
        }, {})

        setStatusOrError(services)

        const repoCount = fetchedServices.reduce((sum, codeHost) => sum + codeHost.repoCount, 0)
        onUserExternalServicesOrRepositoriesUpdate(fetchedServices.length, repoCount)
    }, [owner.id, onUserExternalServicesOrRepositoriesUpdate])

    const handleServiceUpsert = useCallback(
        (service: ListExternalServiceFields): void => {
            if (isServicesByKind(statusOrError)) {
                setStatusOrError({ ...statusOrError, [service.kind]: service })
            }
        },
        [statusOrError]
    )

    const removeService = (kind: ExternalServiceKind) => (): void => {
        if (
            (kind === ExternalServiceKind.GITLAB || kind === ExternalServiceKind.GITHUB) &&
            isTokenUpdateRequired[kind]
        ) {
            setScope(kind, null)
        }

        fetchExternalServices().catch(error => {
            setStatusOrError(asError(error))
        })
    }

    useEffect(() => {
        fetchExternalServices().catch(error => {
            setStatusOrError(asError(error))
        })
    }, [fetchExternalServices])

    const logAddRepositoriesClicked = useCallback(
        (source: string) => () => {
            eventLogger.log('UserSettingsAddRepositoriesCTAClicked', null, { source })
        },
        []
    )

    const getGitHubUpdateAuthBanner = (needsUpdate: boolean): JSX.Element | null =>
        needsUpdate ? (
            <div className="alert alert-info mb-4" role="alert" key="update-github">
                Update your GitHub code host connection to search private code with Sourcegraph.
            </div>
        ) : null

    const getGitLabUpdateAuthBanner = (needsUpdate: boolean): JSX.Element | null =>
        needsUpdate ? (
            <div className="alert alert-info mb-4" role="alert" key="update-gitlab">
                Update your GitLab code host connection to search private code with Sourcegraph.
            </div>
        ) : null

    const getAddReposBanner = (services: string[]): JSX.Element | null =>
        services.length > 0 ? (
            <div className="alert alert-success my-3" role="alert" key="add-repos">
                <h4 className="align-middle mb-1">Connected with {services.join(', ')}</h4>
                <p className="align-middle mb-0">
                    Next,{' '}
                    <Link
                        className="font-weight-normal"
                        to={`${routingPrefix}/repositories/manage`}
                        onClick={logAddRepositoriesClicked('banner')}
                    >
                        add repositories
                    </Link>{' '}
                    to search with Sourcegraph.
                </p>
            </div>
        ) : null

    interface serviceProblem {
        id: string
        displayName: string
        problem: string
    }

    const getErrorAndSuccessBanners = (status: Status): (JSX.Element | null)[] => {
        const servicesWithProblems: serviceProblem[] = []
        const notYetSyncedServiceNames = []

        // check if services are fetched
        if (isServicesByKind(status)) {
            const services = Object.values(status).filter(isDefined)

            for (const service of services) {
                const problem = service.warning || service.lastSyncError
                // if service has warnings or errors
                if (problem) {
                    servicesWithProblems.push({ id: service.id, displayName: service.displayName, problem })
                    continue
                }

                // if service is not synced yet or has a "sync now" timestamp
                // "sync now" timestamp is always less then the epoch time

                // don't display user name in service name
                const serviceName = service.displayName.split(' ')[0]

                if (!service?.lastSyncAt) {
                    notYetSyncedServiceNames.push(serviceName)
                } else {
                    const lastSyncTime = new Date(service.lastSyncAt)
                    const epochTime = new Date(0)

                    if (lastSyncTime < epochTime) {
                        notYetSyncedServiceNames.push(serviceName)
                    }
                }
            }
        }

        return [
            ...servicesWithProblems.map(getServiceWarningFragment),
            getAddReposBanner(notYetSyncedServiceNames),
            getGitHubUpdateAuthBanner(isGitHubTokenUpdateRequired),
            getGitLabUpdateAuthBanner(isGitLabTokenUpdateRequired),
        ]
    }

    const addNewService = useCallback(
        (service: ListExternalServiceFields): void => {
            if (isServicesByKind(statusOrError)) {
                setStatusOrError({ ...statusOrError, [service.kind]: service })
            }
        },
        [statusOrError]
    )

    const handleError = useCallback((error: ErrorLike): void => setStatusOrError(error), [])

    const getServiceWarningFragment = (service: serviceProblem): JSX.Element => (
        <div className="alert alert-warning my-3" key={service.id}>
            <h4 className="align-middle mb-1">Can't connect with {service.displayName}</h4>
            <p className="align-middle mb-0">
                <span className="align-middle">Please try</span>{' '}
                {owner.type === 'org' ? (
                    <button
                        type="button"
                        className="btn btn-link font-weight-normal shadow-none p-0 border-0"
                        onClick={toggleUpdateModal}
                    >
                        updating the code host connection
                    </button>
                ) : (
                    <span className="align-middle">reconnecting the code host connection</span>
                )}{' '}
                <span className="align-middle">with {service.displayName} to restore access.</span>
            </p>
        </div>
    )

    // auth providers by service type
    const authProvidersByKind = context.authProviders.reduce((accumulator: AuthProvidersByKind, provider) => {
        if (provider.authenticationURL) {
            accumulator[provider.serviceType.toLocaleUpperCase() as ExternalServiceKind] = provider
        }
        return accumulator
    }, {})

    const navigateToAuthProvider = useCallback(
        (kind: ExternalServiceKind): void => {
            const authProvider = authProvidersByKind[kind]

            if (authProvider) {
                eventLogger.log('ConnectUserCodeHostClicked', { kind }, { kind })
                window.location.assign(
                    `${authProvider.authenticationURL as string}&redirect=${
                        window.location.href
                    }&op=createCodeHostConnection`
                )
            }
        },
        [authProvidersByKind]
    )

    return (
        <div className="user-code-hosts-page">
            <PageTitle title="Code host connections" />
            <PageHeader
                headingElement="h2"
                path={[{ text: 'Code host connections' }]}
                description={
                    <span className="text-muted">
                        Connect with {owner.name ? owner.name + "'s" : 'your'} code hosts. Then,{' '}
                        <Link
                            to={`${routingPrefix}/repositories/manage`}
                            onClick={logAddRepositoriesClicked('description')}
                        >
                            add repositories
                        </Link>{' '}
                        to search with Sourcegraph.
                    </span>
                }
                className="mb-3"
            />
            {/* display external service errors and success banners */}
            {getErrorAndSuccessBanners(statusOrError)}
            {/* display other errors, e.g. network errors */}
            {isErrorLike(statusOrError) && (
                <ErrorAlert error={statusOrError} prefix="Code host action error" icon={false} />
            )}
            {codeHostExternalServices && isServicesByKind(statusOrError) ? (
                <Container>
                    <ul className="list-group">
                        {Object.entries(codeHostExternalServices).map(([id, { kind, defaultDisplayName, icon }]) =>
                            authProvidersByKind[kind] ? (
                                <li key={id} className="list-group-item user-code-hosts-page__code-host-item">
                                    <CodeHostItem
                                        owner={owner}
                                        service={isServicesByKind(statusOrError) ? statusOrError[kind] : undefined}
                                        kind={kind}
                                        name={defaultDisplayName}
                                        isTokenUpdateRequired={isTokenUpdateRequired[kind]}
                                        navigateToAuthProvider={navigateToAuthProvider}
                                        icon={icon}
                                        isUpdateModalOpen={isUpdateModalOpen}
                                        toggleUpdateModal={toggleUpdateModal}
                                        onDidUpsert={handleServiceUpsert}
                                        onDidAdd={addNewService}
                                        onDidRemove={removeService(kind)}
                                        onDidError={handleError}
                                    />
                                </li>
                            ) : null
                        )}
                    </ul>
                </Container>
            ) : (
                <div className="d-flex justify-content-center">
                    <LoadingSpinner className="icon-inline" />
                </div>
            )}

            <SelfHostedCta className="mt-5" page="settings/code-hosts" telemetryService={telemetryService}>
                <p className="mb-2">
                    <strong>Require support for Bitbucket, or nearly any other code host?</strong>
                </p>
                <p className="mb-2">You may need our self-hosted installation.</p>
            </SelfHostedCta>
        </div>
    )
}
