import React, { useEffect, useState } from 'react'

import { mdiBitbucket, mdiChevronLeft, mdiGithub, mdiGitlab, mdiKeyVariant, mdiMicrosoftAzureDevops } from '@mdi/js'
import classNames from 'classnames'
import { partition } from 'lodash'
import { Navigate, useLocation, useSearchParams } from 'react-router-dom'

import { Alert, Icon, Text, Link, Button, ErrorAlert, AnchorLink } from '@sourcegraph/wildcard'

import type { AuthenticatedUser } from '../auth'
import { HeroPage } from '../components/HeroPage'
import { PageTitle } from '../components/PageTitle'
import type { AuthProvider, SourcegraphContext } from '../jscontext'
import { eventLogger } from '../tracking/eventLogger'
import { checkRequestAccessAllowed } from '../util/checkRequestAccessAllowed'

import { SourcegraphIcon } from './icons'
import { OrDivider } from './OrDivider'
import { getReturnTo } from './SignInSignUpCommon'
import { UsernamePasswordSignInForm } from './UsernamePasswordSignInForm'

import signInSignUpCommonStyles from './SignInSignUpCommon.module.scss'

export interface SignInPageProps {
    authenticatedUser: AuthenticatedUser | null
    context: Pick<
        SourcegraphContext,
        | 'allowSignup'
        | 'authProviders'
        | 'sourcegraphDotComMode'
        | 'primaryLoginProvidersCount'
        // needed for checkRequestAccessAllowed
        | 'authAccessRequest'
        // needed for UsernamePasswordSignInForm
        | 'xhrHeaders'
        | 'resetPasswordEnabled'
    >
}

