import React, { useState, useMemo } from 'react'

import { mdiPuzzleOutline } from '@mdi/js'
import classNames from 'classnames'
import { NavLink, RouteComponentProps } from 'react-router-dom'

import { isErrorLike } from '@sourcegraph/common'
import { isExtensionEnabled, splitExtensionID } from '@sourcegraph/shared/src/extensions/extension'
import { ExtensionManifest } from '@sourcegraph/shared/src/schema/extensionSchema'
import { PageHeader, useTimeoutManager, Alert, Icon, Text } from '@sourcegraph/wildcard'

import { NavItemWithIconDescriptor } from '../../util/contributions'
import { ExtensionToggle } from '../ExtensionToggle'

import { ExtensionAreaRouteContext } from './ExtensionArea'
import { ExtensionStatusBadge } from './ExtensionStatusBadge'

import styles from './ExtensionAreaHeader.module.scss'

interface ExtensionAreaHeaderProps extends ExtensionAreaRouteContext, RouteComponentProps<{}> {
    navItems: readonly ExtensionAreaHeaderNavItem[]
    className?: string
}

export type ExtensionAreaHeaderContext = Pick<ExtensionAreaHeaderProps, 'extension'>

export interface ExtensionAreaHeaderNavItem extends NavItemWithIconDescriptor<ExtensionAreaHeaderContext> {}

/** ms after which to remove visual feedback */
const FEEDBACK_DELAY = 5000

/**
 * Header for the extension area.
 */
export const ExtensionAreaHeader: React.FunctionComponent<React.PropsWithChildren<ExtensionAreaHeaderProps>> = (
    props: ExtensionAreaHeaderProps
) => {
    const manifest: ExtensionManifest | undefined =
        props.extension.manifest && !isErrorLike(props.extension.manifest) ? props.extension.manifest : undefined

    const isWorkInProgress = props.extension.registryExtension?.isWorkInProgress

    const { publisher, name } = splitExtensionID(props.extension.id)

    const isSiteAdmin = props.authenticatedUser?.siteAdmin
    const siteSubject = useMemo(
        () => props.settingsCascade.subjects?.find(settingsSubject => settingsSubject.subject.__typename === 'Site'),
        [props.settingsCascade]
    )

    /**
     * When extension enablement state changes, display visual feedback for $delay seconds.
     * Clear the timeout when the component unmounts or the extension is toggled again.
     */
    const [change, setChange] = useState<'enabled' | 'disabled' | null>(null)
    const feedbackTimeoutManager = useTimeoutManager()

    const onToggleChange = React.useCallback(
        (enabled: boolean): void => {
            // Don't show change alert when the user is a site admin (two toggles)
            if (!isSiteAdmin) {
                setChange(enabled ? 'enabled' : 'disabled')
                feedbackTimeoutManager.setTimeout(() => setChange(null), FEEDBACK_DELAY)
            }
        },
        [feedbackTimeoutManager, isSiteAdmin]
    )

    return (
        <div className={props.className}>
            <div className="container">
                {props.extension && (
                    <>
                        <PageHeader
                            annotation={
                                isWorkInProgress && (
                                    <ExtensionStatusBadge
                                        viewerCanAdminister={
                                            props.extension.registryExtension?.viewerCanAdminister || false
                                        }
                                    />
                                )
                            }
                            path={[
                                { to: '/extensions', icon: mdiPuzzleOutline, ariaLabel: 'Extensions' },
                                { text: publisher },
                                { text: name },
                            ]}
                            description={
                                manifest &&
                                (manifest.description || isWorkInProgress) && (
                                    <Text className="mt-1 mb-0">{manifest.description}</Text>
                                )
                            }
                            actions={
                                <div className={classNames('position-relative', styles.actions)}>
                                    {change && (
                                        <Alert
                                            variant={change === 'enabled' ? 'success' : 'secondary'}
                                            className={classNames('px-2 py-1 mb-0', styles.alert)}
                                        >
                                            <span className="font-weight-medium">{name}</span> is {change}
                                        </Alert>
                                    )}
                                    {/* If site admin, render user toggle and site toggle (both small) */}
                                    {isSiteAdmin && siteSubject?.subject ? (
                                        (() => {
                                            const enabledForMe = isExtensionEnabled(
                                                props.settingsCascade.final,
                                                props.extension.id
                                            )
                                            const enabledForAllUsers = isExtensionEnabled(
                                                siteSubject.settings,
                                                props.extension.id
                                            )

                                            return (
                                                <div className="d-flex flex-column justify-content-center text-muted">
                                                    <div
                                                        className={classNames(
                                                            'd-flex align-items-center mb-2',
                                                            styles.toggleWrapper
                                                        )}
                                                    >
                                                        <span>{enabledForMe ? 'Enabled' : 'Disabled'} for me</span>
                                                        <ExtensionToggle
                                                            className="ml-2 mb-1"
                                                            enabled={enabledForMe}
                                                            extensionID={props.extension.id}
                                                            settingsCascade={props.settingsCascade}
                                                            platformContext={props.platformContext}
                                                            onToggleChange={onToggleChange}
                                                            big={false}
                                                            subject={props.authenticatedUser}
                                                        />
                                                    </div>
                                                    {/* Site admin toggle */}
                                                    <div
                                                        className={classNames(
                                                            'd-flex align-items-center',
                                                            styles.toggleWrapper
                                                        )}
                                                    >
                                                        <span>
                                                            {enabledForAllUsers ? 'Enabled' : 'Not enabled'} for all
                                                            users
                                                        </span>
                                                        <ExtensionToggle
                                                            className="ml-2 mb-1"
                                                            enabled={enabledForAllUsers}
                                                            extensionID={props.extension.id}
                                                            settingsCascade={props.settingsCascade}
                                                            platformContext={props.platformContext}
                                                            onToggleChange={onToggleChange}
                                                            big={false}
                                                            subject={siteSubject.subject}
                                                        />
                                                    </div>
                                                </div>
                                            )
                                        })()
                                    ) : (
                                        <ExtensionToggle
                                            className="mt-md-3"
                                            enabled={isExtensionEnabled(
                                                props.settingsCascade.final,
                                                props.extension.id
                                            )}
                                            extensionID={props.extension.id}
                                            settingsCascade={props.settingsCascade}
                                            platformContext={props.platformContext}
                                            onToggleChange={onToggleChange}
                                            big={true}
                                            subject={props.authenticatedUser}
                                        />
                                    )}
                                </div>
                            }
                        />
                        <div className="mt-4">
                            <ul className="nav nav-tabs">
                                {props.navItems.map(
                                    ({ to, label, exact, icon: ItemIcon, condition = () => true }) =>
                                        condition(props) && (
                                            <li key={label} className="nav-item">
                                                <NavLink
                                                    to={props.url + to}
                                                    className="nav-link"
                                                    activeClassName="active"
                                                    exact={exact}
                                                >
                                                    <span>
                                                        {ItemIcon && <Icon as={ItemIcon} aria-hidden={true} />}{' '}
                                                        <span className="text-content" data-tab-content={label}>
                                                            {label}
                                                        </span>
                                                    </span>
                                                </NavLink>
                                            </li>
                                        )
                                )}
                            </ul>
                        </div>
                    </>
                )}
            </div>
        </div>
    )
}
