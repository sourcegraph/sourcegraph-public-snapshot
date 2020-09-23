import * as H from 'history'
import React, { useEffect, useState } from 'react'
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
import { ErrorAlert } from '../components/alerts'
import { stripURLParameters } from '../tracking/analyticsUtils'
import classNames from 'classnames'
import GithubIcon from 'mdi-react/GithubIcon'

interface SignInPageProps {
    location: H.Location
    history: H.History
    authenticatedUser: AuthenticatedUser | null
}

export const SignInPage: React.FunctionComponent<SignInPageProps> = props => {
    useEffect(() => eventLogger.logViewEvent('SignIn', null, false))

    const [authError, setAuthError] = useState<Error | null>(() => {
        // Display 3rd party auth errors (redirect with param 'auth_error')
        const authErrorMessage = new URLSearchParams(location.search).get('auth_error')
        stripURLParameters(window.location.href, ['auth_error'])
        return authErrorMessage ? new Error(authErrorMessage) : null
    })

    if (props.authenticatedUser) {
        const returnTo = getReturnTo(props.location)
        return <Redirect to={returnTo} />
    }

    const [[builtInAuthProvider], thirdPartyAuthProviders] = partition(
        window.context.authProviders,
        provider => provider.isBuiltin
    )

    const body =
        !builtInAuthProvider && thirdPartyAuthProviders.length === 0 ? (
            <div className="alert alert-info mt-3">
                No authentication providers are available. Contact a site administrator for help.
            </div>
        ) : (
            <div className="mb-4 signin-page__container">
                {authError && (
                    <ErrorAlert className="mt-4 mb-0" error={authError} icon={false} history={props.history} />
                )}
                <div
                    className={classNames(
                        'signin-signup-form signin-form test-signin-form border rounded p-4 my-3',
                        authError ? 'mt-3' : 'mt-4'
                    )}
                >
                    {builtInAuthProvider && <UsernamePasswordSignInForm {...props} onAuthError={setAuthError} />}
                    {builtInAuthProvider && thirdPartyAuthProviders.length > 0 && <OrDivider className="mb-3 py-1" />}
                    {thirdPartyAuthProviders.map((provider, index) => (
                        // Use index as key because display name may not be unique. This is OK
                        // here because this list will not be updated during this component's lifetime.
                        /* eslint-disable react/no-array-index-key */
                        <div className="mb-2" key={index}>
                            <a href={provider.authenticationURL} className="btn btn-secondary btn-block">
                                {provider.displayName === 'GitHub' && (
                                    <>
                                        <GithubIcon className="icon-inline" />{' '}
                                    </>
                                )}
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
