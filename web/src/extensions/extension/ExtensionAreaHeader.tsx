import React, { useState, useCallback, useMemo, useEffect } from 'react'
import { NavLink, RouteComponentProps } from 'react-router-dom'
import { isExtensionEnabled } from '../../../../shared/src/extensions/extension'
import { ExtensionManifest } from '../../../../shared/src/schema/extensionSchema'
import { isErrorLike } from '../../../../shared/src/util/errors'
import { NavItemWithIconDescriptor } from '../../util/contributions'
import { ExtensionToggle } from '../ExtensionToggle'
import { ExtensionAreaRouteContext } from './ExtensionArea'
import { WorkInProgressBadge } from './WorkInProgressBadge'
import { isEncodedImage } from '../../../../shared/src/util/icon'
import { useTimeoutManager } from '../../../../shared/src/util/useTimeoutManager'
import classNames from 'classnames'

interface ExtensionAreaHeaderProps extends ExtensionAreaRouteContext, RouteComponentProps<{}> {
    navItems: readonly ExtensionAreaHeaderNavItem[]
    className: string
}

export type ExtensionAreaHeaderContext = Pick<ExtensionAreaHeaderProps, 'extension'>

export interface ExtensionAreaHeaderNavItem extends NavItemWithIconDescriptor<ExtensionAreaHeaderContext> {}

/** ms after which to remove visual feedback */
const FEEDBACK_DELAY = 5000

/**
 * Header for the extension area.
 */
export const ExtensionAreaHeader: React.FunctionComponent<ExtensionAreaHeaderProps> = (
    props: ExtensionAreaHeaderProps
) => {
    const manifest: ExtensionManifest | undefined =
        props.extension.manifest && !isErrorLike(props.extension.manifest) ? props.extension.manifest : undefined

    const iconURL = useMemo(() => {
        let iconURL: URL | undefined
        try {
            if (manifest?.icon) {
                iconURL = new URL(manifest.icon)
            }
        } catch {
            // noop
        }
        return iconURL
    }, [manifest])

    const isWorkInProgress = props.extension.registryExtension?.isWorkInProgress

    const [, name] = props.extension.id.split('/')

    /**
     * When extension enablement state changes, display visual feedback for $delay seconds.
     * Clear the timeout when the component unmounts or the extension is toggled again.
     */
    const [change, setChange] = useState<'enabled' | 'disabled' | null>(null)
    const feedbackTimeoutManager = useTimeoutManager()

    const onToggleChange = React.useCallback(
        (enabled: boolean): void => {
            setChange(enabled ? 'enabled' : 'disabled')
            feedbackTimeoutManager.setTimeout(() => setChange(null), FEEDBACK_DELAY)
        },
        [feedbackTimeoutManager]
    )

    /**
     * Display a CTA on hover over the toggle only when the user is unauthenticated
     */
    const [showCta, setShowCta] = useState(false)
    const ctaTimeoutManager = useTimeoutManager()

    const onHover = useCallback(() => {
        if (!props.authenticatedUser && !showCta) {
            setShowCta(true)
            ctaTimeoutManager.setTimeout(() => setShowCta(false), FEEDBACK_DELAY * 2)
        }
    }, [ctaTimeoutManager, showCta, props.authenticatedUser])

    return (
        <div className={`extension-area-header ${props.className || ''}`}>
            <div className="container ">
                {props.extension && (
                    <>
                        <div className="extension-area-header__wrapper">
                            <div className="mb-3">
                                <div className="d-flex align-items-start">
                                    {manifest?.icon &&
                                        iconURL &&
                                        iconURL.protocol === 'data:' &&
                                        isEncodedImage(manifest.icon) && (
                                            <img className="extension-area-header__icon mr-2" src={manifest.icon} />
                                        )}
                                    <div>
                                        <h1 className="d-flex align-items-center mb-0 font-weight-normal">{name}</h1>
                                        {manifest && (manifest.description || isWorkInProgress) && (
                                            <p className="mt-1 mb-0">
                                                {isWorkInProgress && (
                                                    <WorkInProgressBadge
                                                        viewerCanAdminister={
                                                            !!props.extension.registryExtension &&
                                                            props.extension.registryExtension.viewerCanAdminister
                                                        }
                                                    />
                                                )}
                                                {manifest.description}
                                            </p>
                                        )}
                                    </div>
                                </div>
                            </div>
                            <div className="d-flex flex-column align-items-center justify-content-end mt-3 mb-2 position-relative">
                                {change && (
                                    <div
                                        className={classNames('alert px-2 py-1 mb-0 extension-area-header__alert', {
                                            'alert-secondary': change === 'disabled',
                                            'alert-success': change === 'enabled',
                                        })}
                                    >
                                        <span className="font-weight-semibold">{name}</span> is {change}
                                    </div>
                                )}
                                {showCta && (
                                    <div className="alert alert-secondary mb-0 px-2 py-1 extension-area-header__alert">
                                        Register now! TODO
                                    </div>
                                )}

                                <ExtensionToggle
                                    enabled={isExtensionEnabled(props.settingsCascade.final, props.extension.id)}
                                    extensionID={props.extension.id}
                                    settingsCascade={props.settingsCascade}
                                    platformContext={props.platformContext}
                                    onToggleChange={onToggleChange}
                                    big={true}
                                    onHover={onHover}
                                    userCannotToggle={!props.authenticatedUser}
                                />
                            </div>
                        </div>
                        <div className="mt-3">
                            <ul className="nav nav-tabs border-bottom-0">
                                {props.navItems.map(
                                    ({ to, label, exact, icon: Icon, condition = () => true }) =>
                                        condition(props) && (
                                            <li key={label} className="nav-item">
                                                <NavLink
                                                    to={props.url + to}
                                                    className="nav-link"
                                                    activeClassName="active"
                                                    exact={exact}
                                                >
                                                    {Icon && <Icon className="icon-inline" />} {label}
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

{
    /*
                                {!props.authenticatedUser && (
                                    <div className="d-flex align-items-center">
                                        <Link to="/sign-in" className="btn btn-primary mr-2">
                                            Sign in to{' '}
                                            {isExtensionEnabled(props.settingsCascade.final, props.extension.id)
                                                ? 'configure'
                                                : 'enable'}
                                        </Link>
                                        <small className="text-muted">
                                            An account is required to{' '}
                                            {isExtensionEnabled(props.settingsCascade.final, props.extension.id)
                                                ? ''
                                                : 'enable and'}{' '}
                                            configure extensions.
                                        </small>
                                    </div>
                                )} */
}
