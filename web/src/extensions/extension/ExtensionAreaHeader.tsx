import * as React from 'react'
import { Link, NavLink, RouteComponentProps } from 'react-router-dom'
import PuzzleOutlineIcon from 'mdi-react/PuzzleOutlineIcon'
import { isExtensionEnabled } from '../../../../shared/src/extensions/extension'
import { ExtensionManifest } from '../../../../shared/src/schema/extensionSchema'
import { isErrorLike } from '../../../../shared/src/util/errors'
import { NavItemWithIconDescriptor } from '../../util/contributions'
import { ExtensionToggle } from '../ExtensionToggle'
import { isExtensionAdded } from './extension'
import { ExtensionAreaRouteContext } from './ExtensionArea'
import { ExtensionConfigurationState } from './ExtensionConfigurationState'
import { isEncodedImage } from '../../../../shared/src/util/icon'
import { PageHeader } from '../../components/PageHeader'
import { upperFirst } from 'lodash'
import { PlatformContextProps } from '../../../../shared/src/platform/context'
import { Omit } from 'utility-types'

export interface ExtensionAreaHeaderProps
    extends Omit<ExtensionAreaRouteContext, 'platformContext'>,
        PlatformContextProps<'updateSettings'>,
        RouteComponentProps<{}> {
    navItems: readonly ExtensionAreaHeaderNavItem[]
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
    } catch {
        // noop
    }

    const isWorkInProgress = props.extension.registryExtension?.isWorkInProgress
    const title = manifest?.name
        ? upperFirst(manifest.name)
        : props.extension.registryExtension?.extensionIDWithoutRegistry ?? props.extension.id
    const icon =
        manifest?.icon && iconURL && iconURL.protocol === 'data:' && isEncodedImage(manifest.icon) ? (
            <img className="extension-area-header__icon" src={manifest.icon} />
        ) : (
            <PuzzleOutlineIcon widths={32} className="extension-area-header__icon" />
        )
    const actions = (
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
                        {isExtensionEnabled(props.settingsCascade.final, props.extension.id) ? 'configure' : 'enable'}
                    </Link>
                    <small className="text-muted">
                        An account is required to{' '}
                        {isExtensionEnabled(props.settingsCascade.final, props.extension.id) ? '' : 'enable and'}{' '}
                        configure extensions.
                    </small>
                </div>
            )}
        </div>
    )
    const label = manifest?.description ?? ''
    const badgeLabel = isWorkInProgress ? 'WIP' : undefined
    const badgeTooltip = badgeLabel
        ? props.extension.registryExtension?.viewerCanAdminister
            ? 'Remove "WIP" from the title when this extension is ready for use.'
            : 'Work in progress (not ready for use)'
        : ''
    return (
        <div className="extension-area-header border-bottom">
            {props.extension && (
                <>
                    <PageHeader
                        title={title}
                        icon={icon}
                        actions={actions}
                        breadcrumbs={[
                            { key: 'extensions', element: <Link to="/extensions">Extensions</Link> },
                            { key: 'extensionId', element: props.extension.id },
                        ]}
                        badge={
                            badgeLabel && {
                                label: badgeLabel,
                                tooltip: badgeTooltip,
                            }
                        }
                        label={label}
                    />
                    <div className="container">
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
                    </div>
                </>
            )}
        </div>
    )
}
