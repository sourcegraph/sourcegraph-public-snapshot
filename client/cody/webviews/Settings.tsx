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
}) => (
    <div className="inner-container">
        <div className="non-transcript-container">
            <div className="settings">
                {serverEndpoint && <p>🟢 Connected to {serverEndpoint}</p>}
                <VSCodeButton className="logout-button" type="button" onClick={onLogout}>
                    Logout
                </VSCodeButton>
            </div>
        </div>
    </div>
)
