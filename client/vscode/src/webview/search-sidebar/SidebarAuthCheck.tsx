import classNames from 'classnames'
import React, { useEffect, useState } from 'react'
import { Form } from 'reactstrap'

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
    const [validAccessToken, setValidAccessToken] = useState<boolean | undefined>(undefined)
    const [signUpUrl, setSignUpUrl] = useState<string>('https://sourcegraph.com/sign-up')
    const [signInUrl, setSignInUrl] = useState<string>('https://sourcegraph.com/sign-in')
    const [hasAccount, setHasAccount] = useState(false)
    const [validating, setValidating] = useState(true)

    useEffect(() => {
        setValidating(true)
        if (hasAccessToken === undefined) {
            sourcegraphVSCodeExtensionAPI
                .getInstanceHostname()
                .then(instance => {
                    setInstanceHostname(instance)
                    setSignUpUrl(new URL('sign-in', instance).href)
                    setSignInUrl(new URL('sign-in', instance).href)
                })
                // TODO error handling
                .catch(() => {})
            sourcegraphVSCodeExtensionAPI
                .hasAccessToken()
                .then(hasAccessToken => {
                    setHasAccessToken(hasAccessToken)
                    setHasAccount(true)
                })
                // TODO error handling
                .catch(() => setHasAccessToken(false))
        }
        if (hasAccessToken && hasAccount && instanceHostname) {
            ;(async () => {
                const currentUser = await platformContext
                    .requestGraphQL<CurrentAuthStateResult, CurrentAuthStateVariables>({
                        request: currentAuthStateQuery,
                        variables: {},
                        mightContainPrivateInfo: true,
                    })
                    .toPromise()
                if (currentUser.data) {
                    console.log(currentUser)
                    setValidAccessToken(true)
                    await sourcegraphVSCodeExtensionAPI.openSearchPanel()
                } else {
                    console.log(currentUser)
                    setValidAccessToken(false)
                }
            })().catch(error => console.error(error))
        }
        setValidating(false)
    }, [sourcegraphVSCodeExtensionAPI, hasAccessToken, hasAccount, instanceHostname, platformContext])

    // On submit, validate access token and update VS Code settings through API.
    // Open search tab on successful validation.
    const onSubmitAccessToken: React.FormEventHandler<HTMLFormElement> = event => {
        event?.preventDefault()
        setValidating(true)
        setValidAccessToken(false)
        ;(async () => {
            const accessToken = (event.currentTarget.elements.namedItem('token') as HTMLInputElement).value

            if (!validating && accessToken) {
                await sourcegraphVSCodeExtensionAPI.updateAccessToken(accessToken)
                // Updating below states  would call useEffect to validate the updated token
                setHasAccessToken(true)
                setHasAccount(true)
                setValidAccessToken(true)
            }
        })().catch(error => {
            console.error(error)
        })
        setValidating(false)
    }

    return (
        <div className={classNames('d-flex flex-column align-items-left justify-content-center', className)}>
            <p className={classNames('mt-3', styles.title)}>Search Your Private Code</p>
            {validating && <LoadingSpinner />}
            {!hasAccount && !hasAccessToken && instanceHostname && !validating && (
                <div>
                    <p className={classNames('my-3', styles.text)}>
                        Create an account to enhance search across your private repositories: search multiple repos &
                        commit history, monitor, save searches, and more.
                    </p>
                    <a
                        href={signUpUrl}
                        className={classNames('btn btn-sm w-100 border-0 font-weight-normal', styles.button)}
                        onClick={() => setHasAccount(true)}
                    >
                        <span className={classNames('my-3', styles.text)}>Create an account</span>
                    </a>
                    <p className={classNames('my-3', styles.text)}>
                        <a href={signInUrl} onClick={() => setHasAccount(true)}>
                            Have an account?
                        </a>
                    </p>
                </div>
            )}
            {hasAccount && !hasAccessToken && instanceHostname && !validating && (
                // eslint-disable-next-line react/forbid-elements
                <Form onSubmit={onSubmitAccessToken}>
                    <p className={classNames('my-3', styles.text)}>
                        Sign in by entering an access token created through your user setting on sourcegraph.com.
                    </p>
                    <p className={classNames('my-3', styles.text)}>
                        See our{' '}
                        <a href="https://docs.sourcegraph.com/cli/how-tos/creating_an_access_token">user docs</a> for a
                        video guide on how to create an access token.
                    </p>
                    <input
                        className="input form-control"
                        type="text"
                        name="token"
                        placeholder="ex 6dfc880b320dff712d9f6cfcac5cbd13ebfad1d8"
                    />
                    <button
                        type="submit"
                        className={classNames(
                            'btn btn-sm btn-link w-100 border-0 font-weight-normal my-3',
                            styles.button
                        )}
                    >
                        <span className={classNames('my-0', styles.text)}>Enter Access Token</span>
                    </button>
                    <p className={classNames('my-3', styles.text)}>
                        <a href={signUpUrl}>Create an account</a>
                    </p>
                </Form>
            )}
            {hasAccessToken && validAccessToken ? (
                <button
                    type="button"
                    onClick={() => sourcegraphVSCodeExtensionAPI.openSearchPanel()}
                    className={classNames('mb-3 btn btn-sm w-100 border-0 font-weight-normal', styles.button)}
                >
                    Access Token Verified! Click here to start searching!
                </button>
            ) : (
                <Form onSubmit={onSubmitAccessToken}>
                    <a
                        href={signInUrl}
                        className="btn btn-sm btn-danger w-100 border-0 font-weight-normal"
                        onClick={() => setHasAccount(true)}
                    >
                        <span className={classNames('my-3', styles.text)}>
                            ERROR: Failed to verify your Access Token for {instanceHostname}. Please try with a new
                            Access Token or use CORS if you are currently on VS Code Web.
                        </span>
                    </a>
                    <input
                        className="input form-control my-3"
                        type="text"
                        name="token"
                        placeholder="ex 6dfc880b320dff712d9f6cfcac5cbd13ebfad1d8"
                    />
                    <button
                        type="submit"
                        className={classNames('btn btn-sm btn-link w-100 border-0 font-weight-normal', styles.button)}
                    >
                        <span className={classNames('my-0', styles.text)}>Update Access Token</span>
                    </button>
                </Form>
            )}
        </div>
    )
}
