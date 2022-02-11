import React, { useMemo, useState } from 'react'
import { Form } from 'reactstrap'

import { LoaderInput } from '@sourcegraph/branded/src/components/LoaderInput'
import { currentAuthStateQuery } from '@sourcegraph/shared/src/auth'
import { CurrentAuthStateResult, CurrentAuthStateVariables } from '@sourcegraph/shared/src/graphql-operations'
import { Alert, Button, ButtonProps } from '@sourcegraph/wildcard'

import { WebviewPageProps } from '../platform/context'

import styles from './AuthSidebarView.module.scss'

/**
 * Rendered by sidebar in search-home state when user doesn't have a valid access token.
 */
export const AuthSidebarView: React.FunctionComponent<WebviewPageProps> = ({
    instanceURL,
    extensionCoreAPI,
    platformContext,
}) => {
    const [state, setState] = useState<'initial' | 'validating' | 'success' | 'failure'>('initial')

    const [hasAccount, setHasAccount] = useState(false)

    const signUpURL = useMemo(() => new URL('sign-up?editor=vscode', instanceURL).href, [instanceURL])
    const instanceHostname = useMemo(() => new URL(instanceURL).hostname, [instanceURL])

    const ctaButtonProps: Partial<ButtonProps> = {
        variant: 'primary',
        className: 'font-weight-normal w-100 my-1 border-0',
    }
    const buttonLinkProps: Partial<ButtonProps> = {
        variant: 'link',
        size: 'sm',
        display: 'block',
        className: 'pl-0',
    }

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
        <div>
            <h5 className="mt-3 mb-2">Welcome!</h5>
            <p>
                The Sourcegraph extension allows you to search millions of open source repositories without cloning them
                to your local machine.
            </p>
            <p>
                Developers at some of the world’s best software companies use Sourcegraph to onboard to new code bases,
                find examples, research errors, and resolve incidents.
            </p>
            <div>
                <p className="mb-0">Learn more:</p>
                <a href="http://sourcegraph.com/" className="my-0" onClick={() => onLinkClick('Sourcegraph')}>
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

            <Form onSubmit={validateAccessToken} className={styles.formContainer}>
                <h5 className="mb-2">Search your private code</h5>
                {content}
            </Form>
        </div>
    )

    if (!hasAccount) {
        return renderCommon(
            <>
                <p>
                    Create an account to enhance search across your private repositories: search multiple repos & commit
                    history, monitor, save searches and more.
                </p>
                <Button onClick={onSignUpClick} {...ctaButtonProps}>
                    Create an account
                </Button>
                <Button onClick={() => setHasAccount(true)} {...buttonLinkProps}>
                    Have an account?
                </Button>
            </>
        )
    }

    return renderCommon(
        <>
            <p>Sign in by entering an access token created through your user settings on {instanceHostname}.</p>
            <p>
                See our <a href="https://docs.sourcegraph.com/cli/how-tos/creating_an_access_token">user docs</a> for a
                video guide on how to create an access token.
            </p>
            {state === 'failure' && (
                <Alert variant="danger">
                    Unable to verify your access token for {instanceHostname}. Please try again with a new access token.
                </Alert>
            )}
            <LoaderInput loading={state === 'validating'}>
                <input
                    className="input form-control mb-1"
                    type="text"
                    name="token"
                    required={true}
                    autoFocus={true}
                    spellCheck={false}
                    disabled={state === 'validating'}
                    placeholder="ex 6dfc880b320dff712d9f6cfcac5cbd13ebfad1d8"
                />
            </LoaderInput>
            <Button type="submit" disabled={state === 'validating'} {...ctaButtonProps}>
                Enter access token
            </Button>
            <Button onClick={onSignUpClick} {...buttonLinkProps}>
                Create an account
            </Button>
        </>
    )
}
