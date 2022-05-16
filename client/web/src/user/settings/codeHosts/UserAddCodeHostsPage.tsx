import React, { useCallback, useState, useEffect } from 'react'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { asError, ErrorLike, isErrorLike, isDefined, keyExistsIn } from '@sourcegraph/common'
import { useQuery } from '@sourcegraph/http-client'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Button, Container, PageHeader, LoadingSpinner, Link, Alert, Typography } from '@sourcegraph/wildcard'

import { queryExternalServices } from '../../../components/externalServices/backend'
import { AddExternalServiceOptions } from '../../../components/externalServices/externalServices'
import { PageTitle } from '../../../components/PageTitle'
import { SelfHostedCta } from '../../../components/SelfHostedCta'
import { useFlagsOverrides } from '../../../featureFlags/featureFlags'
import {
    ExternalServiceKind,
    ListExternalServiceFields,
    OrgFeatureFlagValueResult,
    OrgFeatureFlagValueVariables,
} from '../../../graphql-operations'
import { AuthProvider, SourcegraphContext } from '../../../jscontext'
import { GET_ORG_FEATURE_FLAG_VALUE, GITHUB_APP_FEATURE_FLAG_NAME } from '../../../org/backend'
import { useCodeHostScopeContext } from '../../../site/CodeHostScopeAlerts/CodeHostScopeProvider'
import { eventLogger } from '../../../tracking/eventLogger'
import { UserExternalServicesOrRepositoriesUpdateProps } from '../../../util'
import { githubRepoScopeRequired, gitlabAPIScopeRequired, Owner } from '../cloud-ga'

import { CodeHostItem, ParentWindow } from './CodeHostItem'
import { CodeHostListItem } from './CodeHostListItem'

type AuthProvidersByKind = Partial<Record<ExternalServiceKind, AuthProvider>>

export interface UserAddCodeHostsPageProps
    extends Pick<UserExternalServicesOrRepositoriesUpdateProps, 'onUserExternalServicesOrRepositoriesUpdate'>,
        TelemetryProps {
    owner: Owner
    codeHostExternalServices: Record<string, AddExternalServiceOptions>
    routingPrefix: string
    context: Pick<SourcegraphContext, 'authProviders'>
    onOrgGetStartedRefresh?: () => void
}

type ServicesByKind = Partial<Record<ExternalServiceKind, ListExternalServiceFields>>
type Status = undefined | 'loading' | ServicesByKind | ErrorLike

const isServicesByKind = (status: Status): status is ServicesByKind =>
    typeof status === 'object' && Object.keys(status).every(key => keyExistsIn(key, ExternalServiceKind))

