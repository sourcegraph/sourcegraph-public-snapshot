import { useCallback, useState } from 'react'

import { VSCodeButton } from '@vscode/webview-ui-toolkit/react'
import classNames from 'classnames'

import { AuthStatus, DOTCOM_CALLBACK_URL, DOTCOM_URL } from '../src/chat/protocol'

import { ConnectApp } from './ConnectApp'
import { VSCodeWrapper } from './utils/VSCodeApi'

import styles from './Login.module.css'

interface LoginProps {
    authStatus?: AuthStatus
    serverEndpoint?: string
    isAppInstalled: boolean
    isAppRunning?: boolean
    vscodeAPI: VSCodeWrapper
    callbackScheme?: string
    appOS?: string
    appArch?: string
    isAppConnectEnabled?: boolean
}

const APP_MESSAGES = {
    getStarted: 'Cody for VS Code requires the Cody desktop app to enable context fetching for your private code.',
    download: 'Download and run the Cody desktop app to configure your local code graph.',
    connectApp: 'Cody App detected. All that’s left is to do is connect VS Code with Cody App.',
    appNotRunning: 'Cody for VS Code requires the Cody desktop app to enable context fetching for your private code.',
    comingSoon:
        'We’re working on bringing Cody App to your platform. In the meantime, you can try Cody with open source repositories by signing in to Sourcegraph.com.',
}

const ERROR_MESSAGES = {
    DISABLED: 'Cody is not enabled on your instance. To enable Cody, please contact your site admin.',
    VERSION:
        'Cody is not supported by your Sourcegraph instance version (version: $SITEVERSION). To use Cody, please contact your site admin to upgrade to version 5.1.0 or above.',
    INVALID: 'Invalid credentials. Please check the Sourcegraph instance URL and access token.',
    EMAIL_NOT_VERIFIED: 'Email not verified. Please add a verified email to your Sourcegraph.com account.',
}

export const Login: React.FunctionComponent<React.PropsWithChildren<LoginProps>> = ({
    authStatus,
    serverEndpoint,
    isAppInstalled,
    isAppRunning,
    vscodeAPI,
    callbackScheme,
    appOS,
    appArch,
    isAppConnectEnabled,
}) => {
    const [endpoint, setEndpoint] = useState(serverEndpoint || DOTCOM_URL.href)

    const isOSSupported = appOS === 'darwin' && appArch === 'arm64'

    const loginWithDotCom = useCallback(() => {
        const callbackUri = new URL(DOTCOM_CALLBACK_URL.href)
        callbackUri.searchParams.append('requestFrom', callbackScheme === 'vscode-insiders' ? 'CODY_INSIDERS' : 'CODY')
        setEndpoint(DOTCOM_URL.href)
        vscodeAPI.postMessage({ command: 'links', value: callbackUri.href })
    }, [callbackScheme, vscodeAPI])

    const onFooterButtonClick = useCallback(
        (title: 'signin' | 'support') => {
            vscodeAPI.postMessage({ command: 'auth', type: title })
        },
        [vscodeAPI]
    )

    const title = isAppInstalled ? (isAppRunning ? 'Connect with Cody App' : 'Cody App Not Running') : 'Get Started'

    const SigninWithApp: React.FunctionComponent = () => (
        <div className={styles.sectionsContainer}>
            <section className={classNames(styles.section, isOSSupported ? styles.codyGradient : null)}>
                <h2 className={styles.sectionHeader}>{isAppInstalled ? title : 'Get Started'}</h2>
                <p className={styles.openMessage}>
                    {!isAppInstalled
                        ? APP_MESSAGES.getStarted
                        : !isAppRunning
                        ? APP_MESSAGES.appNotRunning
                        : APP_MESSAGES.connectApp}
                </p>
                {!isAppInstalled && <small className={styles.openMessage}>APP_MESSAGES.download</small>}
                <ConnectApp
                    isAppInstalled={isAppInstalled}
                    vscodeAPI={vscodeAPI}
                    isOSSupported={isOSSupported}
                    appOS={appOS}
                    appArch={appArch}
                    callbackScheme={callbackScheme}
                />
                {!isOSSupported && (
                    <small>
                        Sorry, {appOS} {appArch} is not yet supported.
                    </small>
                )}
            </section>
            {!isOSSupported && (
                <section className={classNames(styles.section, styles.codyGradient)}>
                    <h2 className={styles.sectionHeader}>Cody App for {appOS} coming soon</h2>
                    <p className={styles.openMessage}>{APP_MESSAGES.comingSoon}</p>
                    <VSCodeButton className={styles.button} type="button" onClick={() => loginWithDotCom()}>
                        Signin with Sourcegraph.com
                    </VSCodeButton>
                </section>
            )}
        </div>
    )

    const SigninWithoutApp: React.FunctionComponent = () => (
        <div className={styles.sectionsContainer}>
            {/* Sourcegraph Enterprise */}
            <section className={classNames(styles.section, styles.greyGradient)}>
                <h2 className={styles.sectionHeader}>Sourcegraph Enterprise</h2>
                <p className={styles.openMessage}>
                    Sign in by entering an access token created through your user settings on Sourcegraph.
                </p>
                <VSCodeButton className={styles.button} type="button" onClick={() => onFooterButtonClick('signin')}>
                    Continue with Access Token
                </VSCodeButton>
            </section>
            {/* Sourcegraph DotCom */}
            <section className={classNames(styles.section, styles.greyGradient)}>
                <h2 className={styles.sectionHeader}>Sourcegraph.com</h2>
                <p className={styles.openMessage}>
                    Cody for open source code is available to all users with a Sourcegraph.com account.
                </p>
                <VSCodeButton className={styles.button} type="button" onClick={() => loginWithDotCom()}>
                    Continue with Sourcegraph.com
                </VSCodeButton>
            </section>
            {/* Cody App */}
            <section className={classNames(styles.section, styles.codyGradient)}>
                <h2 className={styles.sectionHeader}>Cody App</h2>
                <p className={styles.openMessage}>{APP_MESSAGES.getStarted}</p>
                <ConnectApp
                    isAppInstalled={isAppInstalled}
                    vscodeAPI={vscodeAPI}
                    isOSSupported={isOSSupported}
                    appOS={appOS || ''}
                    appArch={appArch || ''}
                />
            </section>
        </div>
    )

    return (
        <div className={styles.container}>
            {authStatus && <ErrorContainer authStatus={authStatus} endpoint={endpoint} />}
            {isAppConnectEnabled ? <SigninWithApp /> : <SigninWithoutApp />}
            <footer className={styles.footer}>
                <VSCodeButton className={styles.button} type="button" onClick={() => onFooterButtonClick('signin')}>
                    Other Sign In Options…
                </VSCodeButton>
                <VSCodeButton className={styles.button} type="button" onClick={() => onFooterButtonClick('support')}>
                    Feedback & Support
                </VSCodeButton>
            </footer>
        </div>
    )
}

const ErrorContainer: React.FunctionComponent<{ authStatus: AuthStatus; endpoint: string }> = ({ authStatus }) => {
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
