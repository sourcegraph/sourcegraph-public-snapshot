import classNames from 'classnames'
import WarningIcon from 'mdi-react/WarningIcon'
import React, { useState, useCallback, useMemo, memo } from 'react'
import { Link } from 'react-router-dom'

import { ConfiguredRegistryExtension, splitExtensionID } from '@sourcegraph/shared/src/extensions/extension'
import * as GQL from '@sourcegraph/shared/src/graphql/schema'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import {
    ExtensionHeaderColor,
    ExtensionManifest,
    EXTENSION_HEADER_COLORS,
} from '@sourcegraph/shared/src/schema/extensionSchema'
import { SettingsCascadeProps, SettingsSubject } from '@sourcegraph/shared/src/settings/settings'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { isErrorLike } from '@sourcegraph/shared/src/util/errors'
import { isEncodedImage } from '@sourcegraph/shared/src/util/icon'
import { useTimeoutManager } from '@sourcegraph/shared/src/util/useTimeoutManager'

import { AuthenticatedUser } from '../auth'

import { isExtensionAdded } from './extension/extension'
import { ExtensionConfigurationState } from './extension/ExtensionConfigurationState'
import { ExtensionStatusBadge } from './extension/ExtensionStatusBadge'
import headerColorStyles from './ExtensionHeader.module.scss'
import { ExtensionToggle, OptimisticUpdateFailure } from './ExtensionToggle'
import { DefaultExtensionIcon, DefaultSourcegraphExtensionIcon, SourcegraphExtensionIcon } from './icons'

interface Props extends SettingsCascadeProps, PlatformContextProps<'updateSettings'>, ThemeProps {
    node: Pick<
        ConfiguredRegistryExtension<
            Pick<
                GQL.IRegistryExtension,
                'id' | 'extensionIDWithoutRegistry' | 'isWorkInProgress' | 'viewerCanAdminister' | 'url'
            >
        >,
        'id' | 'manifest' | 'registryExtension'
    >
    subject: Pick<GQL.SettingsSubject, 'id' | 'viewerCanAdminister'>
    viewerSubject: SettingsSubject | undefined
    siteSubject: SettingsSubject | undefined
    enabled: boolean
    /** Displayed to site admins to toggle enablement for all users. */
    enabledForAllUsers: boolean
    settingsURL: string | null | undefined
    /** The currently authenticated user. */
    authenticatedUser: AuthenticatedUser | null

    /** Whether this is a featured extension. */
    featured?: boolean
}

/** ms after which to remove visual feedback */
const FEEDBACK_DELAY = 5000

