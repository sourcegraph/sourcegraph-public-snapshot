import * as React from 'react'
import { SettingsFile } from '../SettingsFile'

interface Props {
    settings: GQL.ISettings | null
}

export const UserSettingsFile = ({ settings }: Props) => (
    <div className="settings-file">
        <h3>Current user configuration</h3>
        {settings && settings.highlighted ? (
            <SettingsFile>{settings.highlighted}</SettingsFile>
        ) : (
            <p className="form-text">No user configuration settings exist yet.</p>
        )}
    </div>
)
