import * as React from 'react'

export interface Props {
    settings: GQL.IOrgSettings | null
}

export const OrgSettingsFile = ({ settings }: Props) => (
    <div className="settings-file">
        <h3> Current Organization Editor Configuration</h3>
        {settings &&
            settings.highlighted && [
                <div
                    key={0}
                    className="settings-file__settings-box"
                    dangerouslySetInnerHTML={{ __html: settings.highlighted }}
                />,
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
