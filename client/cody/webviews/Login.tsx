import { useCallback, useState } from 'react'

import { TextFieldType } from '@vscode/webview-ui-toolkit/dist/text-field'
import { VSCodeTextField, VSCodeButton } from '@vscode/webview-ui-toolkit/react'

import { renderCodyMarkdown } from '@sourcegraph/cody-shared/src/chat/markdown'
import { CODY_TERMS_MARKDOWN } from '@sourcegraph/cody-ui/src/terms'

import { AuthStatus } from '../src/chat/protocol'

import { ConnectApp } from './ConnectApp'
import { VSCodeWrapper } from './utils/VSCodeApi'

import styles from './Login.module.css'

interface LoginProps {
    authStatus?: AuthStatus
    onLogin: (token: string, endpoint: string) => void
    serverEndpoint?: string
    isAppInstalled: boolean
    vscodeAPI: VSCodeWrapper
}

export const Login: React.FunctionComponent<React.PropsWithChildren<LoginProps>> = ({
    authStatus,
    onLogin,
    serverEndpoint,
    isAppInstalled,
    vscodeAPI,
}) => {
    const [token, setToken] = useState<string>('')
    const [endpoint, setEndpoint] = useState(serverEndpoint)

    const onSubmit = useCallback<React.FormEventHandler>(
        event => {
            event.preventDefault()
            if (endpoint) {
                onLogin(token, endpoint)
            }
        },
        [endpoint, onLogin, token]
    )

    return (
        <div className={styles.container}>
            {authStatus && <ErrorContainer authStatus={authStatus} />}
            <section className={styles.section}>
                <h2 className={styles.sectionHeader}>Enterprise User</h2>
                <form className={styles.wrapper} onSubmit={onSubmit}>
                    <VSCodeTextField
                        id="endpoint"
                        value={endpoint || ''}
                        className={styles.input}
                        placeholder="https://example.sourcegraph.com"
                        onChange={e => setEndpoint((e.target as HTMLInputElement).value)}
                        onInput={e => setEndpoint((e.target as HTMLInputElement).value)}
                    >
                        Sourcegraph Instance URL
                    </VSCodeTextField>
                    <VSCodeTextField
                        id="accessToken"
                        value={token}
                        className={styles.input}
                        type={TextFieldType.password}
                        onChange={e => setToken((e.target as HTMLInputElement).value)}
                        onInput={e => setToken((e.target as HTMLInputElement).value)}
                    >
                        Access Token (
                        <a href="https://docs.sourcegraph.com/cli/how-tos/creating_an_access_token">docs</a>)
                    </VSCodeTextField>
                    <VSCodeButton className={styles.button} type="submit">
                        Sign In
                    </VSCodeButton>
                </form>
            </section>
            <div className={styles.divider} />
            <section className={styles.section}>
                <h2 className={styles.sectionHeader}>Everyone Else</h2>
                <p className={styles.openMessage}>
                    Cody for open source code is available to all users with a Sourcegraph.com account
                </p>
                <a href="https://sourcegraph.com/user/settings/tokens/new/callback?requestFrom=CODY">
                    <VSCodeButton
                        className={styles.button}
                        type="button"
                        onClick={() => setEndpoint('https://sourcegraph.com')}
                    >
                        Continue with Sourcegraph.com
                    </VSCodeButton>
                </a>
                {isAppInstalled && <ConnectApp vscodeAPI={vscodeAPI} />}
            </section>
            <div
                className={styles.terms}
                dangerouslySetInnerHTML={{ __html: renderCodyMarkdown(CODY_TERMS_MARKDOWN) }}
            />
        </div>
    )
}

const ERROR_MESSAGES = {
    VERSION:
        'You are trying to connect to a Sourcegraph instance that is not compatible with Cody. Please reach out to your site admin to upgrade your Sourcegraph instance to a compatible version: 5.1.0 or above.',
    INVALID: 'Invalid credentials. Please check the Sourcegraph instance URL and access token.',
    EMAIL: 'Email not verified. Please add a verified email to your Sourcegraph.com account.',
}

const ErrorContainer: React.FunctionComponent<{ authStatus: AuthStatus }> = ({ authStatus }) => {
    const {
        authenticated,
        onSupportedSiteVersion,
        showInvalidAccessTokenError,
        requiresVerifiedEmail,
        hasVerifiedEmail,
    } = authStatus

    if (showInvalidAccessTokenError) {
        return <p className={styles.error}>{ERROR_MESSAGES.INVALID}</p>
    }
    if (authenticated && !onSupportedSiteVersion) {
        return <p className={styles.error}>{ERROR_MESSAGES.VERSION}</p>
    }
    if (authenticated && requiresVerifiedEmail && !hasVerifiedEmail) {
        return <p className={styles.error}>{ERROR_MESSAGES.EMAIL}</p>
    }
    return null
}
