import React, { useEffect } from 'react'

import { Navigate, useLocation } from 'react-router-dom'

import { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import type { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { useIsLightTheme } from '@sourcegraph/shared/src/theme'
import { Container, Link, Text } from '@sourcegraph/wildcard'

import type { AuthenticatedUser } from '../auth'
import { PageTitle } from '../components/PageTitle'
import type { SourcegraphContext } from '../jscontext'
import { PageRoutes } from '../routes.constants'

// import { eventLogger } from '../tracking/eventLogger'

import { AuthPageWrapper } from './AuthPageWrapper'
import { CloudSignUpPage, ShowEmailFormQueryParameter } from './CloudSignUpPage'
import { getReturnTo } from './SignInSignUpCommon'
import { type SignUpArguments, SignUpForm } from './SignUpForm'
import { VsCodeSignUpPage } from './VsCodeSignUpPage'

import styles from './SignUpPage.module.scss'

export interface SignUpPageProps extends TelemetryProps, TelemetryV2Props {
    authenticatedUser: AuthenticatedUser | null
    context: Pick<
        SourcegraphContext,
        | 'allowSignup'
        | 'authProviders'
        | 'sourcegraphDotComMode'
        | 'xhrHeaders'
        | 'authPasswordPolicy'
        | 'authMinPasswordLength'
    >
}

export const SignUpPage: React.FunctionComponent<React.PropsWithChildren<SignUpPageProps>> = ({
    authenticatedUser,
    context,
    telemetryService,
    telemetryRecorder,
}) => {
    const location = useLocation()
    const query = new URLSearchParams(location.search)
    const invitedBy = query.get('invitedBy')
    const returnTo = getReturnTo(location)

    const isLightTheme = useIsLightTheme()

    useEffect(() => {
        // eventLogger.logViewEvent('SignUp', null, false)
        telemetryRecorder.recordEvent('SignUp', 'viewed')

        if (invitedBy !== null) {
            const parameters = {
                isAuthenticated: !!authenticatedUser,
                allowSignup: context.allowSignup,
            }
            // eventLogger.log('SignUpInvitedByUser', parameters, parameters)
            telemetryRecorder.recordEvent('SignUpInvitedByUser', 'clicked', { privateMetadata: { parameters } })
        }
    }, [invitedBy, authenticatedUser, context.allowSignup, telemetryRecorder])

    if (authenticatedUser) {
        return <Navigate to={returnTo} replace={true} />
    }

    if (!context.allowSignup) {
        return <Navigate to="/sign-in" replace={true} />
    }

    const handleSignUp = (args: SignUpArguments): Promise<void> =>
        fetch('/-/sign-up', {
            credentials: 'same-origin',
            method: 'POST',
            headers: {
                ...context.xhrHeaders,
                Accept: 'application/json',
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(args),
        }).then(response => {
            if (response.status !== 200) {
                return response.text().then(text => Promise.reject(new Error(text)))
            }

            // Redirects to the /post-sign-up after successful signup on sourcegraphDotCom.
            window.location.replace(context.sourcegraphDotComMode ? PageRoutes.PostSignUp : returnTo)

            return Promise.resolve()
        })

    if (query.get('editor') === 'vscode') {
        return (
            <VsCodeSignUpPage
                source={query.get('src')}
                onSignUp={handleSignUp}
                showEmailForm={query.has(ShowEmailFormQueryParameter)}
                context={context}
                telemetryService={telemetryService}
                telemetryRecorder={telemetryRecorder}
            />
        )
    }

    if (context.sourcegraphDotComMode) {
        return (
            <CloudSignUpPage
                source={query.get('src')}
                onSignUp={handleSignUp}
                isLightTheme={isLightTheme}
                showEmailForm={query.has(ShowEmailFormQueryParameter)}
                context={context}
                telemetryService={telemetryService}
                telemetryRecorder={telemetryRecorder}
                isSourcegraphDotCom={context.sourcegraphDotComMode}
            />
        )
    }

    return (
        <>
            <PageTitle title="Sign up" />
            <AuthPageWrapper
                title="Welcome to Sourcegraph"
                description={
                    context.sourcegraphDotComMode ? 'Sign up for Sourcegraph.com' : 'Sign up for Sourcegraph Server'
                }
                sourcegraphDotComMode={context.sourcegraphDotComMode}
                className={styles.wrapper}
            >
                {context.sourcegraphDotComMode && <Text className="pt-1 pb-2">Start searching public code now</Text>}
                <Container>
                    <SignUpForm telemetryRecorder={telemetryRecorder} context={context} onSignUp={handleSignUp} />
                </Container>
                <Text className="text-center mt-3">
                    Already have an account? <Link to={`/sign-in${location.search}`}>Sign in</Link>
                </Text>
            </AuthPageWrapper>
        </>
    )
}