export const updateGitHubApp = (event?: { preventDefault(): void }): void => {
    if (event) {
        event.preventDefault()
    }
    window.location.assign(
        `/.auth/github/login?pc=${encodeURIComponent(
            `https://github.com/::${window.context.githubAppCloudClientID}`
        )}&op=createCodeHostConnection&redirect=${window.location.href}`
    )
}

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

export interface ServiceConfig {
    pending: boolean
}

const checkGithubOutage = async (): Promise<boolean> => {
    let status = ''
    await fetch('https://www.githubstatus.com/api/v2/status.json', {
        method: 'GET',
    })
        .then(response => response.json())
        .then(response => (status = response.status.indicator))

    if (status === 'major' || status === 'partial') {
        return true
    }

    return false
}

export const UserAddCodeHostsPage: React.FunctionComponent<React.PropsWithChildren<UserAddCodeHostsPageProps>> = ({
    owner,
    codeHostExternalServices,
    routingPrefix,
    context,
    onUserExternalServicesOrRepositoriesUpdate,
    telemetryService,
    onOrgGetStartedRefresh,
}) => {
    if (window.opener) {
        const parentWindow: ParentWindow = window.opener as ParentWindow
        if (parentWindow.onSuccess) {
            const urlParameters = new URLSearchParams(window.location.search)
            const reason = urlParameters.get('reason')

            parentWindow.onSuccess(reason)
        }
        window.close()
    }
    const [statusOrError, setStatusOrError] = useState<Status>()
    const { scopes, setScope } = useCodeHostScopeContext()
    const codeHostModalRecord: Record<string, boolean> = Object.fromEntries(
        Object.entries(codeHostExternalServices).map(([id_, { kind }]) => [kind, false])
    )
    const [isUpdateModalOpen, setIsUpdateModalOpen] = useState<Record<string, boolean>>(codeHostModalRecord)
    const toggleUpdateModal = (kind: string) => (): void => {
        setIsUpdateModalOpen(modalState => {
            const newModalState = { ...modalState } // You have to create a new object otherwise React won't register the state changed
            newModalState[kind] = !modalState[kind]
            return newModalState
        })
    }
    const [servicesDown, setServicesDown] = useState<string[]>()

    const { data, loading } = useQuery<OrgFeatureFlagValueResult, OrgFeatureFlagValueVariables>(
        GET_ORG_FEATURE_FLAG_VALUE,
        {
            variables: { orgID: owner.id, flagName: GITHUB_APP_FEATURE_FLAG_NAME },
            // Cache this data but always re-request it in the background when we revisit
            // this page to pick up newer changes.
            fetchPolicy: 'cache-and-network',
            skip: !(owner.type === 'org'),
        }
    )

    const useGitHubApp = data?.organizationFeatureFlagValue || false

    const flagsOverridesResult = useFlagsOverrides()
    const isGitHubAppEnabled = flagsOverridesResult.data
        ?.filter(orgFlag => orgFlag.flagName === GITHUB_APP_FEATURE_FLAG_NAME)
        .some(orgFlag => orgFlag.value)
    const isGitHubAppLoading = flagsOverridesResult.loading

    // If we have a GitHub or GitLab services, check whether we need to prompt the user to
    // update their scope
    const isGitHubTokenUpdateRequired = scopes.github
        ? !isGitHubAppEnabled && githubRepoScopeRequired(owner.tags, scopes.github)
        : false
    const isGitLabTokenUpdateRequired = scopes.gitlab ? gitlabAPIScopeRequired(owner.tags, scopes.gitlab) : false

    const isTokenUpdateRequired: Partial<Record<ExternalServiceKind, boolean | undefined>> = {
        [ExternalServiceKind.GITHUB]: githubRepoScopeRequired(owner.tags, scopes.github),
        [ExternalServiceKind.GITLAB]: gitlabAPIScopeRequired(owner.tags, scopes.gitlab),
    }

    useEffect(() => {
        eventLogger.logPageView('UserSettingsCodeHostConnections')
    }, [])

    async function checkAndSetOutageAlert(
        services: Partial<Record<ExternalServiceKind, ListExternalServiceFields>>
    ): Promise<void> {
        const svcs = []
        for (const svc of Object.values(services)) {
            // When there is a sync error, check for potential outages by calling GitHub Status API
            if (svc.displayName === 'GitHub' && svc.lastSyncError !== null) {
                const outage = await checkGithubOutage()
                if (outage) {
                    svcs.push(svc.displayName)
                }
            }
            // GitLab doesn't have a Status API, so check if the error contains a Status Code of 500 or 503
            if (
                (svc.displayName === 'GitLab' && svc.lastSyncError?.includes('500')) ||
                svc.lastSyncError?.includes('503')
            ) {
                svcs.push(svc.displayName)
            }
        }

        setServicesDown(svcs)
        return
    }

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

        await checkAndSetOutageAlert(services)

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

        if (onOrgGetStartedRefresh) {
            onOrgGetStartedRefresh()
        }
    }

    useEffect(() => {
        fetchExternalServices().catch(error => {
            setStatusOrError(asError(error))
        })
    }, [fetchExternalServices])

    const refetchServices = useCallback((): void => {
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

    const getRequestSuccessBanner = (service: ListExternalServiceFields | undefined): JSX.Element | null => {
        if (!service) {
            return null
        }
        interface ServiceConfig {
            pending: boolean
        }
        const serviceConfig = JSON.parse(service.config) as ServiceConfig

        if (serviceConfig.pending) {
            return (
                <Alert className="mb-4" role="alert" key="update-gitlab" variant="info">
                    <Typography.H4>GitHub code host connection pending</Typography.H4>
                    An installation request was sent to your GitHub organization’s owners. After the request is
                    approved, finish connecting with GitHub to choose repositories to sync with Sourcegraph.
                </Alert>
            )
        }

        return null
    }

    const getGitHubUpdateAuthBanner = (needsUpdate: boolean): JSX.Element | null =>
        needsUpdate ? (
            <Alert className="mb-4" role="alert" key="update-github" variant="info">
                Update your GitHub code host connection to search private code with Sourcegraph.
            </Alert>
        ) : null

    const getGitLabUpdateAuthBanner = (needsUpdate: boolean): JSX.Element | null =>
        needsUpdate ? (
            <Alert className="mb-4" role="alert" key="update-gitlab" variant="info">
                Update your GitLab code host connection to search private code with Sourcegraph.
            </Alert>
        ) : null

    const getAddReposBanner = (services: string[]): JSX.Element | null =>
        services.length > 0 ? (
            <Alert className="my-3" role="alert" key="add-repos" variant="success">
                <Typography.H4 className="align-middle mb-1">Connected with {services.join(', ')}</Typography.H4>
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
            </Alert>
        ) : null

    interface serviceProblem {
        id: string
        kind: string
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
                let outage = false

                // Skip when status code >= 500, as they are handled by the outage checkers. This will avoid creating duplicate alert messages.
                if (
                    service.lastSyncError &&
                    (service.lastSyncError?.includes('503') || service.lastSyncError?.includes('500'))
                ) {
                    outage = true
                }

                // if service has warnings or errors
                if (problem && !outage) {
                    servicesWithProblems.push({
                        id: service.id,
                        kind: service.kind,
                        displayName: service.displayName,
                        problem,
                    })
                    continue
                }

                const serviceConfig = JSON.parse(service.config) as ServiceConfig

                if (serviceConfig.pending) {
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
            getRequestSuccessBanner(
                isServicesByKind(statusOrError) ? statusOrError[ExternalServiceKind.GITHUB] : undefined
            ),
            getGitHubUpdateAuthBanner(isGitHubTokenUpdateRequired),
            getGitLabUpdateAuthBanner(isGitLabTokenUpdateRequired),
        ]
    }

    const addNewService = useCallback(
        (service: ListExternalServiceFields): void => {
            if (isServicesByKind(statusOrError)) {
                setStatusOrError({ ...statusOrError, [service.kind]: service })
                if (onOrgGetStartedRefresh) {
                    onOrgGetStartedRefresh()
                }
            }
        },
        [statusOrError, onOrgGetStartedRefresh]
    )

    const handleError = useCallback((error: ErrorLike): void => setStatusOrError(error), [])

    const getServiceWarningFragment = (service: serviceProblem): JSX.Element => (
        <Alert className="my-3" key={service.id} variant="warning">
            <Typography.H4 className="align-middle mb-1">Can’t connect with {service.displayName}</Typography.H4>
            <p className="align-middle mb-0">
                <span className="align-middle">Please try</span>{' '}
                {owner.type === 'org' ? (
                    <Button
                        className="font-weight-normal shadow-none p-0 border-0"
                        onClick={toggleUpdateModal(service.kind)}
                        variant="link"
                    >
                        updating the code host connection
                    </Button>
                ) : (
                    <span className="align-middle">reconnecting the code host connection</span>
                )}{' '}
                <span className="align-middle">with {service.displayName} to restore access.</span>
            </p>
        </Alert>
    )

    const getOutageMessage = (servicesDown: string[]): JSX.Element => (
        <Alert className="my-3" key={servicesDown[0]} variant="warning">
            {servicesDown?.map(svc => (
                <div key={svc}>
                    <Typography.H4 className="align-middle mb-1">
                        We’re having trouble connecting to {svc}{' '}
                    </Typography.H4>
                    <p className="align-middle mb-0">
                        <span className="align-middle">Verify that</span> {svc}
                        <span className="align-middle">
                            {' '}
                            is available by visiting{' '}
                            {svc === 'GitHub' ? (
                                <Link to="https://githubstatus.com" target="_blank" rel="noopener">
                                    githubstatus.com
                                </Link>
                            ) : (
                                <Link to="https://status.gitlab.com" target="_blank" rel="noopener">
                                    status.gitlab.com
                                </Link>
                            )}
                        </span>{' '}
                    </p>
                </div>
            ))}
        </Alert>
    )

    // auth providers by service type
    const authProvidersByKind = context.authProviders.reduce((accumulator: AuthProvidersByKind, provider) => {
        if (provider.authenticationURL) {
            accumulator[provider.serviceType.toLocaleUpperCase() as ExternalServiceKind] = provider
        }
        return accumulator
    }, {})

    const defaultNavigateToAuthProvider = useCallback(
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

    const navigateToAuthProvider = useCallback(
        (kind: ExternalServiceKind): void => {
            const authProvider = authProvidersByKind[kind]

            if (authProvider) {
                eventLogger.log('ConnectUserCodeHostClicked', { kind }, { kind })

                if (kind !== ExternalServiceKind.GITHUB || !isGitHubAppEnabled) {
                    defaultNavigateToAuthProvider(kind)
                } else if (owner.type === 'org') {
                    const secondRedirectURI = `/.auth/github/install-github-app?state=${encodeURIComponent(owner.id)}`
                    const firstRedirectURI = `/.auth/github/login?pc=${encodeURIComponent(
                        `https://github.com/::${window.context.githubAppCloudClientID}`
                    )}&op=createCodeHostConnection&redirect=${encodeURIComponent(secondRedirectURI)}`

                    const browser: ParentWindow = window.self as ParentWindow

                    browser.onSuccess = () => {
                        refetchServices()
                    }
                    const popup = browser.open(
                        `${authProvider.authenticationURL as string}&redirect=${encodeURIComponent(firstRedirectURI)}`,
                        'name',
                        `dependent=${1}, alwaysOnTop=${1}, alwaysRaised=${1}, alwaysRaised=${1}, width=${600}, height=${900}`
                    )

                    const popupTick = setInterval(() => {
                        if (popup?.closed) {
                            clearInterval(popupTick)
                        }
                    }, 500)
                } else {
                    updateGitHubApp()
                }
            }
        },
        [authProvidersByKind, defaultNavigateToAuthProvider, isGitHubAppEnabled, owner, refetchServices]
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
            {/* display outage alert when a service is experiencing an outage */}
            {servicesDown && servicesDown.length > 0 && getOutageMessage(servicesDown)}
            {codeHostExternalServices && isServicesByKind(statusOrError) ? (
                <Container>
                    <ul className="list-group">
                        {Object.entries(codeHostExternalServices).map(([id, { kind, defaultDisplayName, icon }]) =>
                            authProvidersByKind[kind] ? (
                                <CodeHostListItem key={id}>
                                    <CodeHostItem
                                        owner={owner}
                                        service={isServicesByKind(statusOrError) ? statusOrError[kind] : undefined}
                                        kind={kind}
                                        name={defaultDisplayName}
                                        isTokenUpdateRequired={
                                            isTokenUpdateRequired[kind] &&
                                            !(kind === ExternalServiceKind.GITHUB && isGitHubAppEnabled)
                                        }
                                        navigateToAuthProvider={navigateToAuthProvider}
                                        icon={icon}
                                        isUpdateModalOpen={isUpdateModalOpen[kind]}
                                        toggleUpdateModal={toggleUpdateModal(kind)}
                                        onDidUpsert={handleServiceUpsert}
                                        onDidAdd={addNewService}
                                        onDidRemove={removeService(kind)}
                                        onDidError={handleError}
                                        loading={kind === ExternalServiceKind.GITHUB && loading && isGitHubAppLoading}
                                        useGitHubApp={kind === ExternalServiceKind.GITHUB && useGitHubApp}
                                    />
                                </CodeHostListItem>
                            ) : null
                        )}
                    </ul>
                </Container>
            ) : (
                <div className="d-flex justify-content-center">
                    <LoadingSpinner />
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
