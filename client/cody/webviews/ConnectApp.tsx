import { VSCodeButton } from '@vscode/webview-ui-toolkit/react'

import { VSCodeWrapper } from './utils/VSCodeApi'

interface ConnectAppProps {
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
            <VSCodeButton
                type="button"
                onClick={() => openLink('sourcegraph://user/settings/tokens/new/callback?requestFrom=CODY')}
            >
                Connect Sourcegraph App
            </VSCodeButton>
        </p>
    )
}
