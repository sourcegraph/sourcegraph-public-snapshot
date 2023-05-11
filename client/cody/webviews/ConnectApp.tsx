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
            <VSCodeButton
                type="button"
                onClick={() =>
                    openLink(
                        props.isAppInstalled
                            ? 'sourcegraph://user/settings/tokens/new/callback?requestFrom=CODY'
                            : 'https://about.sourcegraph.com/app'
                    )
                }
            >
                {props.isAppInstalled ? 'Connect Sourcegraph App' : 'Get Sourcegraph App'}
            </VSCodeButton>
        </p>
    )
}
