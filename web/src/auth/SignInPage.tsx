/* eslint-disable react/no-array-index-key */
import * as H from 'history'
import KeyIcon from 'mdi-react/KeyIcon'
import React, { useEffect } from 'react'
import { Redirect } from 'react-router-dom'
import { HeroPage } from '../components/HeroPage'
import { PageTitle } from '../components/PageTitle'
import { eventLogger } from '../tracking/eventLogger'
import { getReturnTo } from './SignInSignUpCommon'
import { UsernamePasswordSignInForm } from './UsernamePasswordSignInForm'
import { AuthenticatedUser } from '../auth'

interface SignInPageProps {
    location: H.Location
    history: H.History
    authenticatedUser: AuthenticatedUser | null
}

export const SignInPage: React.FunctionComponent<SignInPageProps> = props => {
    useEffect(() => eventLogger.logViewEvent('SignIn', false))

    if (props.authenticatedUser) {
        const returnTo = getReturnTo(props.location)
        return <Redirect to={returnTo} />
    }

    return (
        <div className="signin-signup-page sign-in-page">
            <PageTitle title="Sign in" />
            <HeroPage
                icon={KeyIcon}
                title="Sign into Sourcegraph"
                body={
                    window.context.authProviders && window.context.authProviders.length > 0 ? (
                        <div className="mb-4">
                            {window.context.authProviders.map((provider, index) =>
                                provider.isBuiltin ? (
                                    <UsernamePasswordSignInForm key={index} {...props} />
                                ) : (
                                    <div className="mb-2" key={index}>
                                        <a href={provider.authenticationURL} className="btn btn-secondary">
                                            Continue with {provider.displayName}
                                        </a>
                                    </div>
                                )
                            )}
                        </div>
                    ) : (
                        <div className="alert alert-info mt-3">
                            No authentication providers are available. Contact a site administrator for help.
                        </div>
                    )
                }
            />
        </div>
    )
}
