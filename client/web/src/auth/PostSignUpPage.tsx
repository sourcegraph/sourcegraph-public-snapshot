import React, { FunctionComponent, useState, useCallback, useEffect } from 'react'
import { useLocation, useHistory } from 'react-router'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { Steps, Step } from '@sourcegraph/wildcard/src/components/Steps'
import { Terminal } from '@sourcegraph/wildcard/src/components/Terminal'

import { AuthenticatedUser } from '../auth'
import { queryExternalServices } from '../components/externalServices/backend'
import { HeroPage } from '../components/HeroPage'
import { PageTitle } from '../components/PageTitle'
import { UserAreaUserFields, ListExternalServiceFields } from '../graphql-operations'
import { SourcegraphContext } from '../jscontext'
import { UserCodeHosts } from '../user/settings/codeHosts/UserCodeHosts'

import { getReturnTo } from './SignInSignUpCommon'

// import { Redirect } from 'react-router-dom'

interface Props {
    authenticatedUser: AuthenticatedUser
    context: Pick<SourcegraphContext, 'authProviders'>
    user: UserAreaUserFields
    routingPrefix: string
}

export const PostSignUpPage: FunctionComponent<Props> = ({ authenticatedUser: user, context }) => {
    // post sign-up flow is available only for .com and only in two cases, user:
    // 1. is authenticated and has AllowUserViewPostSignup tag
    // 2. is authenticated and enablePostSignupFlow experimental feature is ON
    // sourcegraphDotComMode &&
    // ((authenticatedUser && experimentalFeatures.enablePostSignupFlow) ||
    //     authenticatedUser?.tags.includes('AllowUserViewPostSignup')) ? (

    const [currentStepNumber, setCurrentStepNumber] = useState(1)
    const location = useLocation()
    const history = useHistory()

    const [externalServices, setExternalServices] = useState<ListExternalServiceFields[]>()

    const fetchExternalServices = useCallback(async (): Promise<void> => {
        const { nodes: fetchedServices } = await queryExternalServices({
            namespace: user.id,
            first: null,
            after: null,
        }).toPromise()

        setExternalServices(fetchedServices)
    }, [user.id])

    useEffect(() => {
        fetchExternalServices().catch(error => console.log(error))
    }, [fetchExternalServices])

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

                {externalServices ? (
                    <UserCodeHosts
                        user={user}
                        externalServices={externalServices}
                        context={context}
                        onDidError={error => console.warn('<UserCodeHosts .../>', error)}
                        onDidRemove={() => fetchExternalServices()}
                    />
                ) : (
                    <div className="d-flex justify-content-center">
                        <LoadingSpinner className="icon-inline" />
                    </div>
                )}
            </>
        ),
        // step is considered complete when user has at least one external service
        isComplete: (): boolean => Array.isArray(externalServices) && externalServices.length > 0,
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
        isComplete: () => false,
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

    const onStepTabClick = (clickedStepTabIndex: number): void => {
        /**
         * User can navigate through the steps by clicking the step's tab when:
         * 1. navigating back and the navigated to step is complete
         * 2. navigating forward and the current step is complete
         */

        const isValidNavigationBack =
            clickedStepTabIndex < currentStepNumber && steps[clickedStepTabIndex - 1].isComplete()
        const isValidNavigationForward = clickedStepTabIndex > currentStepNumber && isCurrentStepComplete()

        if (isValidNavigationBack || isValidNavigationForward) {
            setCurrentStepNumber(clickedStepTabIndex)
        }
    }

    return (
        <div className="signin-signup-page post-signup-page">
            <PageTitle title="Post sign up page" />

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
                                    onClick={() => history.push(getReturnTo(location))}
                                >
                                    Not right now
                                </button>
                            )}
                        </div>

                        {/* debugging */}
                        <div className="pt-5">
                            <hr />
                            <br />
                            <p>🚧 Debugging buttons 🚧</p>
                            <button
                                type="button"
                                className="btn btn-secondary"
                                disabled={currentStepNumber === 1}
                                onClick={() => setCurrentStepNumber(currentStepNumber - 1)}
                            >
                                previous tab
                            </button>
                            &nbsp;
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
    )
}
// ) : (
//     <Redirect to={getReturnTo(location)} />
// )

// Is this part of the sign-up flow? I think this would ba a getting-started stage isolated from the sign-up
//
