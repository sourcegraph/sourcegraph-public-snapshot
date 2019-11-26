import ChevronRightIcon from 'mdi-react/ChevronRightIcon'
import * as React from 'react'
import { Link, NavLink, RouteComponentProps } from 'react-router-dom'
import { Path } from '../../../../shared/src/components/Path'
import { isExtensionEnabled } from '../../../../shared/src/extensions/extension'
import { ExtensionManifest } from '../../../../shared/src/schema/extensionSchema'
import { isErrorLike } from '../../../../shared/src/util/errors'
import { NavItemWithIconDescriptor } from '../../util/contributions'
import { ExtensionToggle } from '../ExtensionToggle'
import { isExtensionAdded } from './extension'
import { ExtensionAreaRouteContext } from './ExtensionArea'
import { ExtensionConfigurationState } from './ExtensionConfigurationState'
import { WorkInProgressBadge } from './WorkInProgressBadge'

interface ExtensionAreaHeaderProps extends ExtensionAreaRouteContext, RouteComponentProps<{}> {
    navItems: readonly ExtensionAreaHeaderNavItem[]
    className: string
}

export type ExtensionAreaHeaderContext = Pick<ExtensionAreaHeaderProps, 'extension'>

export interface ExtensionAreaHeaderNavItem extends NavItemWithIconDescriptor<ExtensionAreaHeaderContext> {}

/**
 * Header for the extension area.
 */
export const ExtensionAreaHeader: React.FunctionComponent<ExtensionAreaHeaderProps> = (
    props: ExtensionAreaHeaderProps
) => {
    const manifest: ExtensionManifest | undefined =
        props.extension.manifest && !isErrorLike(props.extension.manifest) ? props.extension.manifest : undefined
    let iconURL: URL | undefined
    try {
        if (manifest?.icon) {
            iconURL = new URL(manifest.icon)
        }
    } catch (e) {
        // noop
    }

    const isWorkInProgress = props.extension.registryExtension && props.extension.registryExtension.isWorkInProgress

    return (
        <div className={`extension-area-header ${props.className || ''}`}>
            <div className="container">
                {props.extension && (
                    <>
                        <div className="mb-3">
                            <div className="d-flex align-items-start">
                                {manifest?.icon &&
                                    iconURL &&
                                    iconURL.protocol === 'data:' &&
                                    /^data:image\/png(;base64)?,/.test(manifest.icon) && (
                                        <img className="extension-area-header__icon mr-2" src={manifest.icon} />
                                    )}
                                <div>
                                    <h2 className="d-flex align-items-center mb-0 font-weight-normal">
                                        <Link to="/extensions" className="extensions-nav-link">
                                            Extensions
                                        </Link>
                                        <ChevronRightIcon className="icon-inline extension-area-header__icon-chevron" />{' '}
                                        <Path
                                            path={
                                                props.extension.registryExtension
                                                    ? props.extension.registryExtension.extensionIDWithoutRegistry
                                                    : props.extension.id
                                            }
                                        />
                                    </h2>
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
                        <div className="d-flex align-items-center mt-3 mb-2">
                            {props.authenticatedUser && (
                                <ExtensionToggle
                                    extension={props.extension}
                                    settingsCascade={props.settingsCascade}
                                    platformContext={props.platformContext}
                                    className="mr-2"
                                />
                            )}
                            <ExtensionConfigurationState
                                className="mr-2"
                                isAdded={isExtensionAdded(props.settingsCascade.final, props.extension.id)}
                                isEnabled={isExtensionEnabled(props.settingsCascade.final, props.extension.id)}
                            />
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
                            )}
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
