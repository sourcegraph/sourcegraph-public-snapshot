import React, { useEffect, useState } from 'react'

import * as H from 'history'
import { Redirect, RouteComponentProps } from 'react-router-dom'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { Alert, Link, LoadingSpinner } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../auth'
import { HeroPage } from '../components/HeroPage'
import { PageTitle } from '../components/PageTitle'
import { SourcegraphContext } from '../jscontext'
import { eventLogger } from '../tracking/eventLogger'

import { SourcegraphIcon } from './icons'
import { getReturnTo } from './SignInSignUpCommon'

import unlockAccountStyles from './SignInSignUpCommon.module.scss'

interface UnlockAccountPageProps extends RouteComponentProps<{ token: string }> {
    location: H.Location
    authenticatedUser: AuthenticatedUser | null
    context: Pick<
        SourcegraphContext,
        'allowSignup' | 'authProviders' | 'sourcegraphDotComMode' | 'xhrHeaders' | 'resetPasswordEnabled'
    >
}

export const UnlockAccountPage: React.FunctionComponent<React.PropsWithChildren<UnlockAccountPageProps>> = props => {
    const [error, setError] = useState<Error | null>(null)
    const [loading, setLoading] = useState(true)
    const { token } = props.match.params

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

            eventLogger.log('OkUnlockAccount', { token })
        } catch (error) {
            setError(error)
            eventLogger.log('KoUnlockAccount', { token })
        } finally {
            setLoading(false)
        }
    }, [token, props.context.xhrHeaders])

    useEffect(() => {
        if (props.authenticatedUser) {
            return
        }
        eventLogger.logPageView('UnlockUserAccountRequest', null, false)
        unlockAccount().catch(error => {
            setError(error)
        })
    }, [unlockAccount, props.authenticatedUser])

    if (props.authenticatedUser) {
        const returnTo = getReturnTo(props.location)
        return <Redirect to={returnTo} />
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
                        ? 'Unlock your Sourcegraph Cloud account'
                        : 'Unlock your Sourcegraph Server account'
                }
                body={body}
            />
        </div>
    )
}
