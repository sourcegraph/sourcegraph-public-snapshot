import { upperFirst, words } from 'lodash'
import * as React from 'react'
import { FormGroup, Input, Label } from 'reactstrap'

import { OptionsHeader, OptionsHeaderProps } from './Header'
import { ServerURLForm, ServerURLFormProps } from './ServerURLForm'

export interface OptionsMenuProps
    extends OptionsHeaderProps,
        Pick<ServerURLFormProps, Exclude<keyof ServerURLFormProps, 'value' | 'onChange' | 'onSubmit'>> {
    sourcegraphURL: ServerURLFormProps['value']
    onURLChange: ServerURLFormProps['onChange']
    onURLSubmit: ServerURLFormProps['onSubmit']

    isSettingsOpen?: boolean
    toggleFeatureFlag: (key: string) => void
    featureFlags: { key: string; value: boolean }[]
}

const buildFeatureFlagToggleHandler = (key: string, handler: OptionsMenuProps['toggleFeatureFlag']) => () =>
    handler(key)

export const OptionsMenu: React.SFC<OptionsMenuProps> = ({
    sourcegraphURL,
    onURLChange,
    onURLSubmit,
    isSettingsOpen,
    toggleFeatureFlag,
    featureFlags,
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
                <label>Experimental configuration</label>
                <div>
                    {featureFlags.map(({ key, value }) => (
                        <FormGroup check={true}>
                            <Label check={true}>
                                <Input
                                    onClick={buildFeatureFlagToggleHandler(key, toggleFeatureFlag)}
                                    defaultChecked={value}
                                    type="checkbox"
                                />{' '}
                                {words(key)
                                    .map(upperFirst)
                                    .join(' ')}
                            </Label>
                        </FormGroup>
                    ))}
                </div>
            </div>
        )}
    </div>
)
