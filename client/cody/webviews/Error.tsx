import { AuthStatus } from '../src/chat/protocol'

import styles from './Login.module.css'

const AUTH_ERRORS = {
    DISABLED: 'Cody is not enabled on your instance. To enable Cody, please contact your site admin.',
    VERSION:
        'Cody is not supported by your Sourcegraph instance version (version: $SITEVERSION). To use Cody, please contact your site admin to upgrade to version 5.1.0 or above.',
    INVALID: 'Invalid credentials. Please retry with a valid instance URL and access token.',
    EMAIL_NOT_VERIFIED: 'Email not verified. Please add a verified email to your Sourcegraph.com account.',
    APP_NOT_RUNNING: 'Cody App is not running. Please open the Cody App to continue.',
    INVALID_URL: 'Connection failed due to invalid URL. Please enter a valid Sourcegraph instance URL.',
}
export const ErrorContainer: React.FunctionComponent<{
    authStatus: AuthStatus
    isApp: {
        isInstalled: boolean
        isRunning: boolean
        isAuthenticated: boolean
    }
    endpoint?: string | null
}> = ({ authStatus, isApp, endpoint }) => {
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
    // Errors for app are handled in the ConnectApp component
    if (isApp.isInstalled && (!isApp.isRunning || !isApp.isAuthenticated)) {
        return null
    }

    // new users will not have an endpoint
    if (!endpoint) {
        return null
    }
    const isInsiderBuild = siteVersion.length > 12 || siteVersion.includes('dev')
    const isVersionCompatible = isInsiderBuild || siteVersion >= '5.1.0'
    const isVersionBeforeCody = !isVersionCompatible && siteVersion < '5.0.0'
    const prefix = `Failed: ${isApp.isRunning ? 'Cody App' : endpoint}`
    // When doesn't have a valid token
    if (showInvalidAccessTokenError) {
        return (
            <p className={styles.error}>
                <p>{prefix}</p>
                <p>{AUTH_ERRORS.INVALID}</p>
            </p>
        )
    }
    // When authenticated but doesn't have a verified email
    if (authenticated && requiresVerifiedEmail && !hasVerifiedEmail) {
        return (
            <p className={styles.error}>
                <p>{prefix}</p>
                <p>{AUTH_ERRORS.EMAIL_NOT_VERIFIED}</p>
            </p>
        )
    }
    // When version is lower than 5.0.0
    if (isVersionBeforeCody) {
        return (
            <p className={styles.error}>
                <p>{prefix}</p>
                <p>{AUTH_ERRORS.VERSION.replace('$SITEVERSION', siteVersion)}</p>
            </p>
        )
    }
    // When version is compatible but Cody is not enabled
    if (isVersionCompatible && !siteHasCodyEnabled) {
        return (
            <p className={styles.error}>
                <p>{prefix}</p>
                <p>{AUTH_ERRORS.DISABLED}</p>
            </p>
        )
    }
    return null
}
