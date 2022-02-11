import classNames from 'classnames'
import React, { useMemo, useState } from 'react'

import { Form } from 'reactstrap'

import { LoaderInput } from '@sourcegraph/branded/src/components/LoaderInput'
import { currentAuthStateQuery } from '@sourcegraph/shared/src/auth'
import { CurrentAuthStateResult, CurrentAuthStateVariables } from '@sourcegraph/shared/src/graphql-operations'
import { Alert } from '@sourcegraph/wildcard'

import { WebviewPageProps } from '../platform/context'

import styles from './AuthSidebarView.module.scss'

const SIDEBAR_UTM_PARAMS = 'utm_medium=VSCIDE&utm_source=sidebar&utm_campaign=vsce-sign-up&utm_content=sign-up'

interface AuthSidebarViewProps extends WebviewPageProps {
    stateStatus: string
}

/**
 * Rendered by sidebar in search-home state when user doesn't have a valid access token.
 */
export const AuthSidebarView: React.FunctionComponent<AuthSidebarViewProps> = ({
    instanceURL,
    extensionCoreAPI,
    platformContext,
    stateStatus,
}) => {
    const [state, setState] = useState<'initial' | 'validating' | 'success' | 'failure'>('initial')

    const [hasAccount, setHasAccount] = useState(false)

    const signUpURL = useMemo(() => new URL('sign-up?editor=vscode&' + SIDEBAR_UTM_PARAMS, instanceURL).href, [
        instanceURL,
    ])
    const instanceHostname = useMemo(() => new URL(instanceURL).hostname, [instanceURL])

    const validateAccessToken: React.FormEventHandler<HTMLFormElement> = (event): void => {
        event.preventDefault()
        if (state !== 'validating') {
            const newAccessToken = (event.currentTarget.elements.namedItem('token') as HTMLInputElement).value

            setState('validating')
            const currentAuthStateResult = platformContext
                .requestGraphQL<CurrentAuthStateResult, CurrentAuthStateVariables>({
                    request: currentAuthStateQuery,
                    variables: {},
                    mightContainPrivateInfo: true,
                    overrideAccessToken: newAccessToken,
                })
                .toPromise()

            currentAuthStateResult
                .then(({ data }) => {
                    if (data?.currentUser) {
                        setState('success')
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

    const onSignUpClick = (): void => {
        setHasAccount(true)
        extensionCoreAPI
            .openLink(signUpURL)
            .then(() => {})
            .catch(() => {})

        platformContext.telemetryService.log('VSCESidebarCreateAccount')
    }

    const onLinkClick = (type: 'Sourcegraph' | 'Extension'): void =>
        platformContext.telemetryService.log(`VSCESidebarLearn${type}Click`)

    if (state === 'success') {
        // This form should no longer be rendered as the extension context
        // will be invalidated. We should show a notification that the accessToken
        // has successfully been updated.
        return null
    }

    const renderCommon = (content: JSX.Element): JSX.Element => (
        <div className={classNames(styles.ctaContainer)}>
            {stateStatus === 'search-home' && (
                <div>
                    <button type="button" className={classNames('btn btn-outline-secondary', styles.ctaTitle)}>
                        <h5 className="flex-grow-1">Welcome</h5>
                    </button>
                    <p className={classNames(styles.ctaParagraph)}>
                        The Sourcegraph extension allows you to search millions of open source repositories without
                        cloning them to your local machine.
                    </p>
                    <p className={classNames(styles.ctaParagraph)}>
                        Developers use Sourcegraph every day to onboard to new code bases, find code to reuse, resolve
                        incidents, fix security vulnerabilities, and more.
                    </p>
                    <div className={classNames(styles.ctaParagraph)}>
                        <p className="mb-0">Learn more:</p>
                        <a
                            href={'https://sourcegraph.com/?' + SIDEBAR_UTM_PARAMS}
                            className="my-0"
                            onClick={() => onLinkClick('Sourcegraph')}
                        >
                            Sourcegraph.com
                        </a>
                        <br />
                        <a
                            href="https://marketplace.visualstudio.com/items?itemName=sourcegraph.sourcegraph"
                            className="my-0"
                            onClick={() => onLinkClick('Extension')}
                        >
                            Sourcegraph VS Code extension
                        </a>
                    </div>
                </div>
            )}
            <Form onSubmit={validateAccessToken}>
                <button type="button" className={classNames('btn btn-outline-secondary', styles.ctaTitle)}>
                    <h5 className="flex-grow-1">Search your private code</h5>
                </button>
                {content}
            </Form>
        </div>
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
                <a onClick={() => setHasAccount(true)} className={classNames(styles.ctaParagraph)} href="/">
                    Have an account?
                </a>
            </>
        )
    }

    return renderCommon(
        <>
            <p className={classNames(styles.ctaParagraph)}>
                Sign in by entering an access token created through your user settings on {instanceHostname}.
            </p>
            <p className={classNames(styles.ctaParagraph)}>
                See our{' '}
                <a
                    href={`https://docs.sourcegraph.com/cli/how-tos/creating_an_access_token?${SIDEBAR_UTM_PARAMS}`}
                    onClick={() => platformContext.telemetryService.log('VSCESidebarCreateToken')}
                >
                    user docs
                </a>{' '}
                for a video guide on how to create an access token.
            </p>
            {state === 'failure' && (
                <Alert variant="danger">
                    Unable to verify your access token for {instanceHostname}. Please try again with a new access token.
                </Alert>
            )}
            <p className={classNames(styles.ctaButtonWrapperWithContextBelow)}>
                <LoaderInput loading={state === 'validating'}>
                    <input
                        className={classNames('input form-control', styles.ctaInput)}
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
            <p className={classNames(styles.ctaButtonWrapperWithContextBelow)}>
                <button
                    type="submit"
                    disabled={state === 'validating'}
                    className={classNames('btn my-1', styles.ctaButton)}
                >
                    Enter access token
                </button>
            </p>
            <a href="/" className={classNames(styles.ctaParagraph)} onClick={onSignUpClick}>
                Create an account
            </a>
        </>
    )
}
