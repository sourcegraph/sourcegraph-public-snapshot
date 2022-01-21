import classNames from 'classnames'
import React, { useCallback, useEffect, useState } from 'react'
import { Form } from 'reactstrap'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { WebviewPageProps } from '../platform/context'

import styles from './SearchSidebar.module.scss'

interface SidebarAuthCheckProps extends TelemetryProps, Pick<WebviewPageProps, 'sourcegraphVSCodeExtensionAPI'> {
    hasAccessToken: boolean
    onSubmitAccessToken: React.FormEventHandler<HTMLFormElement>
    validAccessToken: boolean
}

export const SidebarAuthCheck: React.FunctionComponent<SidebarAuthCheckProps> = ({
    sourcegraphVSCodeExtensionAPI,
    hasAccessToken,
    onSubmitAccessToken,
    validAccessToken,
    telemetryService,
}) => {
    // `undefined` while waiting for Comlink response.
    const [instanceHostname, setInstanceHostname] = useState<string | undefined>(undefined)
    const [signUpUrl, setSignUpUrl] = useState<string>('https://sourcegraph.com/sign-up?editor=vscode')
    const [signInUrl, setSignInUrl] = useState<string>('https://sourcegraph.com/sign-in?editor=vscode')
    const [hasAccount, setHasAccount] = useState(false)
    const [validating, setValidating] = useState(true)

    useEffect(() => {
        setValidating(true)
        if (instanceHostname === undefined) {
            sourcegraphVSCodeExtensionAPI
                .getInstanceHostname()
                .then(instance => {
                    setInstanceHostname(instance)
                    setSignUpUrl(new URL('sign-up?editor=vscode', instance).href)
                    setSignInUrl(new URL('sign-in?editor=vscode', instance).href)
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
        <div className={classNames('d-flex flex-column align-items-left justify-content-center')}>
            <p className={classNames(styles.ctaTitle)}>Search Your Private Code</p>
            {validating && <LoadingSpinner />}
            {!validating && !hasAccessToken && (
                <>
                    {!hasAccount ? (
                        <div className={classNames(styles.ctaContainer)}>
                            <p className={classNames(styles.ctaParagraph)}>
                                Create an account to enhance search across your private repositories: search multiple
                                repos & commit history, monitor, save searches, and more.
                            </p>
                            <button
                                type="submit"
                                onClick={onSignUpClick}
                                className={classNames('btn btn-sm font-weight-normal my-1', styles.ctaButton)}
                            >
                                Create an account
                            </button>
                            <p className={classNames(styles.ctaParagraph)}>
                                <a
                                    href="sourcegraph://signup"
                                    className={classNames('my-0', styles.text)}
                                    onClick={() => setHasAccount(true)}
                                >
                                    Have an account?
                                </a>
                            </p>
                        </div>
                    ) : (
                        // eslint-disable-next-line react/forbid-elements
                        <Form onSubmit={onSubmitAccessToken}>
                            <div className={classNames(styles.ctaContainer)}>
                                <p className={classNames(styles.ctaParagraph)}>
                                    Sign in by entering an access token created through your{' '}
                                    <a href={signInUrl} onClick={() => setHasAccount(true)}>
                                        user setting
                                    </a>{' '}
                                    on {instanceHostname}.
                                </p>
                                <p className={classNames(styles.ctaParagraph)}>
                                    See our{' '}
                                    <a href="https://docs.sourcegraph.com/cli/how-tos/creating_an_access_token">
                                        user docs
                                    </a>{' '}
                                    for a video guide on how to create an access token.
                                </p>
                                <p className={classNames(styles.ctaParagraph)}>
                                    <input
                                        className={classNames('input form-control', styles.ctaInput)}
                                        // className="input form-control my-1"
                                        type="text"
                                        name="token"
                                        required={true}
                                        autoFocus={true}
                                        placeholder="ex 6dfc880b320dff712d9f6cfcac5cbd13ebfad1d8"
                                    />
                                </p>
                                <button
                                    type="submit"
                                    className={classNames('btn btn-sm font-weight-normal my-1', styles.ctaButton)}
                                >
                                    Enter Access Token
                                </button>

                                <p className={classNames(styles.ctaParagraph)}>
                                    <a
                                        href={signUpUrl}
                                        className={classNames('my-0', styles.text)}
                                        onClick={onSignUpClick}
                                    >
                                        Create an account
                                    </a>
                                </p>
                            </div>
                        </Form>
                    )}
                </>
            )}

            {!validating && (
                <>
                    {!validAccessToken && hasAccessToken && (
                        <div className={classNames(styles.ctaContainer)}>
                            <Form onSubmit={onSubmitAccessToken}>
                                <p className={classNames(styles.ctaParagraph)}>
                                    See our{' '}
                                    <a href="https://docs.sourcegraph.com/cli/how-tos/creating_an_access_token">
                                        user docs
                                    </a>{' '}
                                    for a video guide on how to create an access token.
                                </p>
                                <p className={classNames(styles.ctaParagraph)}>
                                    <input
                                        className={classNames('input form-control', styles.ctaInput)}
                                        // className="input form-control my-1"
                                        type="text"
                                        name="token"
                                        required={true}
                                        autoFocus={true}
                                        placeholder="ex 6dfc880b320dff712d9f6cfcac5cbd13ebfad1d8"
                                    />
                                </p>
                                <button
                                    type="submit"
                                    className={classNames('btn btn-sm font-weight-normal my-1', styles.ctaButton)}
                                >
                                    Update Access Token
                                </button>
                                <a
                                    href={signInUrl}
                                    className={classNames('btn btn-lg btn-block', styles.ctaErrorContainer)}
                                >
                                    Unable to verify your Access Token for {instanceHostname}. Please try again with a
                                    new Access Token.
                                </a>
                            </Form>
                        </div>
                    )}
                </>
            )}
        </div>
    )
}
