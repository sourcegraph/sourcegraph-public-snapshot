import { VSCodeButton } from '@vscode/webview-ui-toolkit/react'

import { APP_CALLBACK_URL, APP_DOWNLOAD_URLS, APP_LANDING_URL, APP_OPEN_URL } from '../src/chat/protocol'

import { VSCodeWrapper } from './utils/VSCodeApi'

interface ConnectAppProps {
    vscodeAPI: VSCodeWrapper
    isAppInstalled: boolean
    isOSSupported: boolean
    appOS?: string
    appArch?: string
    callbackScheme?: string
    isAppRunning: boolean
    isAppAuthenticated: boolean
}

export const ConnectApp: React.FunctionComponent<ConnectAppProps> = ({
    vscodeAPI,
    isAppInstalled,
    isAppRunning = false,
    isAppAuthenticated,
    isOSSupported,
    appOS = '',
    appArch = '',
    callbackScheme,
}) => {
    const inDownloadMode = !isAppInstalled && isOSSupported && !isAppRunning
    const buttonText = inDownloadMode
        ? 'Download Cody App'
        : isAppRunning
        ? isAppAuthenticated
            ? 'Connect Cody App'
            : 'Open Cody App'
        : 'Open Cody App'
    const buttonIcon = inDownloadMode ? 'cloud-download' : isAppRunning ? 'link' : 'rocket'
    let callbackUri: URL
    if (!isAppInstalled) {
        // Open landing page if download link for user's arch cannot be found
        const DOWNLOAD_URL = APP_DOWNLOAD_URLS[appOS]?.[appArch] || APP_LANDING_URL.href
        callbackUri = new URL(DOWNLOAD_URL)
    } else if (!isAppRunning) {
        // If the user already has the app installed, open the callback URL directly to authorize.
        callbackUri = new URL(APP_CALLBACK_URL.href)
        callbackUri.searchParams.append('requestFrom', callbackScheme === 'vscode-insiders' ? 'CODY_INSIDERS' : 'CODY')
    } else {
        // The user has app installed and authorized, but has not authenticated in app.
        callbackUri = new URL(APP_OPEN_URL)
    }

    // Use postMessage to open because it won't open otherwise due to the sourcegraph:// scheme.
    const authApp = (url: string): void =>
        vscodeAPI.postMessage({
            command: 'auth',
            type: 'app',
            endpoint: url,
        })

    return (
        <div>
            <VSCodeButton type="button" disabled={!isOSSupported} onClick={() => authApp(callbackUri.href)}>
                <i className={'codicon codicon-' + buttonIcon} slot="start" />
                {buttonText}
            </VSCodeButton>
        </div>
    )
}
