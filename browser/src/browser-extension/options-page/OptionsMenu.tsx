import * as React from 'react'
import { OptionsHeader, OptionsHeaderProps } from './OptionsHeader'
import { ServerUrlForm, ServerUrlFormProps } from './ServerUrlForm'
import { OptionFlagWithValue } from '../../shared/util/optionFlags'

export interface OptionsMenuProps
    extends OptionsHeaderProps,
        Pick<ServerUrlFormProps, Exclude<keyof ServerUrlFormProps, 'value' | 'onChange' | 'onSubmit'>> {
    sourcegraphURL: ServerUrlFormProps['value']
    onURLChange: ServerUrlFormProps['onChange']
    onURLSubmit: ServerUrlFormProps['onSubmit']

    isOptionsMenuExpanded?: boolean
    isActivated: boolean
    onChangeOptionFlag: (key: string, value: boolean) => void
    optionFlags?: OptionFlagWithValue[]
    currentTabStatus?: {
        host: string
        protocol: string
        hasPermissions: boolean
    }
}

/**
 * Determine if the options menu is being showed as a popup panel (opened via
 * the toolbar icon) or as a full page (opened via the options URL on a page of
 * its owen)
 */
const isFullPage = (): boolean => !new URLSearchParams(window.location.search).get('popup')

const buildRequestPermissionsHandler = (
    { protocol, host }: NonNullable<OptionsMenuProps['currentTabStatus']>,
    requestPermissions: OptionsMenuProps['requestPermissions']
) => (event: React.MouseEvent) => {
    event.preventDefault()
    requestPermissions(`${protocol}//${host}`)
}

/**
 * A list of protocols where we should *not* show the permissions notification.
 */
const PERMISSIONS_PROTOCOL_BLOCKLIST = new Set(['chrome:', 'about:'])

export const OptionsMenu: React.FunctionComponent<OptionsMenuProps> = ({
    sourcegraphURL,
    onURLChange,
    onURLSubmit,
    isOptionsMenuExpanded,
    isActivated,
    onChangeOptionFlag,
    optionFlags: options,
    status,
    requestPermissions,
    currentTabStatus,
    ...props
}) => (
    <div className={`options-menu ${isFullPage() ? 'options-menu--full' : ''}`}>
        <OptionsHeader {...props} isActivated={isActivated} className="options-menu__section" />
        <ServerUrlForm
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
            !PERMISSIONS_PROTOCOL_BLOCKLIST.has(currentTabStatus.protocol) && (
                <div className="options-menu__section">
                    <div className="alert alert-info">
                        <p>
                            The Sourcegraph browser extension adds hover tooltips to code views on code hosts such as
                            GitHub, GitLab, Bitbucket Server and Phabricator.
                        </p>
                        <p>
                            You must grant permissions to enable Sourcegraph on <strong>{currentTabStatus.host}</strong>
                            .
                        </p>
                        <button
                            type="button"
                            className="btn btn-light request-permissions__test"
                            onClick={buildRequestPermissionsHandler(currentTabStatus, requestPermissions)}
                        >
                            Grant permissions
                        </button>
                    </div>
                </div>
            )}
        <div className="options-menu__section">
            <p>
                Learn more about privacy concerns, troubleshooting and extension features{' '}
                <a href="https://docs.sourcegraph.com/integration/browser_extension" target="blank">
                    here
                </a>
                .
            </p>
            <p>
                Search open source software at{' '}
                <a href="https://sourcegraph.com/search" target="blank">
                    sourcegraph.com/search
                </a>
                .
            </p>
        </div>
        {isOptionsMenuExpanded && options && (
            <div className="options-menu__section">
                <label>Configuration</label>
                <div>
                    {options.map(
                        ({ label, key, value, hidden }) =>
                            !hidden && (
                                <div className="form-check" key={key}>
                                    <label className="form-check-label">
                                        <input
                                            id={key}
                                            onChange={event => onChangeOptionFlag(key, event.target.checked)}
                                            className="form-check-input"
                                            type="checkbox"
                                            checked={value}
                                        />{' '}
                                        {label}
                                    </label>
                                </div>
                            )
                    )}
                </div>
            </div>
        )}
    </div>
)
