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
    callbackScheme?: string
}

export const Login: React.FunctionComponent<React.PropsWithChildren<LoginProps>> = ({
    authStatus,
    onLogin,
    serverEndpoint,
    isAppInstalled,
    vscodeAPI,
    callbackScheme,
}) => {
    const [token, setToken] = useState<string>('')
    const [endpoint, setEndpoint] = useState(serverEndpoint)
    const authUri = new URL('https://sourcegraph.com/user/settings/tokens/new/callback')
    authUri.searchParams.append('requestFrom', callbackScheme === 'vscode-insiders' ? 'CODY_INSIDERS' : 'CODY')

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
                <a href={authUri.href}>
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
    DISABLED: 'Cody is not enabled on your instance. To enable Cody, please contact your site admin.',
    VERSION:
        'Cody is not supported by your Sourcegraph instance version (version: $SITEVERSION). To use Cody, please contact your site admin to upgrade to version 5.1.0 or above.',
    INVALID: 'Invalid credentials. Please check the Sourcegraph instance URL and access token.',
    EMAIL_NOT_VERIFIED: 'Email not verified. Please add a verified email to your Sourcegraph.com account.',
}

const ErrorContainer: React.FunctionComponent<{ authStatus: AuthStatus }> = ({ authStatus }) => {
    const {
        authenticated,
        siteHasCodyEnabled,
        showInvalidAccessTokenError,
        requiresVerifiedEmail,
        hasVerifiedEmail,
        siteVersion,
    } = authStatus
    // Version is compatible if this is an insider build or version is 5.1.0 or above
    // Right now we assumes all insider builds have Cody enabled
    // NOTE: Insider build includes App builds but we should seperate them in the future
    if (!authenticated && !showInvalidAccessTokenError) {
        return null
    }
    const isInsiderBuild = siteVersion.length > 12 || siteVersion.includes('dev')
    const isVersionCompatible = isInsiderBuild || siteVersion >= '5.1.0'
    const isVersionBeforeCody = !isVersionCompatible && siteVersion < '5.0.0'
    // When doesn't have a valid token
    if (showInvalidAccessTokenError) {
        return <p className={styles.error}>{ERROR_MESSAGES.INVALID}</p>
    }
    // When authenticated but doesn't have a verified email
    if (authenticated && requiresVerifiedEmail && !hasVerifiedEmail) {
        return <p className={styles.error}>{ERROR_MESSAGES.EMAIL_NOT_VERIFIED}</p>
    }
    // When version is lower than 5.0.0
    if (isVersionBeforeCody) {
        return <p className={styles.error}>{ERROR_MESSAGES.VERSION.replace('$SITEVERSION', siteVersion)}</p>
    }
    // When version is compatible but Cody is not enabled
    if (isVersionCompatible && !siteHasCodyEnabled) {
        return <p className={styles.error}>{ERROR_MESSAGES.DISABLED}</p>
    }
    return null
}
