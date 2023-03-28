import { VSCodeButton } from '@vscode/webview-ui-toolkit/react'

import './Settings.css'

import { View } from './utils/types'

interface SettingsProps {
    onLogout: () => void
    setView: (view: View) => void
}

export const Settings: React.FunctionComponent<React.PropsWithChildren<SettingsProps>> = ({ onLogout, setView }) => (
    <div className="inner-container">
        <div className="non-transcript-container">
            <div className="settings">
                <VSCodeButton className="logout-button" type="button" onClick={() => setView('about')}>
                    About
                </VSCodeButton>
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
