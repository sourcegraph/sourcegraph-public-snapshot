import { VSCodeButton } from '@vscode/webview-ui-toolkit/react'

import { getVSCodeAPI } from './utils/VSCodeApi'

interface ConnectAppProps {
    isAppInstalled: boolean
}

export const ConnectApp: React.FunctionComponent<ConnectAppProps> = props => {
    const vscodeAPI = getVSCodeAPI()
    return (
        <p>
            {props.isAppInstalled ? (
                <VSCodeButton
                    type="button"
                    onClick={e =>
                        vscodeAPI.postMessage({
                            command: 'links',
                            value: 'sourcegraph://user/settings/tokens/new/callback?requestFrom=CODY',
                        })
                    }
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
