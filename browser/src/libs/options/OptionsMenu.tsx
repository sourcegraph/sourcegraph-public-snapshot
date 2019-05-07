import { lowerCase, upperFirst } from 'lodash'
import * as React from 'react'

import { OptionsHeader, OptionsHeaderProps } from './OptionsHeader'
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

/**
 * A list of protocols where we should *not* show the permissions notification.
 */
const PERMISSIONS_PROTOCOL_BLACKLIST = ['chrome:', 'about:']

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
        <ServerURLForm
            {...props}
            value={sourcegraphURL}
            onChange={onURLChange}
            onSubmit={onURLSubmit}
            status={status}
            requestPermissions={requestPermissions}
            className="options-menu__section"
        />
        {status === 'connected' &&
            currentTabStatus &&
            !currentTabStatus.hasPermissions &&
            !PERMISSIONS_PROTOCOL_BLACKLIST.includes(currentTabStatus.protocol) && (
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
        {isSettingsOpen && featureFlags && (
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
    </div>
)
