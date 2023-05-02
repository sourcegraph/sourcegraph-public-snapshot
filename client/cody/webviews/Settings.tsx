import { VSCodeButton } from '@vscode/webview-ui-toolkit/react'

import './Settings.css'

import type { VSCodeWrapper } from './utils/VSCodeApi'

interface SettingsProps {
    onLogout: () => void
    serverEndpoint?: string
    vscodeAPI: VSCodeWrapper
}

export const Settings: React.FunctionComponent<React.PropsWithChildren<SettingsProps>> = ({
    onLogout,
    serverEndpoint,
    vscodeAPI,
}) => (
    <div className="inner-container">
        <div className="non-transcript-container">
            <div className="settings">
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

                {serverEndpoint && <p>🟢 Connected to {serverEndpoint}</p>}
                <VSCodeButton className="logout-button" type="button" onClick={onLogout}>
                    Logout
                </VSCodeButton>
            </div>
        </div>
    </div>
)
