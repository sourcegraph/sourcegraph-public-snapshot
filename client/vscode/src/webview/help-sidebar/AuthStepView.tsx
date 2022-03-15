import classNames from 'classnames'
import React, { useMemo, useState } from 'react'
import { Form } from 'reactstrap'

import { LoaderInput } from '@sourcegraph/branded/src/components/LoaderInput'
import { currentAuthStateQuery } from '@sourcegraph/shared/src/auth'
import { CurrentAuthStateResult, CurrentAuthStateVariables } from '@sourcegraph/shared/src/graphql-operations'
import { Alert } from '@sourcegraph/wildcard'

import { WebviewPageProps } from '../platform/context'
import styles from '../search-sidebar/AuthSidebarView.module.scss'

const SIDEBAR_UTM_PARAMS = 'utm_medium=VSCIDE&utm_source=sidebar&utm_campaign=vsce-sign-up&utm_content=sign-up'

interface AuthSidebarViewProps extends Pick<WebviewPageProps, 'extensionCoreAPI' | 'platformContext' | 'instanceURL'> {}

/**
 * Rendered by sidebar in search-home state when user doesn't have a valid access token.
 */
export const AuthStepView: React.FunctionComponent<AuthSidebarViewProps> = ({
    instanceURL,
    extensionCoreAPI,
    platformContext,
}) => {
    const [state, setState] = useState<'initial' | 'validating' | 'success' | 'failure'>('initial')
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
            if (usePrivateInstance) {
                const newInstanceUrl = (event.currentTarget.elements.namedItem('instance-url') as HTMLInputElement)
                    .value
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
                        await extensionCoreAPI.setEndpointUri(hostname)
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
        platformContext.telemetryService.log('VSCESidebarCreateAccount')
        await extensionCoreAPI.openLink(signUpURL)
    }

    return (
        <Form onSubmit={validateAccessToken}>
            <p className={classNames(styles.ctaParagraph)}>
                Sign in by entering an access token created through your user settings on {hostname}.
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
                <Alert variant="danger" className={classNames(styles.ctaParagraph, 'my-1')}>
                    Unable to verify your access token for {hostname}. Please try again with a new access token.
                </Alert>
            )}
            <p className={classNames(styles.ctaButtonWrapperWithContextBelow)}>
                <LoaderInput loading={state === 'validating'}>
                    <label htmlFor="access-token-input">Access Token</label>
                    <input
                        className={classNames('form-control', styles.ctaInput)}
                        id="access-token-input"
                        type="text"
                        name="token"
                        required={true}
                        minLength={40}
                        autoFocus={true}
                        spellCheck={false}
                        disabled={state === 'validating'}
                        formNoValidate={true}
                        placeholder="ex 6dfc880b320dff712d9f6cfcac5cbd13ebfad1d8"
                    />
                </LoaderInput>
            </p>
            {usePrivateInstance && (
                <p className={classNames(styles.ctaButtonWrapperWithContextBelow)}>
                    <LoaderInput loading={state === 'validating'}>
                        <label htmlFor="private-url-input">Sourcegraph Instance URL</label>
                        <input
                            className={classNames('form-control', styles.ctaInput)}
                            id="private-url-input"
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
            <div>
                <button
                    type="button"
                    className={classNames(styles.ctaParagraph, 'btn btn-text-link text-left my-0')}
                    onClick={onSignUpClick}
                >
                    Create an account
                </button>
            </div>
        </Form>
    )
}
