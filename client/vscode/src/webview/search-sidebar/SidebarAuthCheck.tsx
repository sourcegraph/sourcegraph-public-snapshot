import classNames from 'classnames'
import React, { useEffect, useState } from 'react'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { currentAuthStateQuery } from '@sourcegraph/shared/src/auth'
import { CurrentAuthStateResult, CurrentAuthStateVariables } from '@sourcegraph/shared/src/graphql-operations'

import { WebviewPageProps } from '../platform/context'

import styles from './OpenSearchPanelCta.module.scss'

interface OpenSearchPanelCtaProps extends Pick<WebviewPageProps, 'platformContext' | 'sourcegraphVSCodeExtensionAPI'> {
    className?: string
}

export const SidebarAuthCheck: React.FunctionComponent<OpenSearchPanelCtaProps> = ({
    sourcegraphVSCodeExtensionAPI,
    platformContext,
    className,
}) => {
    // `undefined` while waiting for Comlink response.
    const [hasAccessToken, setHasAccessToken] = useState<boolean | undefined>(undefined)
    const [instanceHostname, setInstanceHostname] = useState<string | undefined>(undefined)

    useEffect(() => {
        sourcegraphVSCodeExtensionAPI
            .hasAccessToken()
            .then(hasAccessToken => setHasAccessToken(hasAccessToken))
            // TODO error handling
            .catch(() => setHasAccessToken(false))

        sourcegraphVSCodeExtensionAPI
            .getInstanceHostname()
            .then(instanceHostname => setInstanceHostname(instanceHostname))
            // TODO error handling
            .catch(() => setInstanceHostname('https://sourcegraph.com'))
    }, [sourcegraphVSCodeExtensionAPI])

    const [hasAccount, setHasAccount] = useState(false)

    const [validating, setValidating] = useState(false)

    // On submit, validate access token and update VS Code settings through API.
    // Open search tab on successful validation.
    const onSubmitAccessToken: React.FormEventHandler<HTMLFormElement> = event => {
        event?.preventDefault()
        ;(async () => {
            const accessToken = event.currentTarget.elements.token.value

            if (!validating && accessToken) {
                setValidating(true)
                // TODO set loading state
                await sourcegraphVSCodeExtensionAPI.updateAccessToken(accessToken)
                const currentUser = await platformContext
                    .requestGraphQL<CurrentAuthStateResult, CurrentAuthStateVariables>({
                        request: currentAuthStateQuery,
                        variables: {},
                        mightContainPrivateInfo: true,
                    })
                    .toPromise()

                if (currentUser.data) {
                    // TODO Tell the VS Code extension that the access token is valid, open search tab
                }
                setValidating(false)
            }
        })().catch(error => {
            setValidating(false)
            console.error(error)
        })
    }

    if (hasAccessToken === undefined || instanceHostname === undefined) {
        return null
    }

    if (validating) {
        // TODO better loading state
        return <LoadingSpinner />
    }

    const signUpUrl = new URL('/sign-up', instanceHostname)
    const signInUrl = new URL('/sign-up', instanceHostname)

    return (
        <div className={classNames('d-flex flex-column align-items-left justify-content-center', className)}>
            <p className={classNames('mt-3', styles.title)}>Search Your Private Code</p>
            {!hasAccount ? (
                <div>
                    <p className={classNames('my-3', styles.text)}>
                        Create an account to enhance search across your private repositories: search multiple repos &
                        commit history, monitor, save searches, and more.
                    </p>
                    <a
                        href={signUpUrl.href}
                        className={classNames('btn btn-sm w-100 border-0 font-weight-normal', styles.button)}
                        onClick={() => setHasAccount(true)}
                    >
                        <span className={classNames('my-3', styles.text)}>Create an account</span>
                    </a>
                    <p className={classNames('my-3', styles.text)}>
                        <a href={signInUrl.href} onClick={() => setHasAccount(true)}>
                            Have an account?
                        </a>
                    </p>
                </div>
            ) : (
                // eslint-disable-next-line react/forbid-elements
                <form onSubmit={onSubmitAccessToken}>
                    <p className={classNames('my-3', styles.text)}>
                        Sign in by entering an access token created through your user setting on sourcegraph.com.
                    </p>
                    <p className={classNames('my-3', styles.text)}>
                        See our{' '}
                        <a href="https://docs.sourcegraph.com/cli/how-tos/creating_an_access_token">user docs</a> for a
                        video guide on how to create an access token.
                    </p>
                    <input
                        className={classNames('my-3 w-100 p-1', styles.text)}
                        type="text"
                        name="token"
                        placeholder="ex 6dfc880b320dff712d9f6cfcac5cbd13ebfad1d8"
                    />
                    <button
                        type="submit"
                        className={classNames('btn btn-sm btn-link w-100 border-0 font-weight-normal', styles.button)}
                    >
                        <span className={classNames('my-3', styles.text)}>Enter Access Token</span>
                    </button>
                    <p className={classNames('my-3', styles.text)}>
                        <a href={signUpUrl.href}>Create an account</a>
                    </p>
                </form>
            )}
        </div>
    )
}
