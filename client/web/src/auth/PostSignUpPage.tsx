import React, { FunctionComponent, useState } from 'react'
import { Redirect } from 'react-router-dom'

import { Steps, Step } from '@sourcegraph/wildcard/src/components/Steps'

import { AuthenticatedUser } from '../auth'
import { codeHostExternalServices } from '../components/externalServices/externalServices'
import { HeroPage } from '../components/HeroPage'
import { PageTitle } from '../components/PageTitle'
import { UserAreaUserFields } from '../graphql-operations'
import { SourcegraphContext } from '../jscontext'
import { UserAddCodeHostsPage } from '../user/settings/codeHosts/UserAddCodeHostsPage'

// import { getReturnTo } from './SignInSignUpCommon'

interface Props {
    authenticatedUser: AuthenticatedUser
    context: Pick<SourcegraphContext, 'authProviders'>
    user: UserAreaUserFields
    routingPrefix: string
}

export const PostSignUpPage: FunctionComponent<Props> = ({ authenticatedUser, routingPrefix, context }) => {
    // post sign-up flow is available only for .com and only in two cases, user:
    // 1. is authenticated and has AllowUserViewPostSignup tag
    // 2. is authenticated and enablePostSignupFlow experimental feature is ON
    // sourcegraphDotComMode &&
    // ((authenticatedUser && experimentalFeatures.enablePostSignupFlow) ||
    //     authenticatedUser?.tags.includes('AllowUserViewPostSignup')) ? (

    const [currentStep, setCurrentStep] = useState(1)

    const connectCodeHosts = (
        <>
            <div className="mb-4">
                <h3>Connect with code hosts</h3>
                <p className="text-muted">
                    Connect with providers where your source code is hosted. Then, choose the repositories you’d like to
                    search with Sourcegraph.
                </p>
            </div>

            <UserAddCodeHostsPage
                user={{ id: authenticatedUser.id, tags: authenticatedUser.tags }}
                context={context}
                routingPrefix={routingPrefix}
                codeHostExternalServices={{
                    github: codeHostExternalServices.github,
                    gitlabcom: codeHostExternalServices.gitlabcom,
                }}
                showHeader={false}
                onUserExternalServicesOrRepositoriesUpdate={() => {}}
            />
        </>
    )
    const addRepositories = (
        <>
            <h3>Add repositories</h3>
            <p className="text-muted">
                Choose repositories you own or collaborate on from your code hosts to search with Sourcegraph. We’ll
                sync and index these repositories so you can search your code all in one place.
            </p>
        </>
    )
    const startSearching = (
        <>
            <h3>Start searching...</h3>
            <p className="text-muted">
                We’re cloning your repos to Sourcegraph. In just a few moments, you can make your first search!
            </p>
        </>
    )

    const steps = [connectCodeHosts, addRepositories, startSearching]

    return (
        <div className="signin-signup-page post-signup-page">
            <PageTitle title="Post sign up page" />

            <HeroPage
                lessPadding={true}
                className="text-left"
                body={
                    <div className="post-signup-page__container">
                        <h2>Get started with Sourcegraph</h2>
                        <p className="text-muted">
                            Three quick steps to add your repositories and get searching with Sourcegraph
                        </p>
                        <div className="mt-3 pb-3">
                            <Steps current={currentStep} numbered={true}>
                                <Step title="Connect with code hosts" borderColor="purple" />
                                <Step title="Add repositories" borderColor="blue" />
                                <Step title="Start searching" borderColor="orange" />
                            </Steps>
                        </div>
                        <div className="mt-4">{steps[currentStep - 1]}</div>
                        <br />
                        <button
                            type="button"
                            disabled={currentStep === 1}
                            onClick={() => setCurrentStep(currentStep - 1)}
                        >
                            previous
                        </button>
                        &nbsp;
                        <button
                            type="button"
                            disabled={currentStep === steps.length}
                            onClick={() => setCurrentStep(currentStep + 1)}
                        >
                            next
                        </button>
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
