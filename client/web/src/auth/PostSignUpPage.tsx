import React, { useState } from 'react'
import { useLocation, useHistory } from 'react-router'

import { LinkOrSpan } from '@sourcegraph/shared/src/components/LinkOrSpan'
import { BrandLogo } from '@sourcegraph/web/src/components/branding/BrandLogo'
import { Steps, Step, StepList, StepPanels, StepPanel, useSteps } from '@sourcegraph/wildcard/src/components/Steps'

import { HeroPage } from '../components/HeroPage'
import { PageTitle } from '../components/PageTitle'
import { UserAreaUserFields } from '../graphql-operations'
import { SourcegraphContext } from '../jscontext'

import { getReturnTo } from './SignInSignUpCommon'
import { useExternalServices } from './useExternalServices'
import { useRepoCloningStatus } from './useRepoCloningStatus'
import { CodeHostsConnection } from './welcome/CodeHostsConnection'
import { StartSearching } from './welcome/StartSearching'

interface PostSignUpPage {
    authenticatedUser: UserAreaUserFields
    context: Pick<SourcegraphContext, 'authProviders' | 'experimentalFeatures' | 'sourcegraphDotComMode'>
}

interface Step {
    content: React.ReactElement
    isComplete: () => boolean
    prefetch?: () => void
}

