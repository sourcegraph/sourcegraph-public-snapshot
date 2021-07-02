import React, { FunctionComponent, useState } from 'react'
import { useLocation, useHistory } from 'react-router'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { LinkOrSpan } from '@sourcegraph/shared/src/components/LinkOrSpan'
import { useQuery } from '@sourcegraph/shared/src/graphql/graphql'
import { BrandLogo } from '@sourcegraph/web/src/components/branding/BrandLogo'
import { Steps, Step } from '@sourcegraph/wildcard/src/components/Steps'
import { Terminal } from '@sourcegraph/wildcard/src/components/Terminal'

import { AuthenticatedUser } from '../auth'
import { EXTERNAL_SERVICES } from '../components/externalServices/backend'
import { HeroPage } from '../components/HeroPage'
import { PageTitle } from '../components/PageTitle'
import { UserAreaUserFields, ExternalServicesVariables, ExternalServicesResult } from '../graphql-operations'
import { SourcegraphContext } from '../jscontext'
import { UserCodeHosts } from '../user/settings/codeHosts/UserCodeHosts'

import { getReturnTo } from './SignInSignUpCommon'

interface Props {
    authenticatedUser: AuthenticatedUser
    context: Pick<SourcegraphContext, 'authProviders' | 'experimentalFeatures' | 'sourcegraphDotComMode'>
    user: UserAreaUserFields
    routingPrefix: string
}

export const PostSignUpPage: FunctionComponent<Props> = ({ authenticatedUser: user, context }) => {
    const [currentStepNumber, setCurrentStepNumber] = useState(1)
    const location = useLocation()
    const history = useHistory()

    /**
     * post sign-up flow is available only for .com and only in two cases, user:
     * 1. is authenticated and has AllowUserViewPostSignup tag
     * 2. is authenticated and enablePostSignupFlow experimental feature is ON
     */

    if (
        !user ||
        !context.sourcegraphDotComMode ||
        !context.experimentalFeatures?.enablePostSignupFlow ||
        !user?.tags.includes('AllowUserViewPostSignup')
    ) {
        // TODO: do this on the backend
        history.push(getReturnTo(location))
    }

    const { data, loading, error, refetch } = useQuery<ExternalServicesResult, ExternalServicesVariables>(
        EXTERNAL_SERVICES,
        {
            variables: {
                namespace: user.id,
                first: null,
                after: null,
            },
        }
    )

    if (loading) {
        return (
            <div className="d-flex justify-content-center">
                <LoadingSpinner className="icon-inline" />
            </div>
        )
    }

    if (error) {
        console.log(error)
    }

    const connectCodeHosts = {
        content: (
            <>
                <div className="mb-4">
                    <h3>Connect with code hosts</h3>
                    <p className="text-muted">
                        Connect with providers where your source code is hosted. Then, choose the repositories you’d
                        like to search with Sourcegraph.
                    </p>
                </div>
                {data?.externalServices?.nodes && (
                    <UserCodeHosts
                        user={user}
                        externalServices={data.externalServices.nodes}
                        context={context}
                        onDidError={error => console.warn('<UserCodeHosts .../>', error)}
                        onDidRemove={() => refetch()}
                    />
                )}
            </>
        ),
        // step is considered complete when user has at least one external service
        isComplete: (): boolean =>
            !!data && Array.isArray(data?.externalServices?.nodes) && data.externalServices.nodes.length > 0,
    }

    const addRepositories = {
        content: (
            <>
                <h3>Add repositories</h3>
                <p className="text-muted">
                    Choose repositories you own or collaborate on from your code hosts to search with Sourcegraph. We’ll
                    sync and index these repositories so you can search your code all in one place.
                </p>
            </>
        ),
        isComplete: () => true,
    }

    const startSearching = {
        content: (
            <>
                <h3>Start searching...</h3>
                <p className="text-muted">
                    We’re cloning your repos to Sourcegraph. In just a few moments, you can make your first search!
                </p>
                <Terminal />
            </>
        ),
        isComplete: () => false,
    }

    const steps = [connectCodeHosts, addRepositories, startSearching]

    // Steps helpers
    const isLastStep = currentStepNumber === steps.length
    const currentStep = steps[currentStepNumber - 1]

    const goToNextTab = (): void => setCurrentStepNumber(currentStepNumber + 1)
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
                                <Steps current={currentStepNumber} numbered={true} onTabClick={onStepTabClick}>
                                    <Step title="Connect with code hosts" borderColor="purple" />
                                    <Step title="Add repositories" borderColor="blue" />
                                    <Step title="Start searching" borderColor="orange" />
                                </Steps>
                            </div>
                            <div className="mt-4 pb-3">{currentStep.content}</div>

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
