import { VSCodeButton } from '@vscode/webview-ui-toolkit/react'

import { APP_CALLBACK_URL, APP_DOWNLOAD_URLS, APP_LANDING_URL } from '../src/chat/protocol'

import { VSCodeWrapper } from './utils/VSCodeApi'

interface ConnectAppProps {
    vscodeAPI: VSCodeWrapper
    isAppInstalled: boolean
    isOSSupported: boolean
    appOS: string
    appArch: string
}

export const ConnectApp: React.FunctionComponent<ConnectAppProps> = ({
    vscodeAPI,
    isAppInstalled,
    isOSSupported,
    appOS,
    appArch,
}) => {
    const buttonText = !isAppInstalled && isOSSupported ? 'Download Cody App' : 'Connect Cody App'
    // Open landing page if download link for user's arch cannot be found
    const DOWNLOAD_URL = APP_DOWNLOAD_URLS[appOS] ? APP_DOWNLOAD_URLS[appOS][appArch] : APP_LANDING_URL.href
    // Use postMessage to open because it won't open otherwise due to the sourcegraph:// scheme.
    const openLink = (url: string): void =>
        vscodeAPI.postMessage({
            command: 'links',
            value: url,
        })

    return (
        <div>
            <VSCodeButton
                type="button"
                disabled={!isOSSupported}
                onClick={() => openLink(isAppInstalled ? APP_CALLBACK_URL.href : DOWNLOAD_URL)}
            >
                <i className="codicon codicon-cloud-download" slot="start" />
                {buttonText}
            </VSCodeButton>
        </div>
    )
}