export const PostSignUpPage: React.FunctionComponent<PostSignUpPage> = ({ authenticatedUser: user, context }) => {
    const [currentStepNumber, setCurrentStepNumber] = useState(1)
    const toLog = useSteps()
    const location = useLocation()
    const history = useHistory()

    console.log('useSteps =>>>>', toLog)

    const {
        trigger: fetchCloningStatus,
        repos: cloningStatusLines,
        loading: cloningStatusLoading,
        isDoneCloning,
    } = useRepoCloningStatus({ userId: user.id, pollInterval: 2000 })

    const { externalServices, loadingServices, errorServices, refetchExternalServices } = useExternalServices(user.id)

    /**
     * post sign-up flow is available only for .com and only in two cases, user:
     * 1. is authenticated and has AllowUserViewPostSignup tag
     * 2. is authenticated and enablePostSignupFlow experimental feature is ON
     */

    // if (
    //     !user ||
    //     !context.sourcegraphDotComMode ||
    //     !context.experimentalFeatures?.enablePostSignupFlow ||
    //     !user?.tags.includes('AllowUserViewPostSignup')
    // ) {
    //     // TODO: do this on the backend
    //     history.push(getReturnTo(location))
    // }

    const firstStep = {
        content: (
            <>
                {currentStepNumber === 1 && externalServices && (
                    <CodeHostsConnection
                        loading={loadingServices}
                        user={user}
                        error={errorServices}
                        externalServices={externalServices}
                        context={context}
                        refetch={refetchExternalServices}
                    />
                )}
            </>
        ),
        // step is considered complete when user has at least one external service connected.
        isComplete: (): boolean => !!externalServices && externalServices?.length > 0,
    }

    const secondStep = {
        content: (
            <>
                {currentStepNumber === 2 && (
                    <>
                        <h3>Add repositories</h3>
                        <p className="text-muted">
                            Choose repositories you own or collaborate on from your code hosts to search with
                            Sourcegraph. We’ll sync and index these repositories so you can search your code all in one
                            place.
                        </p>
                    </>
                )}
            </>
        ),
        isComplete: () => true,
    }

    const thirdStep = {
        content: (
            <>
                {currentStepNumber === 3 && (
                    <StartSearching
                        isDoneCloning={isDoneCloning}
                        cloningStatusLines={cloningStatusLines}
                        cloningStatusLoading={cloningStatusLoading}
                    />
                )}
            </>
        ),
        isComplete: () => false,
        prefetch: fetchCloningStatus,
    }

    const steps: Step[] = [firstStep, secondStep, thirdStep]

    // Steps helpers
    const isLastStep = currentStepNumber === steps.length
    const currentStep = steps[currentStepNumber - 1]

    const goToNextTab = (): void => {
        // currentStepNumber is not zero based, it'll get the next step
        const nextStep = steps[currentStepNumber]
        if (nextStep.prefetch) {
            nextStep.prefetch()
        }

        setCurrentStepNumber(currentStepNumber + 1)
    }
    const goToSearch = (): void => history.push(getReturnTo(location))
    const isCurrentStepComplete = (): boolean => currentStep?.isComplete()
    const skipPostSignup = (): void => history.push(getReturnTo(location))

    const onStepTabClick = (clickedStepTabNumber: number): void => {
        /**
         * User can navigate through the steps by clicking the step's tab when:
         * 1. navigating back
         * 2. navigating one step forward when the current step is complete
         * 3. navigating many steps forward when all of the steps, from the
         * current one to the clickedStepTabNumber step but not including are
         * complete.
         */

        // do nothing for the current tab
        if (clickedStepTabNumber === currentStepNumber) {
            return
        }

        if (clickedStepTabNumber < currentStepNumber) {
            // allow to navigate back since all of the previous steps had to be completed
            setCurrentStepNumber(clickedStepTabNumber)
        } else if (currentStepNumber - 1 === clickedStepTabNumber) {
            // forward navigation

            // if navigating to the next tab, check if the current step is completed

            if (isCurrentStepComplete()) {
                setCurrentStepNumber(clickedStepTabNumber)
            }
        } else {
            // if navigating further away check [current, ..., clicked)
            const areInBetweenStepsComplete = steps
                .slice(currentStepNumber - 1, clickedStepTabNumber - 1)
                .every(step => step.isComplete())

            if (areInBetweenStepsComplete) {
                setCurrentStepNumber(clickedStepTabNumber)
            }
        }
    }

    return (
        <>
            <LinkOrSpan to={getReturnTo(location)} className="post-signup-page__logo-link">
                <BrandLogo
                    className="position-absolute ml-3 mt-3 post-signup-page__logo"
                    isLightTheme={true}
                    variant="symbol"
                    onClick={skipPostSignup}
                />
            </LinkOrSpan>

            <div className="signin-signup-page post-signup-page">
                <PageTitle title="Welcome" />

                <HeroPage
                    lessPadding={true}
                    className="text-left"
                    body={
                        <div className="post-signup-page__container">
                            <h2>Get started with Sourcegraph</h2>
                            <p className="text-muted pb-3">
                                Three quick steps to add your repositories and get searching with Sourcegraph
                            </p>
                            <div className="mt-4 pb-3">
                                <Steps current={currentStepNumber}>
                                    <StepList numeric={true}>
                                        <Step borderColor="purple">Connect with code hosts</Step>
                                        <Step borderColor="blue">Add repositories</Step>
                                        <Step borderColor="orange">Start searching</Step>
                                    </StepList>
                                    <StepPanels>
                                        <StepPanel>
                                            {externalServices && (
                                                <CodeHostsConnection
                                                    loading={loadingServices}
                                                    user={user}
                                                    error={errorServices}
                                                    externalServices={externalServices}
                                                    context={context}
                                                    refetch={refetchExternalServices}
                                                />
                                            )}
                                        </StepPanel>
                                        <StepPanel>
                                            <>
                                                <h3>Add repositories</h3>
                                                <p className="text-muted">
                                                    Choose repositories you own or collaborate on from your code hosts
                                                    to search with Sourcegraph. We’ll sync and index these repositories
                                                    so you can search your code all in one place.
                                                </p>
                                            </>
                                        </StepPanel>
                                        <StepPanel>
                                            <StartSearching
                                                isDoneCloning={isDoneCloning}
                                                cloningStatusLines={cloningStatusLines}
                                                cloningStatusLoading={cloningStatusLoading}
                                            />
                                        </StepPanel>
                                    </StepPanels>
                                </Steps>
                            </div>
                            {/* This should be part of step panel */}
                            {/* <div className="mt-4 pb-3">{currentStep.content}</div> */}
                            <div className="mt-4">
                                <button
                                    type="button"
                                    className="btn btn-primary float-right ml-2"
                                    disabled={!isCurrentStepComplete()}
                                    onClick={isLastStep ? goToSearch : goToNextTab}
                                >
                                    {isLastStep ? 'Start searching' : 'Continue'}
                                </button>

                                {!isLastStep && (
                                    <button
                                        type="button"
                                        className="btn btn-link font-weight-normal text-secondary float-right"
                                        onClick={skipPostSignup}
                                    >
                                        Not right now
                                    </button>
                                )}
                            </div>
                            {/* debugging */}
                            <div className="pt-5">
                                <hr />
                                <br />
                                <p>🚧&nbsp; Debugging navigation&nbsp;🚧</p>
                                <button
                                    type="button"
                                    className="btn btn-secondary"
                                    disabled={currentStepNumber === 1}
                                    onClick={() => setCurrentStepNumber(currentStepNumber - 1)}
                                >
                                    previous tab
                                </button>
                                &nbsp;&nbsp;
                                <button
                                    type="button"
                                    className="btn btn-secondary"
                                    disabled={currentStepNumber === steps.length}
                                    onClick={goToNextTab}
                                >
                                    next tab
                                </button>
                            </div>
                        </div>
                    }
                />
            </div>
        </>
    )
}
