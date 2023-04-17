import { useCallback, useState } from 'react'

import { TextFieldType } from '@vscode/webview-ui-toolkit/dist/text-field'
import { VSCodeTextField, VSCodeButton } from '@vscode/webview-ui-toolkit/react'

import { renderMarkdown } from '@sourcegraph/cody-shared/src/chat/markdown'
import { CODY_TERMS_MARKDOWN } from '@sourcegraph/cody-ui/src/terms'

import styles from './Login.module.css'

interface LoginProps {
    isValidLogin?: boolean
    onLogin: (token: string, endpoint: string) => void
}

export const Login: React.FunctionComponent<React.PropsWithChildren<LoginProps>> = ({ isValidLogin, onLogin }) => {
    const [token, setToken] = useState<string>('')
    const [endpoint, setEndpoint] = useState('https://sourcegraph.com')

    const onSubmit = useCallback<React.FormEventHandler>(
        event => {
            event.preventDefault()
            onLogin(token, endpoint)
        },
        [endpoint, onLogin, token]
    )

    return (
        <form className={styles.container} onSubmit={onSubmit}>
            <label htmlFor="endpoint" className={styles.label}>
                Sourcegraph URL
            </label>
            <VSCodeTextField
                id="endpoint"
                value={endpoint}
                className={styles.input}
                onInput={(e: any) => setEndpoint(e.target.value)}
            />

            <label htmlFor="accessToken" className={styles.label}>
                Access Token
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
            <div className={styles.terms} dangerouslySetInnerHTML={{ __html: renderMarkdown(CODY_TERMS_MARKDOWN) }} />

            {isValidLogin === false && (
                <p className={styles.error}>
                    Invalid credentials. Please check the Sourcegraph instance URL and access token.
                </p>
            )}
        </form>
    )
}
