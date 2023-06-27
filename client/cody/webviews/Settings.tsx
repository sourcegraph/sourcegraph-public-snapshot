import { VSCodeButton } from '@vscode/webview-ui-toolkit/react'

import './Settings.css'

import { isLocalApp } from '../src/chat/protocol'

interface SettingsProps {
    onLogout: () => void
    endpoint: string
    version?: string
}

export const Settings: React.FunctionComponent<React.PropsWithChildren<SettingsProps>> = ({
    onLogout,
    endpoint,
    version,
}) => (
    <div className="inner-container">
        <div className="non-transcript-container">
            <div className="settings">
                {endpoint && <p>Signed in to {isLocalApp(endpoint) ? 'Cody App' : endpoint}</p>}
                {version && <p>Cody v{version}</p>}
                <VSCodeButton className="logout-button" type="button" onClick={onLogout}>
                    Sign Out
                </VSCodeButton>
            </div>
        </div>
    </div>
)
