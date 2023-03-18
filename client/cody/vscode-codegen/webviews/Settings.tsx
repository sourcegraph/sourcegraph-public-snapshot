import { VSCodeButton } from '@vscode/webview-ui-toolkit/react'

import { UserLocalHistory } from '../src/editor/LocalStorageProvider'

import './Settings.css'

import { View } from './utils/types'

interface SettingsProps {
    onLogout: () => void
    setView: (view: View) => void
    userHistory: UserLocalHistory | null
}

export const Settings: React.FunctionComponent<React.PropsWithChildren<SettingsProps>> = ({ onLogout, setView }) => (
    <div className="inner-container">
        <div className="non-transcript-container">
            <div className="settings">
                <VSCodeButton className="logout-button" type="button" onClick={() => setView('about')}>
                    About
                </VSCodeButton>
                <VSCodeButton className="logout-button" type="button" disabled={true}>
                    Chat History (Coming soon)
                </VSCodeButton>
                <VSCodeButton className="logout-button" type="button" onClick={onLogout}>
                    Logout
                </VSCodeButton>
            </div>
        </div>
    </div>
)
