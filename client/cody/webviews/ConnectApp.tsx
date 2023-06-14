import { VSCodeButton } from '@vscode/webview-ui-toolkit/react'

import { APP_CALLBACK_URL, APP_DOWNLOAD_URL } from '../src/chat/protocol'

import { VSCodeWrapper } from './utils/VSCodeApi'

interface ConnectAppProps {
    vscodeAPI: VSCodeWrapper
    isAppInstalled: boolean
    isOSSupported: boolean
}

export const ConnectApp: React.FunctionComponent<ConnectAppProps> = ({ vscodeAPI, isAppInstalled, isOSSupported }) => {
    const buttonText = !isAppInstalled && isOSSupported ? 'Download Cody App' : 'Connect Cody App'
    // Use postMessage to open because it won't open otherwise due to the sourcegraph:// scheme.
    const openLink = (url: string): void => {
        vscodeAPI.postMessage({
            command: 'links',
            value: url,
        })
    }

    return (
        <p>
            <VSCodeButton
                type="button"
                disabled={!isOSSupported}
                onClick={() => openLink(isAppInstalled ? APP_CALLBACK_URL.href : APP_DOWNLOAD_URL.href)}
            >
                {buttonText}
            </VSCodeButton>
        </p>
    )
}
