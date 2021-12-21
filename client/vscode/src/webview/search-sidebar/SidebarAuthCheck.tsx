import classNames from 'classnames'
import React, { useEffect, useState } from 'react'
import { Form } from 'reactstrap'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'

import { WebviewPageProps } from '../platform/context'

import styles from './SearchSidebar.module.scss'

interface SidebarAuthCheckProps extends Pick<WebviewPageProps, 'sourcegraphVSCodeExtensionAPI'> {
    className?: string
    hasAccessToken: boolean
    onSubmitAccessToken: React.FormEventHandler<HTMLFormElement>
    validAccessToken: boolean
}

export const SidebarAuthCheck: React.FunctionComponent<SidebarAuthCheckProps> = ({
    sourcegraphVSCodeExtensionAPI,
    className,
    hasAccessToken,
    onSubmitAccessToken,
    validAccessToken,
}) => {
    // `undefined` while waiting for Comlink response.
    const [instanceHostname, setInstanceHostname] = useState<string | undefined>(undefined)
    const [signUpUrl, setSignUpUrl] = useState<string>('https://sourcegraph.com/sign-up')
    const [signInUrl, setSignInUrl] = useState<string>('https://sourcegraph.com/sign-in')
    const [hasAccount, setHasAccount] = useState(false)
    const [validating, setValidating] = useState(true)

    useEffect(() => {
        setValidating(true)
        if (instanceHostname === undefined) {
            sourcegraphVSCodeExtensionAPI
                .getInstanceHostname()
                .then(instance => {
                    setInstanceHostname(instance)
                    setSignUpUrl(new URL('sign-in', instance).href)
                    setSignInUrl(new URL('sign-in', instance).href)
                })
                // TODO error handling
                .catch(() => {})
        }
        setValidating(false)
    }, [sourcegraphVSCodeExtensionAPI, instanceHostname])

    return (
        <div className={classNames('d-flex flex-column align-items-left justify-content-center', className)}>
            <p className={classNames('mt-3', styles.title)}>Search Your Private Code</p>
            {validating && <LoadingSpinner />}
            {!hasAccessToken && !validAccessToken && !validating && (
                <>
                    {!hasAccount ? (
                        <div>
                            <p className={classNames('my-3', styles.text)}>
                                Create an account to enhance search across your private repositories: search multiple
                                repos & commit history, monitor, save searches, and more.
                            </p>
                            <a
                                href={signUpUrl}
                                className={classNames('btn btn-sm w-100 border-0 font-weight-normal', styles.button)}
                                onClick={() => setHasAccount(true)}
                            >
                                <span className={classNames('my-3', styles.text)}>Create an account</span>
                            </a>
                            <button
                                type="button"
                                className={classNames('my-3 btn btn-link', styles.text)}
                                onClick={() => setHasAccount(true)}
                            >
                                Have an account?
                            </button>
                        </div>
                    ) : (
                        // eslint-disable-next-line react/forbid-elements
                        <Form onSubmit={onSubmitAccessToken}>
                            <p className={classNames('my-3', styles.text)}>
                                Sign in by entering an access token created through your{' '}
                                <a href={signInUrl} onClick={() => setHasAccount(true)}>
                                    user setting
                                </a>{' '}
                                on {instanceHostname}.
                            </p>
                            <p className={classNames('my-3', styles.text)}>
                                See our{' '}
                                <a href="https://docs.sourcegraph.com/cli/how-tos/creating_an_access_token">
                                    user docs
                                </a>{' '}
                                for a video guide on how to create an access token.
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
                </>
            )}

            {hasAccessToken && !validating && (
                <>
                    {!validAccessToken ? (
                        <Form onSubmit={onSubmitAccessToken}>
                            <a
                                href={signInUrl}
                                className="btn btn-sm btn-danger w-100 border-0 font-weight-normal"
                                onClick={() => setHasAccount(true)}
                            >
                                <span className={classNames('my-3', styles.text)}>
                                    ERROR: Unable to verify your Access Token for {instanceHostname}. Please try with a
                                    new Access Token and add CORS if you are currently on VS Code Web.
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
                                className={classNames(
                                    'btn btn-sm btn-link w-100 border-0 font-weight-normal',
                                    styles.button
                                )}
                            >
                                <span className={classNames('my-0', styles.text)}>Update Access Token</span>
                            </button>
                        </Form>
                    ) : (
                        <button
                            type="button"
                            onClick={() => sourcegraphVSCodeExtensionAPI.openSearchPanel()}
                            className={classNames(
                                'mb-3 btn btn-sm w-100 border-0 font-weight-normal disabled',
                                styles.button
                            )}
                        >
                            Access Token Verified!
                        </button>
                    )}
                </>
            )}
        </div>
    )
}
