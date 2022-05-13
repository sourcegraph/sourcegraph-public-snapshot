import React, { FunctionComponent, useState, useEffect, useCallback, useRef } from 'react'

import classNames from 'classnames'
import { useLocation, useHistory } from 'react-router'

import { ErrorLike } from '@sourcegraph/common'
import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary/useTemporarySetting'
import { TelemetryService } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Alert, Link, Typography } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../auth'
import { BrandLogo } from '../components/branding/BrandLogo'
import { HeroPage } from '../components/HeroPage'
import { PageTitle } from '../components/PageTitle'
import { SourcegraphContext } from '../jscontext'
import { PageRoutes } from '../routes.constants'
import { eventLogger } from '../tracking/eventLogger'
import { SelectAffiliatedRepos } from '../user/settings/repositories/SelectAffiliatedRepos'
import { UserExternalServicesOrRepositoriesUpdateProps } from '../util'

import { getReturnTo } from './SignInSignUpCommon'
import { Steps, Step, StepList, StepPanels, StepPanel } from './Steps'
import { useExternalServices } from './useExternalServices'
import { CodeHostsConnection } from './welcome/CodeHostsConnection'
import { Footer } from './welcome/Footer'
import { InviteCollaborators } from './welcome/InviteCollaborators/InviteCollaborators'

import styles from './PostSignUpPage.module.scss'
import signInSignUpCommonStyles from './SignInSignUpCommon.module.scss'

interface PostSignUpPage {
    authenticatedUser: AuthenticatedUser
    context: Pick<SourcegraphContext, 'authProviders'>
    telemetryService: TelemetryService
    onUserExternalServicesOrRepositoriesUpdate: UserExternalServicesOrRepositoriesUpdateProps['onUserExternalServicesOrRepositoriesUpdate']
    setSelectedSearchContextSpec: (spec: string) => void
}

interface Step {
    content: React.ReactElement
    isComplete: () => boolean
    prefetch?: () => void
    onNextButtonClick?: () => Promise<void>
}

interface FinishEventPayload {
    eventName?: string
    tabNumber?: number
}
export type RepoSelectionMode = 'all' | 'selected' | undefined

export type FinishWelcomeFlow = (event: React.MouseEvent<HTMLElement>, payload: FinishEventPayload) => void

export const getPostSignUpEvent = (action?: string): string => `PostSignUp${action ? '_' + action : ''}`

