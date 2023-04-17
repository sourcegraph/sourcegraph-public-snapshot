import { VSCodeButton } from '@vscode/webview-ui-toolkit/react'

import { View } from './NavBar'

import './Settings.css'

interface SettingsProps {
    onLogout: () => void
    setView: (view: View) => void
}

export const Settings: React.FunctionComponent<React.PropsWithChildren<SettingsProps>> = ({ onLogout, setView }) => (
    <div className="inner-container">
        <div className="non-transcript-container">
            <div className="settings">
                <VSCodeButton className="logout-button" type="button" onClick={() => setView('history')}>
                    Chat History
                </VSCodeButton>
                <VSCodeButton className="logout-button" type="button" onClick={onLogout}>
                    Logout
                </VSCodeButton>
            </div>
        </div>
    </div>
)