/** Displays an extension as a card. */
export const ExtensionCard = memo<Props>(function ExtensionCard({
    node: extension,
    settingsCascade,
    platformContext,
    subject,
    enabled,
    enabledForAllUsers,
    settingsURL,
    isLightTheme,
    viewerSubject,
    siteSubject,
    authenticatedUser,
    featured,
}) {
    const manifest: ExtensionManifest | undefined =
        extension.manifest && !isErrorLike(extension.manifest) ? extension.manifest : undefined

    const icon = React.useMemo(() => {
        let url: string | undefined

        if (isLightTheme) {
            if (manifest?.icon && isEncodedImage(manifest.icon)) {
                url = manifest.icon
            }
        } else if (manifest?.iconDark && isEncodedImage(manifest.iconDark)) {
            url = manifest.iconDark
        } else if (manifest?.icon && isEncodedImage(manifest.icon)) {
            // fallback: show default icon on dark theme if dark icon isn't specified
            url = manifest.icon
        }

        return url
    }, [manifest?.icon, manifest?.iconDark, isLightTheme])

    const { name, publisher, isSourcegraphExtension } = useMemo(
        () =>
            splitExtensionID(
                extension.registryExtension ? extension.registryExtension.extensionIDWithoutRegistry : extension.id
            ),
        [extension]
    )

    const actionableErrorMessage = (error: Error): JSX.Element => {
        let errorMessage

        if (error.message.startsWith('invalid settings') && settingsURL) {
            errorMessage = (
                <>
                    Could not enable / disable {name}. Edit your <Link to={settingsURL}>user settings</Link> to fix this
                    error. <br />
                    <br /> ({error.message})
                </>
            )
        } else {
            errorMessage = <>{error.message}</>
        }

        return errorMessage
    }

    /**
     * When extension enablement state changes, display visual feedback for $delay seconds.
     * Clear the timeout when the component unmounts or the extension is toggled again.
     */
    const [change, setChange] = useState<'enabled' | 'disabled' | null>(null)
    const feedbackManager = useTimeoutManager()

    const [optimisticFailure, setOptimisticFailure] = useState<OptimisticUpdateFailure<boolean> | null>(null)
    const optimisticFailureManager = useTimeoutManager()

    const onToggleChange = useCallback(
        (enabled: boolean): void => {
            setChange(enabled ? 'enabled' : 'disabled')
            // Common: clear possible error, queue timeout to clear change feedback
            setOptimisticFailure(null)
            optimisticFailureManager.cancelTimeout()
            feedbackManager.setTimeout(() => {
                setChange(null)
            }, FEEDBACK_DELAY)
        },
        [feedbackManager, optimisticFailureManager]
    )

    /**
     * When an optimistic update results in an error, we want to show different
     * feedback and cancel current feedback/animations
     */
    const onToggleError = useCallback(
        (optimisticUpdateFailure: OptimisticUpdateFailure<boolean>) => {
            // Cancel feedback timeout
            feedbackManager.cancelTimeout()
            // Revert state
            setChange(null)
            // Set error state and timeout to clear
            setOptimisticFailure(optimisticUpdateFailure)
            optimisticFailureManager.setTimeout(() => setOptimisticFailure(null), FEEDBACK_DELAY)
        },
        [feedbackManager, optimisticFailureManager]
    )

    const renderUserToggleText = useCallback(
        (enabled: boolean) => (
            <span className="text-muted">
                {enabled ? 'Enabled' : 'Disabled'}
                {authenticatedUser?.siteAdmin && ' for me'}
            </span>
        ),
        [authenticatedUser?.siteAdmin]
    )

    const renderAdminExtensionToggleText = useCallback(
        (enabled: boolean) => <span className="text-muted">{enabled ? 'Enabled' : 'Disabled'} by default for all</span>,
        []
    )

    // Determine header color classname (either defined in manifest, or pseudorandom).
    const headerColorClassName = useMemo(() => {
        if (manifest?.headerColor && EXTENSION_HEADER_COLORS.has(manifest?.headerColor)) {
            return manifest.headerColor
        }

        return headerColorFromExtensionID(extension.id)
    }, [manifest?.headerColor, extension.id])

    const iconClassName = classNames('extension-card__icon', featured && 'extension-card__icon--featured')

    return (
        <div
            className={classNames('extension-card card position-relative flex-1', {
                'p-0 m-0 extension-card--enabled': change === 'enabled',
            })}
        >
            <div className="card-body p-0 extension-card__body d-flex flex-column position-relative">
                {/* Section 1: Icon w/ background */}
                <div
                    className={classNames(
                        'extension-card__background-section d-flex align-items-center',
                        headerColorStyles[headerColorClassName],
                        featured && 'extension-card__background-section--featured'
                    )}
                >
                    {icon ? (
                        <img className={iconClassName} src={icon} alt="" />
                    ) : isSourcegraphExtension ? (
                        <DefaultSourcegraphExtensionIcon className={iconClassName} />
                    ) : (
                        <DefaultExtensionIcon className={iconClassName} />
                    )}
                    {extension.registryExtension?.isWorkInProgress && (
                        <ExtensionStatusBadge
                            viewerCanAdminister={extension.registryExtension.viewerCanAdminister}
                            className="extension-card__badge"
                        />
                    )}
                </div>
                {/* Section 2: Extension details. This should be the section that grows to fill remaining space. */}
                <div className="extension-card__details-section w-100 flex-grow-1">
                    <div className="mb-2">
                        <h3 className="mb-0 mr-1 text-truncate flex-1">
                            <Link to={`/extensions/${extension.id}`}>{name}</Link>
                        </h3>
                        <span>
                            by {publisher}
                            {isSourcegraphExtension && (
                                <SourcegraphExtensionIcon className="icon-inline extension-card__logo" />
                            )}
                        </span>
                    </div>
                    <div
                        className={classNames(
                            'mt-3 extension-card__description',
                            featured && 'extension-card__description--featured'
                        )}
                    >
                        {extension.manifest ? (
                            isErrorLike(extension.manifest) ? (
                                <span className="text-danger small" title={extension.manifest.message}>
                                    <WarningIcon className="icon-inline" /> Invalid manifest
                                </span>
                            ) : (
                                extension.manifest.description && (
                                    <span className="">{extension.manifest.description}</span>
                                )
                            )
                        ) : (
                            <span className="text-warning small">
                                <WarningIcon className="icon-inline" /> No manifest
                            </span>
                        )}
                    </div>
                </div>
                {/* Item 3: Toggle(s) */}
                <div className="extension-card__toggles-section d-flex flex-column align-items-end mt-1">
                    <div className="px-1">
                        {/* User toggle */}
                        {subject &&
                            (subject.viewerCanAdminister && viewerSubject ? (
                                <ExtensionToggle
                                    extensionID={extension.id}
                                    enabled={enabled}
                                    settingsCascade={settingsCascade}
                                    platformContext={platformContext}
                                    onToggleChange={onToggleChange}
                                    onToggleError={onToggleError}
                                    subject={viewerSubject}
                                    className="mx-2"
                                    renderText={renderUserToggleText}
                                />
                            ) : (
                                <ExtensionConfigurationState
                                    isAdded={isExtensionAdded(settingsCascade.final, extension.id)}
                                    isEnabled={enabled}
                                    enabledIconOnly={true}
                                    className="small"
                                />
                            ))}
                    </div>
                    {/* Site admin toggle */}
                    {authenticatedUser?.siteAdmin && siteSubject && (
                        <div className="px-1 mt-2">
                            <ExtensionToggle
                                extensionID={extension.id}
                                enabled={enabledForAllUsers}
                                settingsCascade={settingsCascade}
                                platformContext={platformContext}
                                onToggleChange={onToggleChange}
                                onToggleError={onToggleError}
                                subject={siteSubject}
                                className="mx-2"
                                renderText={renderAdminExtensionToggleText}
                            />
                        </div>
                    )}
                </div>
            </div>

            {/* Visual feedback: alert when optimistic update fails */}
            {optimisticFailure && (
                <div className="alert alert-danger px-2 py-1 extension-card__alert">
                    <span className="font-weight-medium">Error:</span> {actionableErrorMessage(optimisticFailure.error)}
                </div>
            )}
        </div>
    )
},
areEqual)

/**
 * Custom compareFunction for ExtensionCard.
 *
 * Rendering all ExtensionCards on settings changes significantly affects performance,
 * so only render when necessary.
 */
function areEqual(oldProps: Props, newProps: Props): boolean {
    if (newProps.authenticatedUser?.siteAdmin) {
        // Also check if the extension is enabled for all users if the user is a site admin
        return (
            oldProps.enabledForAllUsers === newProps.enabledForAllUsers &&
            oldProps.enabled === newProps.enabled &&
            oldProps.isLightTheme === newProps.isLightTheme
        )
    }
    return oldProps.enabled === newProps.enabled && oldProps.isLightTheme === newProps.isLightTheme
}

const extensionHeaderColors = [...EXTENSION_HEADER_COLORS]

/**
 * Pseudorandom header color for extensions that haven't set `headerColor` in their manifest.
 */
function headerColorFromExtensionID(extensionID: string): ExtensionHeaderColor {
    const dividend = [...extensionID].reduce((sum, character) => (sum += character.charCodeAt(0)), 0)
    const divisor = extensionHeaderColors.length
    const remainder = dividend % divisor

    return extensionHeaderColors[remainder] || 'blue' // Fallback, but should never reach this state
}
