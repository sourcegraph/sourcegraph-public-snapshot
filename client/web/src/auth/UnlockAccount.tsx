import React, { useEffect, useState } from 'react'

import { Navigate, useLocation, useParams } from 'react-router-dom'

import { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import { Alert, Link, LoadingSpinner, ErrorAlert } from '@sourcegraph/wildcard'

import type { AuthenticatedUser } from '../auth'
import { HeroPage } from '../components/HeroPage'
import { PageTitle } from '../components/PageTitle'
import type { SourcegraphContext } from '../jscontext'
import { eventLogger } from '../tracking/eventLogger'

import { SourcegraphIcon } from './icons'
import { getReturnTo } from './SignInSignUpCommon'

import unlockAccountStyles from './SignInSignUpCommon.module.scss'

interface UnlockAccountPageProps extends TelemetryV2Props {
    authenticatedUser: AuthenticatedUser | null
    context: Pick<
        SourcegraphContext,
        'allowSignup' | 'authProviders' | 'sourcegraphDotComMode' | 'xhrHeaders' | 'resetPasswordEnabled'
    >
}

export const UnlockAccountPage: React.FunctionComponent<React.PropsWithChildren<UnlockAccountPageProps>> = props => {
    const [error, setError] = useState<Error | null>(null)
    const [loading, setLoading] = useState(true)

    const location = useLocation()
    const { token } = useParams()

    const unlockAccount = React.useCallback(async (): Promise<void> => {
        try {
            setLoading(true)
            const response = await fetch('/-/unlock-account', {
                credentials: 'same-origin',
                method: 'POST',
                headers: {
                    ...props.context.xhrHeaders,
                    Accept: 'application/json',
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    token,
                }),
            })

            if (!response.ok) {
                throw new Error('The url you provided is either expired or invalid.')
            }

            props.telemetryRecorder.recordEvent('okUnlockAccount', 'succeeded')
            eventLogger.log('unlockAccount', { token })
        } catch (error) {
            setError(error)
            props.telemetryRecorder.recordEvent('koUnlockAccount', 'failed')
            eventLogger.log('unlockAccount', { token })
        } finally {
            setLoading(false)
        }
    }, [token, props.context.xhrHeaders, props.telemetryRecorder])

    useEffect(() => {
        if (props.authenticatedUser) {
            return
        }
        props.telemetryRecorder.recordEvent('unlockUserAccountRequest', 'viewed')
        eventLogger.logPageView('UnlockUserAccountRequest', null, false)
        unlockAccount().catch(error => {
            setError(error)
        })
    }, [unlockAccount, props.authenticatedUser, props.telemetryRecorder])

    if (props.authenticatedUser) {
        const returnTo = getReturnTo(location)
        return <Navigate to={returnTo} replace={true} />
    }

    const body = (
        <div>
            {loading && <LoadingSpinner />}
            {error && <ErrorAlert className="mt-2" error={error} />}
            {!loading && !error && (
                <>
                    <Alert variant="success">
                        Your account was unlocked. Please try to <Link to="/sign-in">sign in</Link> to continue.
                    </Alert>
                </>
            )}
        </div>
    )

    return (
        <div className={unlockAccountStyles.signinSignupPage}>
            <PageTitle title="Unlock account" />
            <HeroPage
                icon={SourcegraphIcon}
                iconLinkTo={props.context.sourcegraphDotComMode ? '/search' : undefined}
                iconClassName="bg-transparent"
                lessPadding={true}
                title={
                    props.context.sourcegraphDotComMode
                        ? 'Unlock your Sourcegraph.com account'
                        : 'Unlock your Sourcegraph Server account'
                }
                body={body}
            />
        </div>
    )
}
