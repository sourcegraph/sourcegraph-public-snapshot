import { lowerCase, upperFirst } from 'lodash'
import * as React from 'react'

import { OptionsHeader, OptionsHeaderProps } from './OptionsHeader'
import { ServerURLForm, ServerURLFormProps } from './ServerURLForm'
import { Omit } from 'utility-types'
import { DEFAULT_SOURCEGRAPH_URL } from '../../shared/util/context'

export interface CurrentTabStatus {
    host: string
    protocol: string
    hasPermissions: boolean
}

export interface OptionsMenuProps
    extends Omit<OptionsHeaderProps, 'onSettingsClick' | 'className'>,
        Omit<ServerURLFormProps, 'className' | 'requestSourcegraphURLPermissions'> {
    toggleFeatureFlag: (key: string) => void
    requestPermissions: (url: string) => void
    featureFlags?: Record<string, boolean>
    currentTabStatus?: CurrentTabStatus
}

const buildFeatureFlagToggleHandler = (key: string, handler: OptionsMenuProps['toggleFeatureFlag']) => () =>
    handler(key)

const isFullPage = (): boolean => !new URLSearchParams(window.location.search).get('popup')

const buildRequestPermissionsHandler = (
    { protocol, host }: Pick<URL, 'protocol' | 'host'>,
    requestPermissions: OptionsMenuProps['requestPermissions']
) => (event: React.MouseEvent) => {
    event.preventDefault()
    requestPermissions(`${protocol}//${host}`)
}

/**
 * A list of protocols where we should *not* show the permissions notification.
 */
const PERMISSIONS_PROTOCOL_BLACKLIST = ['chrome:', 'about:']

export const OptionsMenu: React.FunctionComponent<OptionsMenuProps> = ({
    toggleFeatureFlag,
    featureFlags,
    currentTabStatus,
    requestPermissions,
    sourcegraphURL,
    connectionStatus,
    ...props
}) => {
    const [isSettingsOpen, setIsSettingsOpen] = React.useState<boolean>(false)
    const onSettingsClick = React.useCallback(() => setIsSettingsOpen(!isSettingsOpen), [isSettingsOpen])
    return (
        <div className={`options-menu ${isFullPage() ? 'options-menu--full' : ''}`}>
            <OptionsHeader {...props} onSettingsClick={onSettingsClick} className="options-menu__section" />
            <ServerURLForm
                {...props}
                connectionStatus={connectionStatus}
                sourcegraphURL={sourcegraphURL}
                requestSourcegraphURLPermissions={buildRequestPermissionsHandler(
                    DEFAULT_SOURCEGRAPH_URL,
                    requestPermissions
                )}
                className="options-menu__section"
            />
            {connectionStatus &&
                connectionStatus.type === 'connected' &&
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
                        {Object.entries(featureFlags).map(([key, value]) => (
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
}
