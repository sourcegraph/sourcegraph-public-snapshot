import React, { useMemo, useState } from 'react'

import classNames from 'classnames'

import { Form } from '@sourcegraph/branded/src/components/Form'
import { LoaderInput } from '@sourcegraph/branded/src/components/LoaderInput'
import { currentAuthStateQuery } from '@sourcegraph/shared/src/auth'
import { CurrentAuthStateResult, CurrentAuthStateVariables } from '@sourcegraph/shared/src/graphql-operations'
import { Alert, Typography } from '@sourcegraph/wildcard'

import { WebviewPageProps } from '../../platform/context'

import styles from './AuthSidebarView.module.scss'

const SIDEBAR_UTM_PARAMS = 'utm_medium=VSCODE&utm_source=sidebar&utm_campaign=vsce-sign-up&utm_content=sign-up'

interface AuthSidebarViewProps
    extends Pick<WebviewPageProps, 'extensionCoreAPI' | 'platformContext' | 'instanceURL' | 'authenticatedUser'> {}

interface AuthSidebarCtaProps extends Pick<WebviewPageProps, 'platformContext'> {}

/**
 * Rendered by sidebar in search-home state when user doesn't have a valid access token.
 */
export const AuthSidebarView: React.FunctionComponent<React.PropsWithChildren<AuthSidebarViewProps>> = ({
    instanceURL,
    extensionCoreAPI,
    platformContext,
    authenticatedUser,
}) => {
    const [state, setState] = useState<'initial' | 'validating' | 'success' | 'failure'>('initial')
    const [hasAccount, setHasAccount] = useState(!authenticatedUser)
    const [usePrivateInstance, setUsePrivateInstance] = useState(false)
    const signUpURL = `https://sourcegraph.com/sign-up?editor=vscode&${SIDEBAR_UTM_PARAMS}`
    const instanceHostname = useMemo(() => new URL(instanceURL).hostname, [instanceURL])
    const [hostname, setHostname] = useState(instanceHostname)

    const validateAccessToken: React.FormEventHandler<HTMLFormElement> = (event): void => {
        event.preventDefault()
        if (state !== 'validating') {
            const newAccessToken = (event.currentTarget.elements.namedItem('token') as HTMLInputElement).value
            let authStateVariables = {
                request: currentAuthStateQuery,
                variables: {},
                mightContainPrivateInfo: true,
                overrideAccessToken: newAccessToken,
            }
            let newInstanceUrl: string
            if (usePrivateInstance) {
                newInstanceUrl = (event.currentTarget.elements.namedItem('instance-url') as HTMLInputElement).value
                setHostname(newInstanceUrl)
                authStateVariables = { ...authStateVariables, ...{ overrideSourcegraphURL: newInstanceUrl } }
            }
            setState('validating')
            const currentAuthStateResult = platformContext
                .requestGraphQL<CurrentAuthStateResult, CurrentAuthStateVariables>(authStateVariables)
                .toPromise()

            currentAuthStateResult
                .then(async ({ data }) => {
                    if (data?.currentUser) {
                        setState('success')
                        if (newInstanceUrl) {
                            await extensionCoreAPI.setEndpointUri(newInstanceUrl)
                        }
                        return extensionCoreAPI.setAccessToken(newAccessToken)
                    }
                    setState('failure')
                    return
                })
                // v2/debt: Disambiguate network vs auth errors like we do in the browser extension.
                .catch(() => setState('failure'))
        }
        // If successful, update setting. This form will no longer be rendered
    }

    const onSignUpClick = async (): Promise<void> => {
        setHasAccount(true)
        platformContext.telemetryService.log('VSCESidebarCreateAccount')
        await extensionCoreAPI.openLink(signUpURL)
    }

    if (state === 'success') {
        // This form should no longer be rendered as the extension context
        // will be invalidated. We should show a notification that the accessToken
        // has successfully been updated.
        return null
    }

    const renderCommon = (content: JSX.Element): JSX.Element => (
        <>
            <div className={classNames(styles.ctaContainer)}>
                <Form onSubmit={validateAccessToken}>
                    <button type="button" className={classNames('btn btn-outline-secondary', styles.ctaTitle)}>
                        <Typography.H5 className="flex-grow-1">Search your private code</Typography.H5>
                    </button>
                    {content}
                </Form>
            </div>
        </>
    )

    if (!hasAccount) {
        return renderCommon(
            <>
                <p className={classNames(styles.ctaParagraph)}>
                    Create an account to search across your private repositories and access advanced features: search
                    multiple repositories & commit history, monitor code changes, save searches, and more.
                </p>
                <p className={classNames(styles.ctaButtonWrapperWithContextBelow)}>
                    <button type="button" onClick={onSignUpClick} className={classNames('btn my-1', styles.ctaButton)}>
                        Create an account
                    </button>
                </p>
                <button
                    type="button"
                    className={classNames(styles.ctaParagraph, 'btn btn-text-link text-left')}
                    onClick={() => setHasAccount(true)}
                >
                    Have an account?
                </button>
            </>
        )
    }

    return renderCommon(
        <>
            <p className={classNames(styles.ctaParagraph)}>
                Sign in by entering an access token created through your user settings on {hostname}.
            </p>
            <p className={classNames(styles.ctaParagraph)}>
                See our {/* eslint-disable-next-line react/forbid-elements */}{' '}
                <a
                    href={`https://docs.sourcegraph.com/cli/how-tos/creating_an_access_token?${SIDEBAR_UTM_PARAMS}`}
                    onClick={() => platformContext.telemetryService.log('VSCESidebarCreateToken')}
                >
                    user docs
                </a>{' '}
                for a video guide on how to create an access token.
            </p>
            <p className={classNames(styles.ctaButtonWrapperWithContextBelow)}>
                <LoaderInput loading={state === 'validating'}>
                    <label htmlFor="access-token-input">Access Token</label>
                    <input
                        className={classNames('input form-control', styles.ctaInput)}
                        id="access-token-input"
                        type="text"
                        name="token"
                        required={true}
                        autoFocus={true}
                        spellCheck={false}
                        disabled={state === 'validating'}
                        placeholder="ex 6dfc880b320dff712d9f6cfcac5cbd13ebfad1d8"
                    />
                </LoaderInput>
            </p>
            {usePrivateInstance && (
                <p className={classNames(styles.ctaButtonWrapperWithContextBelow)}>
                    <LoaderInput loading={state === 'validating'}>
                        <label htmlFor="instance-url-input">Sourcegraph Instance URL</label>
                        <input
                            className={classNames('input form-control', styles.ctaInput)}
                            id="instance-url-input"
                            type="url"
                            name="instance-url"
                            required={true}
                            autoFocus={true}
                            spellCheck={false}
                            disabled={state === 'validating'}
                            placeholder="ex https://sourcegraph.example.com"
                        />
                    </LoaderInput>
                </p>
            )}
            <button
                type="submit"
                disabled={state === 'validating'}
                className={classNames('btn my-1', styles.ctaButton, styles.ctaButtonWrapperWithContextBelow)}
            >
                Authenticate account
            </button>
            {state === 'failure' && (
                <Alert variant="danger" className={classNames(styles.ctaParagraph, 'my-1')}>
                    Unable to verify your access token for {hostname}. Please try again with a new access token.
                </Alert>
            )}
            {!usePrivateInstance ? (
                <button
                    type="button"
                    className={classNames(styles.ctaParagraph, 'btn btn-text-link text-left my-0')}
                    onClick={() => setUsePrivateInstance(true)}
                >
                    Need to connect to a private instance?
                </button>
            ) : (
                <button
                    type="button"
                    className={classNames(styles.ctaParagraph, 'btn btn-text-link text-left my-0')}
                    onClick={() => setUsePrivateInstance(false)}
                >
                    Not a private instance user?
                </button>
            )}
            <div className="my-0">
                <button
                    type="button"
                    className={classNames(styles.ctaParagraph, 'btn btn-text-link text-left my-0')}
                    onClick={onSignUpClick}
                >
                    Create an account
                </button>
            </div>
        </>
    )
}

