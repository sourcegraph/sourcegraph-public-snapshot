import React, { useEffect, useState } from 'react'

import classNames from 'classnames'
import * as H from 'history'
import { Redirect, RouteComponentProps } from 'react-router-dom'

// import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
// import { Button, Link, Alert, Icon } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../auth'
import { HeroPage } from '../components/HeroPage'
import { PageTitle } from '../components/PageTitle'
import { SourcegraphContext } from '../jscontext'
import { eventLogger } from '../tracking/eventLogger'

import { SourcegraphIcon } from './icons'
import { getReturnTo } from './SignInSignUpCommon'

import unlockAccountStyles from './SignInSignUpCommon.module.scss'
import { Button } from '@sourcegraph/wildcard'

interface UnlockAccountPageProps extends RouteComponentProps<{ token: string; userId: string }> {
    location: H.Location
    authenticatedUser: AuthenticatedUser | null
    context: Pick<
        SourcegraphContext,
        'allowSignup' | 'authProviders' | 'sourcegraphDotComMode' | 'xhrHeaders' | 'resetPasswordEnabled'
    >
}

export const UnlockAccountPage: React.FunctionComponent<UnlockAccountPageProps> = props => {
    useEffect(() => eventLogger.logViewEvent('unlockaccount', null, false))

    const [error, setError] = useState<Error | null>(null)

    const unlockAccount = async (): Promise<void> => {
        await fetch('/-/unlock-account', {
            credentials: 'same-origin',
            method: 'POST',
            headers: {
                ...props.context.xhrHeaders,
                Accept: 'application/json',
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                token: props.match.params.token,
                userId: Number(props.match.params.userId),
            }),
        })
    }

    if (props.authenticatedUser) {
        const returnTo = getReturnTo(props.location)
        return <Redirect to={returnTo} />
    }

    const body = (
        <div>
            <Button onClick={unlockAccount}>unlock</Button>
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
                        ? 'Sign in to Sourcegraph Cloud'
                        : 'Sign in to Sourcegraph Server'
                }
                body={body}
            />
        </div>
    )
}
