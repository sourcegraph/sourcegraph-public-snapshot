import { VSCodeButton } from '@vscode/webview-ui-toolkit/react'

import { APP_CALLBACK_URL, APP_DOWNLOAD_URLS, APP_LANDING_URL } from '../src/chat/protocol'

import { VSCodeWrapper } from './utils/VSCodeApi'

import styles from './ConnectApp.module.css'

interface ConnectAppProps {
    vscodeAPI: VSCodeWrapper
    isAppInstalled: boolean
    isOSSupported: boolean
    appOS?: string
    appArch?: string
    callbackScheme?: string
    isAppRunning: boolean
}

export const ConnectApp: React.FunctionComponent<ConnectAppProps> = ({
    vscodeAPI,
    isAppInstalled,
    isAppRunning = false,
    isOSSupported,
    appOS = '',
    appArch = '',
    callbackScheme,
}) => {
    const inDownloadMode = !isAppInstalled && isOSSupported && !isAppRunning
    const buttonText = inDownloadMode ? 'Download Cody App' : isAppRunning ? 'Connect Cody App' : 'Open Cody App'
    const buttonIcon = inDownloadMode ? 'cloud-download' : isAppRunning ? 'link' : 'rocket'
    // Open landing page if download link for user's arch cannot be found
    const DOWNLOAD_URL = APP_DOWNLOAD_URLS[appOS]?.[appArch] || APP_LANDING_URL.href
    // If the user already has the app installed, open the callback URL directly.
    const callbackUri = new URL(APP_CALLBACK_URL.href)
    callbackUri.searchParams.append('requestFrom', callbackScheme === 'vscode-insiders' ? 'CODY_INSIDERS' : 'CODY')

    // Use postMessage to open because it won't open otherwise due to the sourcegraph:// scheme.
    const authApp = (url: string): void =>
        vscodeAPI.postMessage({
            command: 'auth',
            type: 'app',
            endpoint: url,
        })

    return (
        <div className={styles.buttonContainer}>
            <VSCodeButton
                type="button"
                disabled={!isOSSupported}
                onClick={() => authApp(isAppInstalled ? callbackUri.href : DOWNLOAD_URL)}
            >
                <i className={'codicon codicon-' + buttonIcon} slot="start" />
                {buttonText}
            </VSCodeButton>
        </div>
    )
}
