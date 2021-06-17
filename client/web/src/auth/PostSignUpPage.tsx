import * as H from 'history'
import React, { FunctionComponent } from 'react'
import { Redirect } from 'react-router-dom'

import { AuthenticatedUser } from '../auth'
import { HeroPage } from '../components/HeroPage'
import { PageTitle } from '../components/PageTitle'
import { SourcegraphContext } from '../jscontext'

import { getReturnTo } from './SignInSignUpCommon'

interface Props {
    authenticatedUser?: AuthenticatedUser | null
    context: Pick<SourcegraphContext, 'allowSignup' | 'sourcegraphDotComMode' | 'experimentalFeatures'>
    location: H.Location
}

export const PostSignUpPage: FunctionComponent<Props> = ({
    authenticatedUser,
    location,
    context: { sourcegraphDotComMode, experimentalFeatures },
}) =>
    // post sign-up flow is available only for .com and only in two cases, user:
    // 1. is authenticated and has AllowUserViewPostSignup tag
    // 2. is authenticated and enablePostSignupFlow experimental feature is ON
    sourcegraphDotComMode &&
    ((authenticatedUser && experimentalFeatures.enablePostSignupFlow) ||
        authenticatedUser?.tags.includes('AllowUserViewPostSignup')) ? (
        <div className="signin-signup-page post-signup-page">
            <PageTitle title="Post sign up page" />

            <HeroPage
                lessPadding={true}
                className="text-left"
                body={
                    <div className="post-signup-page__container">
                        <h2>Get started with Sourcegraph</h2>
                        <p>Three quick steps to add your repositories and get searching with Sourcegraph</p>
                    </div>
                }
            />
        </div>
    ) : (
        <Redirect to={getReturnTo(location)} />
    )
