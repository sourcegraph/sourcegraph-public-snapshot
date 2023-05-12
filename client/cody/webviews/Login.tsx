import { useCallback, useState } from 'react'

import { TextFieldType } from '@vscode/webview-ui-toolkit/dist/text-field'
import { VSCodeTextField, VSCodeButton } from '@vscode/webview-ui-toolkit/react'

import { renderCodyMarkdown } from '@sourcegraph/cody-shared/src/chat/markdown'
import { CODY_TERMS_MARKDOWN } from '@sourcegraph/cody-ui/src/terms'

import { AuthStatus } from '../src/chat/protocol'

import styles from './Login.module.css'

interface LoginProps {
    authStatus?: AuthStatus
    onLogin: (token: string, endpoint: string) => void
    serverEndpoint?: string
}

export const Login: React.FunctionComponent<React.PropsWithChildren<LoginProps>> = ({
    authStatus,
    onLogin,
    serverEndpoint,
}) => {
    const [token, setToken] = useState<string>('')
    const [endpoint, setEndpoint] = useState(serverEndpoint)

    const onSubmit = useCallback<React.FormEventHandler>(
        event => {
            event.preventDefault()
            if (endpoint && token) {
                onLogin(token, endpoint)
            }
        },
        [endpoint, onLogin, token]
    )

    return (
        <div className={styles.container}>
            {authStatus && (
                <p className={styles.error}>
                    {!authStatus.loggedIn
                        ? 'Invalid credentials. Please check the Sourcegraph instance URL and access token.'
                        : !authStatus?.hasVerifiedEmail
                        ? 'Email not verified. Please add a verified email to your Sourcegraph instance account.'
                        : null}
                </p>
            )}
            <section className={styles.section}>
                <h2 className={styles.sectionHeader}>Enterprise User</h2>
                <form className={styles.wrapper} onSubmit={onSubmit}>
                    <VSCodeTextField
                        id="endpoint"
                        value={endpoint}
                        className={styles.input}
                        placeholder="https://sourcegraph.com"
                        onInput={e => setEndpoint((e.target as HTMLInputElement).value)}
                    >
                        Sourcegraph Instance URL
                    </VSCodeTextField>
                    <VSCodeTextField
                        id="accessToken"
                        value={token}
                        className={styles.input}
                        type={TextFieldType.password}
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
            </section>
            <div
                className={styles.terms}
                dangerouslySetInnerHTML={{ __html: renderCodyMarkdown(CODY_TERMS_MARKDOWN) }}
            />
        </div>
    )
}
