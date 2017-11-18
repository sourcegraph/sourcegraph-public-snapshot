import * as React from 'react'
import { SettingsFile } from '../SettingsFile'

export interface Props {
    settings: GQL.IOrgSettings | null
}

export const OrgSettingsFile = ({ settings }: Props) => (
    <div className="settings-file">
        <h3> Current Organization Editor Configuration</h3>
        {settings &&
            settings.highlighted && [
                <SettingsFile key={0}>{settings.highlighted}</SettingsFile>,
                <small key={1} className="form-text">
                    Run the 'Preferences: Open Organization Settings' command inside of Sourcegraph Editor to change
                    this configuration.
                </small>,
            ]}
        {!settings && (
            <p className="form-text">
                This organization hasn't created a configuration file yet. Run the 'Preferences: Open Organization
                Settings' command inside of Sourcegraph Editor to create one.
            </p>
        )}
    </div>
)
