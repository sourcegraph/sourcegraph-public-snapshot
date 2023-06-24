import { VSCodeButton } from '@vscode/webview-ui-toolkit/react'

import { APP_CALLBACK_URL, APP_DOWNLOAD_URLS, APP_LANDING_URL } from '../src/chat/protocol'

interface ConnectAppProps {
    isAppInstalled: boolean
    isOSSupported: boolean
    appOS?: string
    appArch?: string
    callbackScheme?: string
    isAppRunning: boolean
    onAppButtonClick: (type: string) => void
}

export const ConnectApp: React.FunctionComponent<ConnectAppProps> = ({
    isAppInstalled,
    isAppRunning = false,
    isOSSupported,
    appOS = '',
    appArch = '',
    callbackScheme,
    onAppButtonClick,
}) => {
    const inDownloadMode = !isAppInstalled && isOSSupported && !isAppRunning
    const buttonText = inDownloadMode ? 'Download Cody App' : isAppRunning ? 'Connect Cody App' : 'Open Cody App'
    const buttonIcon = inDownloadMode ? 'cloud-download' : isAppRunning ? 'link' : 'rocket'
    // Open landing page if download link for user's arch cannot be found
    const DOWNLOAD_URL = APP_DOWNLOAD_URLS[appOS]?.[appArch] || APP_LANDING_URL.href
    // If the user already has the app installed, open the callback URL directly.
    const callbackUri = new URL(APP_CALLBACK_URL.href)
    callbackUri.searchParams.append('requestFrom', callbackScheme === 'vscode-insiders' ? 'CODY_INSIDERS' : 'CODY')

    return (
        <div>
            <VSCodeButton
                type="button"
                disabled={!isOSSupported}
                onClick={() => onAppButtonClick(isAppInstalled ? callbackUri.href : DOWNLOAD_URL)}
            >
                <i className={'codicon codicon-' + buttonIcon} slot="start" />
                {buttonText}
            </VSCodeButton>
        </div>
    )
}
