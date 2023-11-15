import React from 'react'

import { Navigate } from 'react-router-dom'

import { logger } from '@sourcegraph/common'
import { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import { Text, Container } from '@sourcegraph/wildcard'

import type { AuthenticatedUser } from '../../auth'
import { AuthPageWrapper } from '../../auth/AuthPageWrapper'
import { type SignUpArguments, SignUpForm } from '../../auth/SignUpForm'
import { PageTitle } from '../../components/PageTitle'
import type { SourcegraphContext } from '../../jscontext'
import { PageRoutes } from '../../routes.constants'

import styles from './SiteInitPage.module.scss'

const initSite = async (args: SignUpArguments): Promise<void> => {
    const pingUrl = new URL('https://sourcegraph.com/ping-from-self-hosted')
    pingUrl.searchParams.set('email', args.email)
    pingUrl.searchParams.set('tos_accepted', 'true') // Terms of Service are required to be accepted

    await fetch(pingUrl.toString(), {
        credentials: 'include',
    })
        .then() // no op
        .catch((error): void => {
            logger.error(error)
        })
    const response = await fetch('/-/site-init', {
        credentials: 'same-origin',
        method: 'POST',
        headers: {
            ...window.context.xhrHeaders,
            Accept: 'application/json',
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(args),
    })
    if (response.status !== 200) {
        const text = await response.text()
        throw new Error(text)
    }
    window.location.replace('/site-admin')
}

interface Props extends TelemetryV2Props {
    authenticatedUser: Pick<AuthenticatedUser, 'username'> | null

    /**
     * Whether site initialization is needed. If not set, the global value from
     * `window.context.needsSiteInit` is used.
     */
    needsSiteInit?: typeof window.context.needsSiteInit
    context: Pick<SourcegraphContext, 'authPasswordPolicy' | 'authMinPasswordLength'>
}

/**
 * A page that is shown when the Sourcegraph instance has not yet been initialized.
 * Only the person who first accesses the instance will see this.
 */
export const SiteInitPage: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    authenticatedUser,
    needsSiteInit = window.context.needsSiteInit,
    context,
    telemetryRecorder,
}) => {
    // This page is never shown on dotcom, to keep the API surface
    // of this component clean, we don't expose this option.
    const sourcegraphDotComMode = false

    if (!needsSiteInit) {
        return <Navigate to={PageRoutes.Search} replace={true} />
    }

    return (
        <>
            <PageTitle title="Site initialization" />
            <AuthPageWrapper
                title="Welcome to Sourcegraph"
                description="Create an admin account to get started"
                sourcegraphDotComMode={sourcegraphDotComMode}
                className={styles.wrapper}
            >
                {authenticatedUser ? (
                    // If there's already a user but the site is not initialized, then the we're in an
                    // unexpected state, likely because of a previous bug or because someone manually modified
                    // the site_config DB table.
                    <Container>
                        <Text className="mb-0">
                            You're signed in as <strong>{authenticatedUser.username}</strong>. A site admin must
                            initialize Sourcegraph before you can continue.
                        </Text>
                    </Container>
                ) : (
                    <Container>
                        <SignUpForm
                            className="w-100"
                            buttonLabel="Create admin account and continue"
                            onSignUp={initSite}
                            // This page is never shown on dotcom, to keep the API surface
                            // of this component clean, we don't expose this option.
                            context={{ ...context, sourcegraphDotComMode, authProviders: [] }}
                            telemetryRecorder={telemetryRecorder}
                        />
                    </Container>
                )}
            </AuthPageWrapper>
        </>
    )
}
