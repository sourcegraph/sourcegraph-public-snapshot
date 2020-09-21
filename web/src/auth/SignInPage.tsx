import * as H from 'history'
import React, { useEffect } from 'react'
import { Link, Redirect } from 'react-router-dom'
import { HeroPage } from '../components/HeroPage'
import { PageTitle } from '../components/PageTitle'
import { eventLogger } from '../tracking/eventLogger'
import { getReturnTo } from './SignInSignUpCommon'
import { UsernamePasswordSignInForm } from './UsernamePasswordSignInForm'
import { AuthenticatedUser } from '../auth'
import { SourcegraphIcon } from './icons'
import { partition } from 'lodash'
import { OrDivider } from './OrDivider'

interface SignInPageProps {
    location: H.Location
    history: H.History
    authenticatedUser: AuthenticatedUser | null
}

export const SignInPage: React.FunctionComponent<SignInPageProps> = props => {
    useEffect(() => eventLogger.logViewEvent('SignIn', null, false))

    if (props.authenticatedUser) {
        const returnTo = getReturnTo(props.location)
        return <Redirect to={returnTo} />
    }

    const [[builtInAuthProvider], thirdPartyAuthProviders] = partition(
        window.context.authProviders,
        provider => provider.isBuiltin
    )

    // builtInAuthProvider = null
    // thirdPartyAuthProviders = []

    const body =
        !builtInAuthProvider && thirdPartyAuthProviders.length === 0 ? (
            <div className="alert alert-info mt-3">
                No authentication providers are available. Contact a site administrator for help.
            </div>
        ) : (
            <div className="mb-4">
                <div className="signin-signup-form signin-form test-signin-form border rounded p-4 mt-4 my-3">
                    {builtInAuthProvider && <UsernamePasswordSignInForm {...props} />}
                    {builtInAuthProvider && thirdPartyAuthProviders.length > 0 && <OrDivider className="mb-3" />}
                    {thirdPartyAuthProviders.map((provider, index) => (
                        // Use index as key because display name may not be unique. This is OK
                        // here because this list will not be updated during this component's lifetime.
                        /* eslint-disable react/no-array-index-key */
                        <div className="mb-2" key={index}>
                            <a href={provider.authenticationURL} className="btn btn-secondary btn-block">
                                Continue with {provider.displayName}
                            </a>
                        </div>
                    ))}
                </div>
                {window.context.allowSignup ? (
                    <p>
                        <Link to={`/sign-up${location.search}`}>New to Sourcegraph? Sign up.</Link>
                    </p>
                ) : (
                    <p className="text-muted">To create an account, contact the site admin.</p>
                )}
            </div>
        )

    return (
        <div className="signin-signup-page sign-in-page web-content">
            <PageTitle title="Sign in" />
            <HeroPage
                icon={SourcegraphIcon}
                iconLinkTo={window.context.sourcegraphDotComMode ? '/search' : undefined}
                iconClassName="bg-transparent"
                title={
                    window.context.sourcegraphDotComMode
                        ? 'Sign in to Sourcegraph Cloud'
                        : 'Sign in to Sourcegraph Server'
                }
                body={body}
            />
        </div>
    )
}
