import * as H from 'history'
import React, { FunctionComponent, useState } from 'react'
// import { Redirect } from 'react-router-dom'

import { Steps, Step } from '@sourcegraph/wildcard/src/components/Steps'

import { AuthenticatedUser } from '../auth'
import { HeroPage } from '../components/HeroPage'
import { PageTitle } from '../components/PageTitle'
import { SourcegraphContext } from '../jscontext'

// import { getReturnTo } from './SignInSignUpCommon'

interface Props {
    authenticatedUser?: AuthenticatedUser | null
    context: Pick<SourcegraphContext, 'allowSignup' | 'sourcegraphDotComMode' | 'experimentalFeatures'>
    location: H.Location
}

export const PostSignUpPage: FunctionComponent<Props> = ({
    authenticatedUser,
    location,
    context: { sourcegraphDotComMode, experimentalFeatures },
}) => {
    // post sign-up flow is available only for .com and only in two cases, user:
    // 1. is authenticated and has AllowUserViewPostSignup tag
    // 2. is authenticated and enablePostSignupFlow experimental feature is ON
    // sourcegraphDotComMode &&
    // ((authenticatedUser && experimentalFeatures.enablePostSignupFlow) ||
    //     authenticatedUser?.tags.includes('AllowUserViewPostSignup')) ? (

    const [current, setCurrent] = useState(1)

    const content1 = <p>Content #1</p>
    const content2 = <p>Content #2</p>
    const content3 = <p>Content #3</p>

    const content = [content1, content2, content3]

    return (
        <div className="signin-signup-page post-signup-page">
            <PageTitle title="Post sign up page" />

            <HeroPage
                lessPadding={true}
                className="text-left"
                body={
                    <div className="post-signup-page__container">
                        <h2>Get started with Sourcegraph</h2>
                        <p>Three quick steps to add your repositories and get searching with Sourcegraph</p>
                        <Steps current={current} numbered={true}>
                            <Step title="Connect with code hosts" />
                            <Step title="Add repositories" />
                            <Step title="Start searching" />
                        </Steps>
                        <div>{content[current - 1]}</div>
                        <button type="button" onClick={() => setCurrent(current + 1)}>
                            next
                        </button>
                        <button type="button" onClick={() => setCurrent(current - 1)}>
                            prev
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
