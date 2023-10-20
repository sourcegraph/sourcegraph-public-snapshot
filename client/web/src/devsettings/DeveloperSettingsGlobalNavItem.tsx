import { type FC, useMemo } from 'react'

import { mdiAlertOctagon, mdiRefresh } from '@mdi/js'
import classNames from 'classnames'
import { useLocation } from 'react-router-dom'

import { pluralize } from '@sourcegraph/common'
import { Shortcut } from '@sourcegraph/shared/src/react-shortcuts'
import { Button, type ButtonProps, Icon, Tooltip } from '@sourcegraph/wildcard'

import { getFeatureFlagOverrides } from '../featureFlags/lib/feature-flag-local-overrides'
import {
    formatUrlOverrideFeatureFlags,
    parseUrlOverrideFeatureFlags,
} from '../featureFlags/lib/parseUrlOverrideFeatureFlags'
import { useFeatureFlagOverrides } from '../featureFlags/useFeatureFlagOverrides'
import { NavAction } from '../nav'
import { toggleDevSettingsDialog, useOverrideCounter } from '../stores'

/**
 * This function adds `feat` query parameters for every enabled and disabled
 * feature flag to the current URL.
 */
function getReloadURL(): string {
    const url = new URL(window.location.toString())
    url.searchParams.delete('feat')
    const overrides = formatUrlOverrideFeatureFlags(getFeatureFlagOverrides()).join(',')
    if (overrides) {
        url.searchParams.set('feat', overrides)
    }
    return url.toString()
}

export const DeveloperSettingsGlobalNavItem: FC<{}> = () => {
    const counter = useOverrideCounter()
    const hasOverrides = counter.featureFlags + counter.temporarySettings > 0
    const showReloadButton = useMighNeedReload()

    return (
        <NavAction>
            <span className="d-flex">
                <Button
                    className={classNames(showReloadButton && 'pr-1')}
                    variant="link"
                    onClick={() => toggleDevSettingsDialog(true)}
                >
                    Developer Settings
                    {hasOverrides && (
                        <Tooltip
                            content={`You have ${counter.featureFlags} local feature flag and ${counter.temporarySettings} temporary settings overrides.`}
                        >
                            <Icon
                                className="ml-1"
                                style={{ color: 'var(--orange)' }}
                                svgPath={mdiAlertOctagon}
                                aria-hidden={true}
                            />
                        </Tooltip>
                    )}
                </Button>
                {showReloadButton && (
                    <>
                        <ReloadButton variant="icon" />
                        <Shortcut
                            held={['Mod']}
                            ordered={['r']}
                            onMatch={() => (window.location.href = getReloadURL())}
                        />
                    </>
                )}
            </span>
        </NavAction>
    )
}

/**
 * Renders a button which reloads the page with the currently overridden feature flags set.
 * Reloading the page when feature flags change is important to ensure that they have been
 * properly processed on the server side.
 */
export const ReloadButton: FC<ButtonProps> = ({ children, ...buttonProps }) => {
    const { featureFlags } = useOverrideCounter()
    const needsReload = useMighNeedReload()
    const tooltipContent = needsReload
        ? featureFlags === 0
            ? 'Reload page to reset feature flags.'
            : `Reload page to apply ${featureFlags} feature ${pluralize('flag', featureFlags)}.`
        : ''

    return (
        <Tooltip content={tooltipContent}>
            <Button {...buttonProps} disabled={!needsReload} onClick={() => (window.location.href = getReloadURL())}>
                <Icon svgPath={mdiRefresh} aria-hidden={true} /> {children}
            </Button>
        </Tooltip>
    )
}

/**
 * This hook compares the feature flags present in the URL against the local
 * overrides and returns `true` if they differ.
 */
function useMighNeedReload(): boolean {
    const featureFlags = useFeatureFlagOverrides()
    const location = useLocation()

    const needsReload = useMemo(() => {
        const urlOverrides = parseUrlOverrideFeatureFlags(location.search)

        if (featureFlags.size === 0 && urlOverrides.size === 0) {
            return false
        }

        return (
            formatUrlOverrideFeatureFlags(urlOverrides).sort().join(',') !==
            formatUrlOverrideFeatureFlags(featureFlags).sort().join(',')
        )
    }, [featureFlags, location.search])

    return needsReload
}
