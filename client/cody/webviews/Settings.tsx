import { VSCodeButton } from '@vscode/webview-ui-toolkit/react'

import './Settings.css'

import { useMemo } from 'react'

import { LOCAL_APP_URL } from '../src/chat/protocol'

interface SettingsProps {
    onLogout: () => void
    serverEndpoint?: string
}

export const Settings: React.FunctionComponent<React.PropsWithChildren<SettingsProps>> = ({
    onLogout,
    serverEndpoint,
}) => {
    const isLocalApp = useMemo(() => serverEndpoint === LOCAL_APP_URL.toString(), [serverEndpoint])
    return (
        <div className="inner-container">
            <div className="non-transcript-container">
                <div className="settings">
                    {serverEndpoint && <p>🟢 Connected to {isLocalApp ? 'Sourcegraph App' : serverEndpoint}</p>}
                    <VSCodeButton className="logout-button" type="button" onClick={onLogout}>
                        Logout
                    </VSCodeButton>
                </div>
            </div>
        </div>
    )
}
