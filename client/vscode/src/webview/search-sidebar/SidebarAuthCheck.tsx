import classNames from 'classnames'
import React, { useCallback, useState } from 'react'

import { WebviewPageProps } from '../platform/context'

import styles from './OpenSearchPanelCta.module.scss'

interface OpenSearchPanelCtaProps extends Pick<WebviewPageProps, 'sourcegraphVSCodeExtensionAPI'> {
    className?: string
}

export const SidebarAuthCheck: React.FunctionComponent<OpenSearchPanelCtaProps> = ({
    sourcegraphVSCodeExtensionAPI,
    className,
}) => {

    const accessToken = await sourcegraphVSCodeExtensionAPI.hasAccessToken

    return (
    <div className={classNames('d-flex flex-column align-items-left justify-content-center', className)}>
        <p className={classNames('mt-3', styles.title)}>Search Your Private Code</p>
        {accessToken ? (
            <div>
                <p className={classNames('my-3', styles.text)}>
                    Create an account to enhance search across your private repositories: search multiple repos & commit
                    history, monitor, save searches, and more.
                </p>
                <a
                    href="https://sourcegraph.com/sign-up"
                    className={classNames('btn btn-sm w-100 border-0 font-weight-normal', styles.button)}
                >
                    <span className={classNames('my-3', styles.text)}>Create an account</span>
                </a>
                <p className={classNames('my-3', styles.text)}>
                    <a href="https://sourcegraph.com/sign-in">Have an account?</a>
                </p>
            </div>
        ) : (
            <div>
                <p className={classNames('my-3', styles.text)}>
                    Sign in by entering an access token created through your user setting on sourcegraph.com.
                </p>
                <p className={classNames('my-3', styles.text)}>
                    See our <a href="https://docs.sourcegraph.com/cli/how-tos/creating_an_access_token">user docs</a>{' '}
                    for a video guide on how to create an access token.
                </p>
                <input
                    className={classNames('my-3 w-100 p-1', styles.text)}
                    type="text"
                    placeholder="ex 6dfc880b320dff712d9f6cfcac5cbd13ebfad1d8"
                />
                <a
                    href="https://sourcegraph.com/sign-up"
                    className={classNames('btn btn-sm w-100 border-0 font-weight-normal', styles.button)}
                >
                    <span className={classNames('my-3', styles.text)}>Enter Access Token</span>
                </a>
                <p className={classNames('my-3', styles.text)}>
                    <a href="https://sourcegraph.com/sign-in">Create an account</a>
                </p>
            </div>
        )}
    </div>
)
