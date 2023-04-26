import { useCallback, useState } from 'react'

import { TextFieldType } from '@vscode/webview-ui-toolkit/dist/text-field'
import { VSCodeTextField, VSCodeButton } from '@vscode/webview-ui-toolkit/react'

import { renderCodyMarkdown } from '@sourcegraph/cody-shared/src/chat/markdown'
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
            {isValidLogin === false && (
                <p className={styles.error}>
                    Invalid credentials. Please check the Sourcegraph instance URL and access token.
                </p>
            )}
            <section className={styles.section}>
                <h2 className={styles.sectionHeader}>Enterprise User</h2>
                <form className={styles.wrapper} onSubmit={onSubmit}>
                    <label htmlFor="endpoint" className={styles.label}>
                        Sourcegraph Instance URL
                    </label>
                    <VSCodeTextField
                        id="endpoint"
                        value={endpoint || ''}
                        className={styles.input}
                        placeholder="https://example.sourcegraph.com"
                        onInput={(e: any) => setEndpoint(e.target.value)}
                    />

                    <label htmlFor="accessToken" className={styles.label}>
                        Access Token (
                        <a href="https://docs.sourcegraph.com/cli/how-tos/creating_an_access_token">docs</a>)
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
            </section>
            <section className={styles.section}>
                <h2 className={styles.sectionHeader}>Everyone Else</h2>
                <div className={styles.wrapper}>
                    <p className={styles.linkToForm}>
                        <a href="https://discord.gg/sourcegraph-969688426372825169">
                            Join our Discord to request access.
                        </a>
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
                </div>
            </section>
            <div
                className={styles.terms}
                dangerouslySetInnerHTML={{ __html: renderCodyMarkdown(CODY_TERMS_MARKDOWN) }}
            />
        </div>
    )
}
