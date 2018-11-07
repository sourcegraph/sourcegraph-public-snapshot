import * as React from 'react'
import { JSONEditor, JSONEditorProps } from '../../shared/components/JSONEditor'
import { OptionsActionButton } from './ActionButton'
import { OptionsHeader, OptionsHeaderProps } from './Header'
import { ServerURLForm, ServerURLFormProps } from './ServerURLForm'

export interface OptionsMenuProps
    extends OptionsHeaderProps,
        Pick<ServerURLFormProps, Exclude<keyof ServerURLFormProps, 'value' | 'onChange' | 'onSubmit'>>,
        Pick<JSONEditorProps, Exclude<keyof JSONEditorProps, 'value' | 'onChange'>> {
    sourcegraphURL: ServerURLFormProps['value']
    onURLChange: ServerURLFormProps['onChange']
    onURLSubmit: ServerURLFormProps['onSubmit']

    isSettingsOpen: boolean
    settingsHaveChanged: boolean
    settings: JSONEditorProps['value']
    onSettingsChange: JSONEditorProps['onChange']
    onSettingsSave: () => void
}

export const OptionsMenu: React.SFC<OptionsMenuProps> = ({
    sourcegraphURL,
    onURLChange,
    onURLSubmit,
    settings,
    onSettingsChange,
    onSettingsSave,
    isSettingsOpen,
    settingsHaveChanged,
    ...props
}) => (
    <div className="options-menu">
        <OptionsHeader {...props} className="options-menu__section options-menu__no-border" />
        <ServerURLForm
            {...props}
            value={sourcegraphURL}
            onChange={onURLChange}
            onSubmit={onURLSubmit}
            className="options-menu__section"
        />
        {isSettingsOpen && (
            <div className="options-menu__section">
                <label>Extended configuration</label>
                <JSONEditor {...props} value={settings} onChange={onSettingsChange} />
                {settingsHaveChanged && (
                    <OptionsActionButton className="options-menu__section__button" onClick={onSettingsSave}>
                        Update configuration JSON
                    </OptionsActionButton>
                )}
            </div>
        )}
    </div>
)
