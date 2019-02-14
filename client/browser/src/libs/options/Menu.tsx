import { lowerCase, upperFirst } from 'lodash'
import * as React from 'react'

import { OptionsHeader, OptionsHeaderProps } from './Header'
import { ServerURLForm, ServerURLFormProps } from './ServerURLForm'

interface ConfigurableFeatureFlag {
    key: string
    value: boolean
}

export interface OptionsMenuProps
    extends OptionsHeaderProps,
        Pick<ServerURLFormProps, Exclude<keyof ServerURLFormProps, 'value' | 'onChange' | 'onSubmit'>> {
    sourcegraphURL: ServerURLFormProps['value']
    onURLChange: ServerURLFormProps['onChange']
    onURLSubmit: ServerURLFormProps['onSubmit']

    isSettingsOpen?: boolean
    toggleFeatureFlag: (key: string) => void
    featureFlags?: ConfigurableFeatureFlag[]
}

const buildFeatureFlagToggleHandler = (key: string, handler: OptionsMenuProps['toggleFeatureFlag']) => () =>
    handler(key)

const isFullPage = (): boolean => !new URLSearchParams(window.location.search).get('popup')

export const OptionsMenu: React.FunctionComponent<OptionsMenuProps> = ({
    sourcegraphURL,
    onURLChange,
    onURLSubmit,
    isSettingsOpen,
    toggleFeatureFlag,
    featureFlags,
    ...props
}) => (
    <div className={`options-menu ${isFullPage() ? 'options-menu--full' : ''}`}>
        <OptionsHeader {...props} className="options-menu__section options-menu__no-border" />
        <ServerURLForm
            {...props}
            value={sourcegraphURL}
            onChange={onURLChange}
            onSubmit={onURLSubmit}
            className="options-menu__section"
        />
        {isSettingsOpen && featureFlags && (
            <div className="options-menu__section">
                <label>Configuration</label>
                <div>
                    {featureFlags.map(({ key, value }) => (
                        <div className="form-check" key={key}>
                            <label className="form-check-label">
                                <input
                                    id={key}
                                    onClick={buildFeatureFlagToggleHandler(key, toggleFeatureFlag)}
                                    className="form-check-input"
                                    type="checkbox"
                                    checked={value}
                                />{' '}
                                {upperFirst(lowerCase(key))}
                            </label>
                        </div>
                    ))}
                </div>
            </div>
        )}
    </div>
)