export const SignInPage: React.FunctionComponent<React.PropsWithChildren<SignInPageProps>> = props => {
    const { context, authenticatedUser } = props
    useEffect(() => {
        window.context.telemetryRecorder?.recordEvent('signIn', 'viewed')
        eventLogger.logViewEvent('SignIn', null, false)
    })

    const location = useLocation()
    const [error, setError] = useState<Error | null>(null)
    const [searchParams, setSearchParams] = useSearchParams()
    const isRequestAccessAllowed = checkRequestAccessAllowed(props.context)

    const returnTo = getReturnTo(location)
    if (authenticatedUser) {
        return <Navigate to={returnTo} replace={true} />
    }

    const [[builtInAuthProvider], nonBuiltinAuthProviders] = partition(
        context.authProviders,
        provider => provider.isBuiltin
    )

    const shouldShowProvider = function (provider: AuthProvider): boolean {
        // Hide the Sourcegraph Operator authentication provider by default because it is
        // not useful to customer users and may even cause confusion.
        if (provider.serviceType === 'sourcegraph-operator') {
            return searchParams.has('sourcegraph-operator')
        }
        if (provider.serviceType === 'gerrit') {
            return false
        }
        return true
    }

    const toggleMoreProviders = (showMore: boolean): void => {
        const param = 'showMore'
        if (showMore) {
            searchParams.set(param, '')
        } else {
            searchParams.delete(param)
        }
        setSearchParams(searchParams)
    }

    const thirdPartyAuthProviders = nonBuiltinAuthProviders.filter(provider => shouldShowProvider(provider))
    const primaryProviders = thirdPartyAuthProviders.slice(0, context.primaryLoginProvidersCount)
    const moreProviders = thirdPartyAuthProviders.slice(context.primaryLoginProvidersCount)

    const showMoreProviders = searchParams.has('showMore') && (builtInAuthProvider || moreProviders.length > 0)
    const hasProviders = builtInAuthProvider || thirdPartyAuthProviders.length > 0
    const showMoreWaysToLogin =
        !showMoreProviders && (moreProviders.length > 0 || (primaryProviders.length > 0 && builtInAuthProvider))

    const providers = showMoreProviders ? moreProviders : primaryProviders

    const body = !hasProviders ? (
        <Alert className="mt-3" variant="info">
            No authentication providers are available. Contact a site administrator for help.
        </Alert>
    ) : (
        <div className={classNames('mb-4 pb-5', signInSignUpCommonStyles.signinPageContainer)}>
            {error && <ErrorAlert className="mt-4 mb-0 text-left" error={error} />}
            <div
                className={classNames(
                    'test-signin-form rounded p-4 my-3',
                    signInSignUpCommonStyles.signinSignupForm,
                    error ? 'mt-3' : 'mt-4'
                )}
            >
                {showMoreProviders && (
                    <div className="mb-3 text-left">
                        <Button
                            variant="link"
                            className="p-0 border-0 font-weight-normal"
                            onClick={() => toggleMoreProviders(false)}
                        >
                            <Icon aria-hidden={true} svgPath={mdiChevronLeft} />
                            Back
                        </Button>
                    </div>
                )}
                {builtInAuthProvider && (showMoreProviders || thirdPartyAuthProviders.length === 0) && (
                    <UsernamePasswordSignInForm
                        {...props}
                        onAuthError={setError}
                        className={classNames({ 'mb-3': providers.length > 0 })}
                    />
                )}
                {builtInAuthProvider && showMoreProviders && providers.length > 0 && (
                    <OrDivider className="mb-3 py-1" />
                )}
                {providers.map((provider, index) => (
                    // Use index as key because display name may not be unique. This is OK
                    // here because this list will not be updated during this component's lifetime.
                    /* eslint-disable react/no-array-index-key */
                    <div className="mb-2" key={index}>
                        <Button
                            to={provider.authenticationURL}
                            display="block"
                            variant={showMoreProviders ? 'secondary' : 'primary'}
                            as={AnchorLink}
                        >
                            {provider.serviceType === 'github' && <Icon aria-hidden={true} svgPath={mdiGithub} />}
                            {provider.serviceType === 'gitlab' && <Icon aria-hidden={true} svgPath={mdiGitlab} />}
                            {provider.serviceType === 'bitbucketCloud' && (
                                <Icon aria-hidden={true} svgPath={mdiBitbucket} />
                            )}
                            {provider.serviceType === 'azuredevops' && (
                                <Icon aria-hidden={true} svgPath={mdiMicrosoftAzureDevops} />
                            )}{' '}
                            {provider.displayPrefix ?? 'Continue with'} {provider.displayName}
                        </Button>
                    </div>
                ))}
                {showMoreWaysToLogin && (
                    <div className="mb-2">
                        <Button display="block" variant="secondary" onClick={() => toggleMoreProviders(true)}>
                            <Icon aria-hidden={true} svgPath={mdiKeyVariant} /> Other login methods
                        </Button>
                    </div>
                )}
            </div>
            {context.allowSignup ? (
                <Text>
                    New to Sourcegraph? <Link to="/sign-up">Sign up.</Link>{' '}
                    {context.sourcegraphDotComMode && (
                        <>
                            To use Sourcegraph on private repositories,{' '}
                            <Link
                                to="https://about.sourcegraph.com/app"
                                onClick={() => {
                                    window.context.telemetryRecorder?.recordEvent('appCta.signInPage', 'clicked')
                                    eventLogger.log('ClickedOnAppCTA', { location: 'SignInPage' })
                                }}
                            >
                                download Cody app
                            </Link>{' '}
                            or{' '}
                            <Link
                                to="https://sourcegraph.com/get-started?t=enterprise"
                                onClick={() => {
                                    window.context.telemetryRecorder?.recordEvent('enterpriseCta.signInPage', 'clicked')
                                    eventLogger.log('ClickedOnEnterpriseCTA', { location: 'SignInPage' })
                                }}
                            >
                                get Sourcegraph Enterprise
                            </Link>
                            .
                        </>
                    )}
                </Text>
            ) : isRequestAccessAllowed ? (
                <Text className="text-muted">
                    Need an account? <Link to="/request-access">Request access</Link> or contact your site admin.
                </Text>
            ) : (
                <Text className="text-muted">Need an account? Contact your site admin.</Text>
            )}
        </div>
    )

    return (
        <div className={signInSignUpCommonStyles.signinSignupPage}>
            <PageTitle title="Sign in" />
            <HeroPage
                icon={SourcegraphIcon}
                iconLinkTo={context.sourcegraphDotComMode ? '/search' : undefined}
                iconClassName="bg-transparent"
                lessPadding={true}
                title="Sign in to Sourcegraph"
                body={body}
            />
        </div>
    )
}
