import React, { useEffect, useState } from 'react'

import classNames from 'classnames'
import * as H from 'history'
import { partition } from 'lodash'
import GithubIcon from 'mdi-react/GithubIcon'
import GitlabIcon from 'mdi-react/GitlabIcon'
import { Redirect } from 'react-router-dom'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { Button, Link, Alert, Icon } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../auth'
import { HeroPage } from '../components/HeroPage'
import { PageTitle } from '../components/PageTitle'
import { SourcegraphContext } from '../jscontext'
import { eventLogger } from '../tracking/eventLogger'

import { SourcegraphIcon } from './icons'
import { OrDivider } from './OrDivider'
import { getReturnTo, maybeAddPostSignUpRedirect } from './SignInSignUpCommon'
import { UsernamePasswordSignInForm } from './UsernamePasswordSignInForm'

import signInSignUpCommonStyles from './SignInSignUpCommon.module.scss'

interface SignInPageProps {
    location: H.Location
    history: H.History
    authenticatedUser: AuthenticatedUser | null
    context: Pick<
        SourcegraphContext,
        'allowSignup' | 'authProviders' | 'sourcegraphDotComMode' | 'xhrHeaders' | 'resetPasswordEnabled'
    >
}

export const SignInPage: React.FunctionComponent<React.PropsWithChildren<SignInPageProps>> = props => {
    useEffect(() => eventLogger.logViewEvent('SignIn', null, false))

    const [error, setError] = useState<Error | null>(null)

    if (props.authenticatedUser) {
        const returnTo = getReturnTo(props.location)
        return <Redirect to={returnTo} />
    }

    const [[builtInAuthProvider], thirdPartyAuthProviders] = partition(
        props.context.authProviders,
        provider => provider.isBuiltin
    )

    const body =
        !builtInAuthProvider && thirdPartyAuthProviders.length === 0 ? (
            <Alert className="mt-3" variant="info">
                No authentication providers are available. Contact a site administrator for help.
            </Alert>
        ) : (
            <div className={classNames('mb-4 pb-5', signInSignUpCommonStyles.signinPageContainer)}>
                {error && <ErrorAlert className="mt-4 mb-0 text-left" error={error} icon={false} />}
                <div
                    className={classNames(
                        'test-signin-form rounded p-4 my-3',
                        signInSignUpCommonStyles.signinSignupForm,
                        error ? 'mt-3' : 'mt-4'
                    )}
                >
                    {builtInAuthProvider && (
                        <UsernamePasswordSignInForm
                            {...props}
                            onAuthError={setError}
                            noThirdPartyProviders={thirdPartyAuthProviders.length === 0}
                        />
                    )}
                    {builtInAuthProvider && thirdPartyAuthProviders.length > 0 && <OrDivider className="mb-3 py-1" />}
                    {thirdPartyAuthProviders.map((provider, index) => (
                        // Use index as key because display name may not be unique. This is OK
                        // here because this list will not be updated during this component's lifetime.
                        /* eslint-disable react/no-array-index-key */
                        <div className="mb-2" key={index}>
                            <Button
                                href={maybeAddPostSignUpRedirect(provider.authenticationURL)}
                                className="btn-block"
                                variant="secondary"
                                as="a"
                            >
                                {provider.serviceType === 'github' && (
                                    <>
                                        <Icon as={GithubIcon} />{' '}
                                    </>
                                )}
                                {provider.serviceType === 'gitlab' && (
                                    <>
                                        <Icon as={GitlabIcon} />{' '}
                                    </>
                                )}
                                Continue with {provider.displayName}
                            </Button>
                        </div>
                    ))}
                </div>
                {props.context.allowSignup ? (
                    <p>
                        New to Sourcegraph? <Link to={`/sign-up${location.search}`}>Sign up</Link>
                    </p>
                ) : (
                    <p className="text-muted">Need an account? Contact your site admin</p>
                )}
            </div>
        )

    return (
        <div className={signInSignUpCommonStyles.signinSignupPage}>
            <PageTitle title="Sign in" />
            <HeroPage
                icon={SourcegraphIcon}
                iconLinkTo={props.context.sourcegraphDotComMode ? '/search' : undefined}
                iconClassName="bg-transparent"
                lessPadding={true}
                title={
                    props.context.sourcegraphDotComMode
                        ? 'Sign in to Sourcegraph Cloud'
                        : 'Sign in to Sourcegraph Server'
                }
                body={body}
            />
        </div>
    )
}
