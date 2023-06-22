import { VSCodeButton } from '@vscode/webview-ui-toolkit/react'

import { APP_CALLBACK_URL, APP_DOWNLOAD_URLS, APP_LANDING_URL } from '../src/chat/protocol'

import { VSCodeWrapper } from './utils/VSCodeApi'

interface ConnectAppProps {
    vscodeAPI: VSCodeWrapper
    isAppInstalled: boolean
    isOSSupported: boolean
    appOS?: string
    appArch?: string
    callbackScheme?: string
}

export const ConnectApp: React.FunctionComponent<ConnectAppProps> = ({
    vscodeAPI,
    isAppInstalled,
    isOSSupported,
    appOS = '',
    appArch = '',
    callbackScheme,
}) => {
    const inDownloadMode = !isAppInstalled && isOSSupported
    const buttonText = inDownloadMode ? 'Download Cody App' : 'Open Cody App'
    const buttonIcon = inDownloadMode ? 'codicon codicon-cloud-download' : 'codicon codicon-vm-running'
    // Open landing page if download link for user's arch cannot be found
    const DOWNLOAD_URL = APP_DOWNLOAD_URLS[appOS]?.[appArch] || APP_LANDING_URL.href
    // Use postMessage to open because it won't open otherwise due to the sourcegraph:// scheme.
    const openLink = (url: string): void =>
        vscodeAPI.postMessage({
            command: 'links',
            value: url,
        })
    // If the user already has the app installed, open the callback URL directly.
    const callbackUri = new URL(APP_CALLBACK_URL.href)
    callbackUri.searchParams.append('requestFrom', callbackScheme === 'vscode-insiders' ? 'CODY_INSIDERS' : 'CODY')
    return (
        <div>
            <VSCodeButton
                type="button"
                disabled={!isOSSupported}
                onClick={() => openLink(isAppInstalled ? callbackUri.href : DOWNLOAD_URL)}
            >
                <i className={buttonIcon} slot="start" />
                {buttonText}
            </VSCodeButton>
        </div>
    )
}
