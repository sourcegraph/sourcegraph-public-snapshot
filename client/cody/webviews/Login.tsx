import { useCallback, useState } from 'react'

import { TextFieldType } from '@vscode/webview-ui-toolkit/dist/text-field'
import { VSCodeTextField, VSCodeButton } from '@vscode/webview-ui-toolkit/react'

import { renderMarkdown } from '@sourcegraph/cody-shared/src/chat/markdown'
import { CODY_TERMS_MARKDOWN } from '@sourcegraph/cody-ui/src/terms'

import styles from './Login.module.css'

interface LoginProps {
    isValidLogin?: boolean
    onLogin: (token: string, endpoint: string) => void
    serverEndpoint?: string
}

export const Login: React.FunctionComponent<React.PropsWithChildren<LoginProps>> = ({
    isValidLogin,
    onLogin,
    serverEndpoint,
}) => {
    const [token, setToken] = useState<string>('')
    const [endpoint, setEndpoint] = useState(serverEndpoint || 'https://example.sourcegraph.com')

    const onSubmit = useCallback<React.FormEventHandler>(
        event => {
            event.preventDefault()
            onLogin(token, endpoint)
        },
        [endpoint, onLogin, token]
    )

    return (
        <div className={styles.container}>
            {isValidLogin === false && (
                <p className={styles.error}>
                    Invalid credentials. Please check the Sourcegraph instance URL and access token.
                </p>
            )}
            <p className={styles.inputLabel}>
                <i className="codicon codicon-organization" />
                <span>Enterprise User</span>
            </p>
            <form className={styles.wrapper} onSubmit={onSubmit}>
                <label htmlFor="endpoint" className={styles.inputLabel}>
                    <i className="codicon codicon-link" />
                    <span>Sourcegraph Instance URL</span>
                </label>
                <VSCodeTextField
                    id="endpoint"
                    value={endpoint}
                    className={styles.input}
                    onInput={(e: any) => setEndpoint(e.target.value)}
                />

                <label htmlFor="accessToken" className={styles.inputLabel}>
                    <i className="codicon codicon-key" />
                    <span>
                        Personal Access Token (
                        <a href="https://docs.sourcegraph.com/cli/how-tos/creating_an_access_token">docs</a>)
                    </span>
                </label>
                <VSCodeTextField
                    id="accessToken"
                    value={token}
                    placeholder=""
                    className={styles.input}
                    type={TextFieldType.password}
                    onInput={(e: any) => setToken(e.target.value)}
                />

                <VSCodeButton className={styles.button} type="submit">
                    Sign In
                </VSCodeButton>
            </form>
            <p className={styles.inputLabel}>
                <i className="codicon codicon-account" />
                <span>Community User</span>
            </p>
            <div className={styles.wrapper}>
                <span>Connect Cody to sourcegraph.com in browser.</span>
                <p className={styles.input}>
                    <a href="https://docs.google.com/forms/d/e/1FAIpQLScSI06yGMls-V1FALvFyURi8U9bKRTSKPworBhzZEHDQvo0HQ/viewform">
                        Fill out this form to request access.
                    </a>
                </p>
                <a href="https://sourcegraph.com/user/settings/tokens/new/callback?requestFrom=CODY">
                    <VSCodeButton
                        className={styles.button}
                        type="button"
                        onClick={() => setEndpoint('https://sourcegraph.com')}
                    >
                        Continue with sourcegraph.com
                    </VSCodeButton>
                </a>
            </div>
            <div className={styles.terms} dangerouslySetInnerHTML={{ __html: renderMarkdown(CODY_TERMS_MARKDOWN) }} />
        </div>
    )
}