export const PostSignUpPage: FunctionComponent<React.PropsWithChildren<PostSignUpPage>> = ({
    authenticatedUser: user,
    context,
    telemetryService,
    onUserExternalServicesOrRepositoriesUpdate,
    setSelectedSearchContextSpec,
}) => {
    const [didUserFinishWelcomeFlow, setUserFinishedWelcomeFlow] = useTemporarySetting(
        'signup.finishedWelcomeFlow',
        false
    )

    const isOAuthCall = useRef(false)
    const location = useLocation()
    const history = useHistory()

    const debug = new URLSearchParams(location.search).get('debug')

    const goToSearch = (): void => history.push(getReturnTo(location))

    useEffect(() => {
        eventLogger.logViewEvent(getPostSignUpEvent())
    }, [])

    if (debug && !didUserFinishWelcomeFlow) {
        setUserFinishedWelcomeFlow(false)
    } else if (didUserFinishWelcomeFlow) {
        // if the welcome flow was already finished - navigate to search
        goToSearch()
    }

    const finishWelcomeFlow: FinishWelcomeFlow = (event, { eventName, tabNumber }) => {
        event.currentTarget.blur()
        setUserFinishedWelcomeFlow(true)

        const fullEventName = getPostSignUpEvent(eventName)
        if (tabNumber) {
            eventLogger.log(fullEventName, { tabNumber }, { tabNumber })
        } else {
            eventLogger.log(fullEventName)
        }

        goToSearch()
    }

    const [repoSelectionMode, setRepoSelectionMode] = useState<RepoSelectionMode>()
    const [error, setError] = useState<ErrorLike>()
    const { externalServices, loadingServices, errorServices, refetchExternalServices } = useExternalServices(user.id)

    const hasErrors = error || errorServices

    const beforeUnload = useCallback((): void => {
        // user is not leaving the flow, it's an OAuth page refresh
        if (isOAuthCall.current) {
            return
        }

        eventLogger.log(getPostSignUpEvent('Page_NavigatedAway'))
        setUserFinishedWelcomeFlow(true)
    }, [setUserFinishedWelcomeFlow])

    useEffect(() => {
        if (hasErrors) {
            return
        }

        window.addEventListener('beforeunload', beforeUnload)

        return () => window.removeEventListener('beforeunload', beforeUnload)
    }, [beforeUnload, error, hasErrors])

    const onError = useCallback((error: ErrorLike) => setError(error), [])

    return (
        <div className={styles.wrapper}>
            <BrandLogo className={styles.logo} isLightTheme={true} variant="symbol" />
            <div className={classNames(signInSignUpCommonStyles.signinSignupPage, styles.postSignupPage)}>
                <PageTitle title="Welcome" />
                <HeroPage
                    lessPadding={false}
                    className="text-left"
                    body={
                        <div className="pb-1 d-flex flex-column align-items-center w-100">
                            <div className={styles.container}>
                                {hasErrors && (
                                    <Alert className="mb-4" role="alert" variant="danger">
                                        Sorry, something went wrong. Try refreshing the page or{' '}
                                        <Link to={PageRoutes.Search}>skip to code search</Link>.
                                    </Alert>
                                )}
                                <Typography.H2>Get started with Sourcegraph</Typography.H2>
                                <p className="text-muted pb-3">Follow these steps to set up your account</p>
                            </div>
                            <div className="mt-2 pb-3 d-flex flex-column align-items-center w-100">
                                <Steps initialStep={debug ? parseInt(debug, 10) : 1} totalSteps={3}>
                                    <StepList numeric={true} className={styles.container}>
                                        <Step borderColor="purple">Connect with code hosts</Step>
                                        <Step borderColor="blue">Add repositories</Step>
                                        <Step borderColor="orange">Invite collaborators</Step>
                                    </StepList>
                                    <StepPanels>
                                        <StepPanel>
                                            <div className={styles.container}>
                                                <CodeHostsConnection
                                                    user={user}
                                                    onNavigation={(called: boolean) => {
                                                        isOAuthCall.current = called
                                                    }}
                                                    loading={loadingServices}
                                                    onError={onError}
                                                    externalServices={externalServices}
                                                    context={context}
                                                    refetch={refetchExternalServices}
                                                />
                                                <Footer onFinish={finishWelcomeFlow} />
                                            </div>
                                        </StepPanel>
                                        <StepPanel>
                                            <div className={classNames('mt-3', styles.container)}>
                                                <Typography.H3>Add repositories</Typography.H3>
                                                <p className="text-muted mb-4">
                                                    Choose repositories you own or collaborate on from your code hosts.
                                                    We’ll sync and index these repositories so you can search your code
                                                    all in one place.
                                                    <Link
                                                        to="https://docs.sourcegraph.com/code_search/how-to/adding_repositories_to_cloud"
                                                        target="_blank"
                                                        rel="noopener noreferrer"
                                                    >
                                                        {' '}
                                                        Learn more
                                                    </Link>
                                                </p>
                                                <SelectAffiliatedRepos
                                                    authenticatedUser={user}
                                                    onRepoSelectionModeChange={setRepoSelectionMode}
                                                    repoSelectionMode={repoSelectionMode}
                                                    telemetryService={telemetryService}
                                                    onError={onError}
                                                />
                                                <Footer onFinish={finishWelcomeFlow} isSkippable={true} />
                                            </div>
                                        </StepPanel>
                                        <StepPanel>
                                            <InviteCollaborators
                                                className={styles.container}
                                                user={user}
                                                repoSelectionMode={repoSelectionMode}
                                                onUserExternalServicesOrRepositoriesUpdate={
                                                    onUserExternalServicesOrRepositoriesUpdate
                                                }
                                                setSelectedSearchContextSpec={setSelectedSearchContextSpec}
                                                onError={onError}
                                                onFinish={finishWelcomeFlow}
                                            />
                                        </StepPanel>
                                    </StepPanels>
                                </Steps>
                            </div>
                        </div>
                    }
                />
            </div>
        </div>
    )
}
