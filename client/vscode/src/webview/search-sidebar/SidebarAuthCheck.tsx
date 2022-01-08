import classNames from 'classnames'
import React, { useCallback, useEffect, useState } from 'react'
import { Form } from 'reactstrap'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { WebviewPageProps } from '../platform/context'

import styles from './SearchSidebar.module.scss'

interface SidebarAuthCheckProps extends TelemetryProps, Pick<WebviewPageProps, 'sourcegraphVSCodeExtensionAPI'> {
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
    telemetryService,
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

    const onSignUpClick = useCallback(
        (event?: React.FormEvent): void => {
            event?.preventDefault()
            setHasAccount(true)
            sourcegraphVSCodeExtensionAPI
                .openLink(signInUrl)
                .then(() => {})
                .catch(() => {})
            telemetryService.log('VSCESearchBarClicked', { campaign: 'Sign up link' }, { campaign: 'Sign up link' })
        },
        [signInUrl, sourcegraphVSCodeExtensionAPI, telemetryService]
    )

    return (
        <div className={classNames('d-flex flex-column align-items-left justify-content-center', className)}>
            <p className={classNames('mt-3', styles.title)}>Search Your Private Code</p>
            {validating && <LoadingSpinner />}
            {!validating && !hasAccessToken && (
                <>
                    {!hasAccount ? (
                        <div>
                            <p className={classNames('my-3', styles.text)}>
                                Create an account to enhance search across your private repositories: search multiple
                                repos & commit history, monitor, save searches, and more.
                            </p>
                            <button
                                type="submit"
                                onClick={onSignUpClick}
                                className={classNames(
                                    'btn btn-sm btn-primary btn-link w-100 border-0 font-weight-normal my-3',
                                    styles.button
                                )}
                            >
                                <span className="my-0">Create an account</span>
                            </button>
                            <button
                                type="button"
                                className={classNames('my-3 btn btn-link', styles.textLink)}
                                onClick={() => setHasAccount(true)}
                            >
                                Create an account
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
                                    'btn btn-sm btn-primary btn-link w-100 border-0 font-weight-normal my-3',
                                    styles.button
                                )}
                            >
                                <span className="my-0">Enter Access Token</span>
                            </button>
                            <p className={classNames('mb-3', styles.textLink)}>
                                <a href={signUpUrl} onClick={onSignUpClick}>
                                    Create an account
                                </a>
                            </p>
                        </Form>
                    )}
                </>
            )}

            {!validating && (
                <>
                    {!validAccessToken && hasAccessToken && (
                        <Form onSubmit={onSubmitAccessToken}>
                            <a
                                href={signInUrl}
                                className="btn btn-lg btn-block btn-danger border-0 font-weight-normal"
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
                    )}
                </>
            )}
        </div>
    )
}
