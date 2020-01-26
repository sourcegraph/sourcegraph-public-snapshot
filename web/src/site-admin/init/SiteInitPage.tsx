import React, { useCallback } from 'react'
import { Redirect } from 'react-router'
import * as GQL from '../../../../shared/src/graphql/schema'
import { SignUpArgs, SignUpForm } from '../../auth/SignUpForm'
import { BrandLogo } from '../../components/branding/BrandLogo'
import { ThemeProps } from '../../../../shared/src/theme'

interface Props extends ThemeProps {
    authenticatedUser: Pick<GQL.IUser, 'username'> | null

    /**
     * Whether site initialization is needed. If not set, the global value from
     * `window.context.needsSiteInit` is used.
     */
    needsSiteInit?: typeof window.context.needsSiteInit
}

/**
 * A page that is shown when the Sourcegraph instance has not yet been initialized.
 * Only the person who first accesses the instance will see this.
 */
export const SiteInitPage: React.FunctionComponent<Props> = ({
    authenticatedUser,
    needsSiteInit = window.context.needsSiteInit,
    isLightTheme,
}) => {
    const doSiteInit = useCallback(
        (args: SignUpArgs): Promise<void> =>
            fetch('/-/site-init', {
                credentials: 'same-origin',
                method: 'POST',
                headers: {
                    ...window.context.xhrHeaders,
                    Accept: 'application/json',
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(args),
            }).then(resp => {
                if (resp.status !== 200) {
                    return resp.text().then(text => Promise.reject(new Error(text)))
                }

                // Force hard-reload because site initialization changes global window.context
                // values that our application can't react to without a hard reload.
                window.location.replace('/site-admin')

                return Promise.resolve()
            }),
        []
    )

    if (!needsSiteInit) {
        return <Redirect to="/search" />
    }

    return (
        <div className="site-init-page">
            <div className="site-init-page__content">
                <BrandLogo className="site-init-page__logo" isLightTheme={isLightTheme} />
                {authenticatedUser ? (
                    // If there's already a user but the site is not initialized, then the we're in an
                    // unexpected state, likely because of a previous bug or because someone manually modified
                    // the site_config DB table.
                    <p>
                        You're signed in as <strong>{authenticatedUser.username}</strong>. A site admin must initialize
                        Sourcegraph before you can continue.
                    </p>
                ) : (
                    <>
                        <h2 className="site-init-page__header">Welcome</h2>
                        <p>Create an admin account to start using Sourcegraph.</p>
                        <SignUpForm buttonLabel="Create admin account & continue" doSignUp={doSiteInit} />
                    </>
                )}
            </div>
        </div>
    )
}