export const AuthSidebarCta: React.FunctionComponent<React.PropsWithChildren<AuthSidebarCtaProps>> = ({
    platformContext,
}) => {
    const onLinkClick = (type: 'Sourcegraph' | 'Extension'): void =>
        platformContext.telemetryService.log(`VSCESidebarLearn${type}Click`)

    return (
        <div>
            <button type="button" className={classNames('btn btn-outline-secondary', styles.ctaTitle)}>
                <Typography.H5 className="flex-grow-1">Welcome</Typography.H5>
            </button>
            <p className={classNames(styles.ctaParagraph)}>
                The Sourcegraph extension allows you to search millions of open source repositories without cloning them
                to your local machine.
            </p>
            <p className={classNames(styles.ctaParagraph)}>
                Developers use Sourcegraph every day to onboard to new code bases, find code to reuse, resolve
                incidents, fix security vulnerabilities, and more.
            </p>
            <div className={classNames(styles.ctaParagraph)}>
                <p className="mb-0">Learn more:</p>
                {/* eslint-disable-next-line react/forbid-elements */}
                <a
                    href={'https://sourcegraph.com/?' + SIDEBAR_UTM_PARAMS}
                    className="my-0"
                    onClick={() => onLinkClick('Sourcegraph')}
                >
                    Sourcegraph.com
                </a>
                <br />
                {/* eslint-disable-next-line react/forbid-elements */}
                <a
                    href="https://marketplace.visualstudio.com/items?itemName=sourcegraph.sourcegraph"
                    className="my-0"
                    onClick={() => onLinkClick('Extension')}
                >
                    Sourcegraph VS Code extension
                </a>
            </div>
        </div>
    )
}
