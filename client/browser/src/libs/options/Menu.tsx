import { lowerCase, upperFirst } from 'lodash'
import SettingsOutlineIcon from 'mdi-react/SettingsOutlineIcon'
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
    onSettingsClick: (event: React.MouseEvent<HTMLButtonElement>) => void

    sourcegraphURL: ServerURLFormProps['value']
    onURLChange: ServerURLFormProps['onChange']
    onURLSubmit: ServerURLFormProps['onSubmit']

    isSettingsOpen?: boolean
    toggleFeatureFlag: (key: string) => void
    featureFlags?: ConfigurableFeatureFlag[]
    currentTabStatus?: {
        host: string
        protocol: string
        hasPermissions: boolean
    }
}

const buildFeatureFlagToggleHandler = (key: string, handler: OptionsMenuProps['toggleFeatureFlag']) => () =>
    handler(key)

const isFullPage = (): boolean => !new URLSearchParams(window.location.search).get('popup')

const buildRequestPermissionsHandler = (
    { protocol, host }: NonNullable<OptionsMenuProps['currentTabStatus']>,
    requestPermissions: OptionsMenuProps['requestPermissions']
) => (event: React.MouseEvent<HTMLAnchorElement, MouseEvent>) => {
    event.preventDefault()
    requestPermissions(`${protocol}//${host}`)
}

export const OptionsMenu: React.FunctionComponent<OptionsMenuProps> = ({
    sourcegraphURL,
    onURLChange,
    onURLSubmit,
    isSettingsOpen,
    toggleFeatureFlag,
    featureFlags,
    status,
    requestPermissions,
    currentTabStatus,
    ...props
}) => (
    <div className={`options-menu ${isFullPage() ? 'options-menu--full' : ''}`}>
        <OptionsHeader {...props} className="options-menu__section" />
        {status === 'connected' && currentTabStatus && !currentTabStatus.hasPermissions && (
            <div className="options-menu__section">
                <div className="alert alert-danger">
                    Sourcegraph is not enabled on <strong>{currentTabStatus.host}</strong>.{' '}
                    <a
                        href=""
                        onClick={buildRequestPermissionsHandler(currentTabStatus, requestPermissions)}
                        className="request-permissions__test"
                    >
                        Grant permissions
                    </a>{' '}
                    to enable Sourcegraph.
                </div>
            </div>
        )}
        <div className="options-menu__section d-flex justify-content-between align-items-center">
            <span>
                <img
                    src="https://lh4.googleusercontent.com/-78ZFwlXv0dc/AAAAAAAAAAI/AAAAAAAAAm8/LA_GSrgCsOU/s50/photo.jpg"
                    className="rounded mr-1"
                    style={{ height: '32px' }}
                />{' '}
                <strong>@sqs</strong>
            </span>
            <button className="options-menu__settings btn btn-icon" onClick={props.onSettingsClick}>
                <SettingsOutlineIcon className="icon-inline" />
            </button>
        </div>
        {isSettingsOpen && (
            <>
                <ServerURLForm
                    {...props}
                    value={sourcegraphURL}
                    onChange={onURLChange}
                    onSubmit={onURLSubmit}
                    status={status}
                    requestPermissions={requestPermissions}
                    className="options-menu__section"
                />
                {featureFlags && (
                    <div className="options-menu__section">
                        <label>Configuration</label>
                        <div>
                            {featureFlags.map(({ key, value }) => (
                                <div className="form-check" key={key}>
                                    <label className="form-check-label">
                                        <input
                                            id={key}
                                            onChange={buildFeatureFlagToggleHandler(key, toggleFeatureFlag)}
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
            </>
        )}
    </div>
)
