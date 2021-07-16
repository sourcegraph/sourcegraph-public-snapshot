import React, { FunctionComponent, useState } from 'react'
import { useLocation, useHistory } from 'react-router'

import { LinkOrSpan } from '@sourcegraph/shared/src/components/LinkOrSpan'
import { TelemetryService } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { BrandLogo } from '@sourcegraph/web/src/components/branding/BrandLogo'
import { HeroPage } from '@sourcegraph/web/src/components/HeroPage'
import { LoaderButton } from '@sourcegraph/web/src/components/LoaderButton'
import { Steps, Step, StepList, StepPanels, StepPanel, StepActions } from '@sourcegraph/wildcard/src/components/Steps'

import { PageTitle } from '../components/PageTitle'
import { UserAreaUserFields } from '../graphql-operations'
import { SourcegraphContext } from '../jscontext'
import { SelectAffiliatedRepos, AffiliatedReposReference } from '../user/settings/repositories/SelectAffiliatedRepos'

import { getReturnTo } from './SignInSignUpCommon'
import { useAffiliatedRepos } from './useAffiliatedRepos'
import { useExternalServices } from './useExternalServices'
import { useRepoCloningStatus } from './useRepoCloningStatus'
import { useSelectedRepos } from './useSelectedRepos'
import { CodeHostsConnection } from './welcome/CodeHostsConnection'
import { Footer } from './welcome/Footer'
import { StartSearching } from './welcome/StartSearching'

interface PostSignUpPage {
    authenticatedUser: UserAreaUserFields
    context: Pick<SourcegraphContext, 'authProviders' | 'experimentalFeatures' | 'sourcegraphDotComMode'>
    telemetryService: TelemetryService
}

interface Step {
    content: React.ReactElement
    isComplete: () => boolean
    prefetch?: () => void
    onNextButtonClick?: () => Promise<void>
}

// const delay = (milliseconds: number): Promise<void> => new Promise(resolve => setTimeout(resolve, milliseconds))

