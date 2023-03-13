import { VSCodeButton } from '@vscode/webview-ui-toolkit/react'

import './Settings.css'

interface SettingsProps {
    onLogout: () => void
}

export const Settings: React.FunctionComponent<React.PropsWithChildren<SettingsProps>> = ({ onLogout }) => (
    <div className="inner-container">
        <div className="non-transcript-container">
            <div className="settings">
                <VSCodeButton className="logout-button" type="button" onClick={onLogout}>
                    Logout
                </VSCodeButton>
            </div>
        </div>
    </div>
)
