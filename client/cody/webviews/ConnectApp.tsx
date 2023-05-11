import { VSCodeButton } from '@vscode/webview-ui-toolkit/react'

import { VSCodeWrapper } from './utils/VSCodeApi'

interface ConnectAppProps {
    isAppInstalled: boolean
    vscodeAPI: VSCodeWrapper
}

export const ConnectApp: React.FunctionComponent<ConnectAppProps> = props => {
    // Use postMessage to open because it won't open otherwise due to the sourcegraph:// scheme.
    const openLink = (url: string): void => {
        props.vscodeAPI.postMessage({
            command: 'links',
            value: url,
        })
    }

    return (
        <p>
            {props.isAppInstalled ? (
                <VSCodeButton
                    type="button"
                    onClick={e => openLink('sourcegraph://user/settings/tokens/new/callback?requestFrom=CODY')}
                >
                    Connect Sourcegraph App
                </VSCodeButton>
            ) : (
                <a href="https://about.sourcegraph.com/app">
                    <VSCodeButton type="button">Get Sourcegraph App</VSCodeButton>
                </a>
            )}
        </p>
    )
}