export const PostSignUpPage: FunctionComponent<PostSignUpPage> = ({
    authenticatedUser: user,
    context,
    // telemetryService,
}) => {
    // const [currentStepNumber, setCurrentStepNumber] = useState(1)
    const location = useLocation()
    // const history = useHistory()

    const {
        trigger: fetchCloningStatus,
        repos: cloningStatusLines,
        loading: cloningStatusLoading,
        isDoneCloning,
    } = useRepoCloningStatus({ userId: user.id, pollInterval: 2000 })

    const { externalServices, loadingServices, errorServices, refetchExternalServices } = useExternalServices(user.id)
    // const { fetchAffiliatedRepos, affiliatedRepos } = useAffiliatedRepos(user.id)
    // const { fetchSelectedRepos, selectedRepos } = useSelectedRepos(user.id)

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

    // const firstStep = {
    //     content: (
    //         <>
    //             {currentStepNumber === 1 && externalServices && (
    //                 <CodeHostsConnection
    //                     loading={loadingServices}
    //                     user={user}
    //                     error={errorServices}
    //                     externalServices={externalServices}
    //                     context={context}
    //                     refetch={refetchExternalServices}
    //                 />
    //             )}
    //         </>
    //     ),
    //     // step is considered complete when user has at least one external service connected.
    //     isComplete: (): boolean => !!externalServices && externalServices?.length > 0,
    // }

    // const secondStep = {
    //     content: (
    //         <>
    //             {currentStepNumber === 2 && (
    //                 <>
    //                     <h3>Add repositories</h3>
    //                     <p className="text-muted">
    //                         Choose repositories you own or collaborate on from your code hosts to search with
    //                         Sourcegraph. We’ll sync and index these repositories so you can search your code all in one
    //                         place.
    //                     </p>
    //                     <SelectAffiliatedRepos
    //                         ref={AffiliatedReposReference}
    //                         onSelection={setDidSelectAffiliatedRepos}
    //                         repos={affiliatedRepos}
    //                         externalServices={externalServices}
    //                         selectedRepos={selectedRepos}
    //                         authenticatedUser={user}
    //                         telemetryService={telemetryService}
    //                     />
    //                 </>
    //             )}
    //         </>
    //     ),
    //     isComplete: () => true /* didSelectAffiliatedRepos */,
    //     onNextButtonClick: async () => {
    //         await AffiliatedReposReference.current?.submit()
    //     },
    //     prefetch: () => {
    //         fetchSelectedRepos()
    //         fetchAffiliatedRepos()
    //     },
    // }

    // const thirdStep = {
    //     content: (
    //         <>
    //             {currentStepNumber === 3 && (
    //                 <StartSearching
    //                     isDoneCloning={isDoneCloning}
    //                     cloningStatusLines={cloningStatusLines}
    //                     cloningStatusLoading={cloningStatusLoading}
    //                 />
    //             )}
    //         </>
    //     ),
    //     isComplete: () => isDoneCloning,
    //     prefetch: () => fetchCloningStatus(),
    // }

    // const steps: Step[] = [firstStep, secondStep, thirdStep]

    // // Steps helpers
    // const isLastStep = currentStepNumber === steps.length
    // const currentStep = steps[currentStepNumber - 1]

    // const goToNextTab = async (): Promise<void> => {
    //     if (currentStep.onNextButtonClick) {
    //         setIsNextStepLoading(true)
    //         await currentStep.onNextButtonClick()
    //         // TODO: remove this
    //         await delay(3000)
    //         setIsNextStepLoading(false)
    //     }

    //     // currentStepNumber is not zero based, it'll get the next step
    //     const nextStep = steps[currentStepNumber]
    //     if (nextStep.prefetch) {
    //         nextStep.prefetch()
    //     }

    //     setCurrentStepNumber(currentStepNumber + 1)
    // }
    // const goToSearch = (): void => history.push(getReturnTo(location))
    // const isCurrentStepComplete = (): boolean => currentStep?.isComplete()
    // const skipPostSignup = (): void => history.push(getReturnTo(location))

    // const onStepTabClick = (clickedStepTabNumber: number): void => {
    //     /**
    //      * User can navigate through the steps by clicking the step's tab when:
    //      * 1. navigating back
    //      * 2. navigating one step forward when the current step is complete
    //      * 3. navigating many steps forward when all of the steps, from the
    //      * current one to the clickedStepTabNumber step but not including are
    //      * complete.
    //      */

    //     // do nothing for the current tab
    //     if (clickedStepTabNumber === currentStepNumber) {
    //         return
    //     }

    //     if (clickedStepTabNumber < currentStepNumber) {
    //         // allow to navigate back since all of the previous steps had to be completed
    //         setCurrentStepNumber(clickedStepTabNumber)
    //     } else if (currentStepNumber - 1 === clickedStepTabNumber) {
    //         // forward navigation

    //         // if navigating to the next tab, check if the current step is completed
    //         if (isCurrentStepComplete()) {
    //             setCurrentStepNumber(clickedStepTabNumber)
    //         }
    //     } else {
    //         // if navigating further away check [current, ..., clicked)
    //         const areInBetweenStepsComplete = steps
    //             .slice(currentStepNumber - 1, clickedStepTabNumber - 1)
    //             .every(step => step.isComplete())

    //         if (areInBetweenStepsComplete) {
    //             setCurrentStepNumber(clickedStepTabNumber)
    //         }
    //     }
    // }

    return (
        <>
            <LinkOrSpan to={getReturnTo(location)} className="post-signup-page__logo-link">
                <BrandLogo
                    className="position-absolute ml-3 mt-3 post-signup-page__logo"
                    isLightTheme={true}
                    variant="symbol"
                    // onClick={skipPostSignup}
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
                                <Steps initialStep={1}>
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
                                    <StepList numeric={true}>
                                        <Step borderColor="purple">Connect with code hosts</Step>
                                        <Step borderColor="blue">Add repositories</Step>
                                        <Step borderColor="orange">Start searching</Step>
                                    </StepList>
                                    <StepActions>
                                        <Footer />
                                    </StepActions>
                                </Steps>
                            </div>
                            {/* This should be part of step panel */}
                            {/* <div className="mt-4 pb-3">{currentStep.content}</div>
                            <div className="mt-4">
                                <LoaderButton
                                    type="button"
                                    alwaysShowLabel={true}
                                    label={isLastStep ? 'Start searching' : 'Continue'}
                                    className="btn btn-primary float-right ml-2"
                                    disabled={!!externalServices && externalServices?.length === 0}
                                    // disabled={!isCurrentStepComplete()}
                                    onClick={isLastStep ? goToSearch : goToNextTab}
                                />

                                {!isLastStep && (
                                    <button
                                        type="button"
                                        className="btn btn-link font-weight-normal text-secondary float-right"
                                        onClick={skipPostSignup}
                                    >
                                        Not right now
                                    </button>
                                )}
                            </div> */}
                        </div>
                    }
                />
            </div>
        </>
    )
}

PostSignUpPage.whyDidYouRender = true
