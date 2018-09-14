import { isExtensionAdded, isExtensionEnabled } from '@sourcegraph/extensions-client-common/lib/extensions/extension'
import PuzzleIcon from '@sourcegraph/icons/lib/Puzzle'
import * as React from 'react'
import { Link, NavLink, RouteComponentProps } from 'react-router-dom'
import { NavItemWithIconDescriptor } from '../../util/contributions'
import { isErrorLike } from '../../util/errors'
import { ExtensionAreaRouteContext } from './ExtensionArea'
import { ExtensionConfigurationState } from './ExtensionConfigurationState'
import { RegistryExtensionDetailActionButton } from './RegistryExtensionDetailActionButton'

interface ExtensionAreaHeaderProps extends ExtensionAreaRouteContext, RouteComponentProps<{}> {
    navItems: ReadonlyArray<ExtensionAreaHeaderNavItem>
}

export type ExtensionAreaHeaderContext = Pick<ExtensionAreaHeaderProps, 'extension'>

export interface ExtensionAreaHeaderNavItem extends NavItemWithIconDescriptor<ExtensionAreaHeaderContext> {}

/**
 * Header for the extension area.
 */
export const ExtensionAreaHeader: React.SFC<ExtensionAreaHeaderProps> = (props: ExtensionAreaHeaderProps) => (
    <div className="extension-area-header border-bottom simple-area-header pt-4">
        <div className="container">
            {props.extension && (
                <>
                    <div className="mb-3">
                        <div className="d-flex align-items-start">
                            <PuzzleIcon className="extension-area-header__icon mr-3 icon-inline" />
                            <div>
                                <div className="d-flex align-items-center">
                                    <h2 className="mb-0">
                                        {(props.extension.manifest &&
                                            !isErrorLike(props.extension.manifest) &&
                                            props.extension.manifest.title) ||
                                            props.extension.id}
                                    </h2>
                                </div>
                                {props.extension.manifest &&
                                    !isErrorLike(props.extension.manifest) &&
                                    props.extension.manifest.title && (
                                        <div className="text-muted">{props.extension.id}</div>
                                    )}
                                {props.extension.manifest &&
                                    !isErrorLike(props.extension.manifest) &&
                                    props.extension.manifest.description && (
                                        <p className="mt-1 mb-0">{props.extension.manifest.description}</p>
                                    )}
                            </div>
                        </div>
                        <div className="d-flex align-items-center mt-3 mb-2">
                            <ExtensionConfigurationState
                                isAdded={isExtensionAdded(props.configurationCascade.merged, props.extension.id)}
                                isEnabled={isExtensionEnabled(props.configurationCascade.merged, props.extension.id)}
                                enabledIconOnly={!!props.authenticatedUser}
                                className="mr-2"
                            />
                            {props.authenticatedUser && (
                                <RegistryExtensionDetailActionButton
                                    extension={props.extension}
                                    onUpdate={props.onDidUpdateExtension}
                                    nonButtonClassName="d-block"
                                    configurationCascade={props.configurationCascade}
                                    extensions={props.extensions}
                                />
                            )}
                        </div>
                        {!props.authenticatedUser && (
                            <div className="d-flex align-items-center mt-3 mb-2">
                                <Link to="/sign-in" className="btn btn-primary mr-2">
                                    Sign in to{' '}
                                    {isExtensionEnabled(props.configurationCascade.merged, props.extension.id)
                                        ? 'configure'
                                        : 'enable'}
                                </Link>
                                <small className="text-muted">
                                    An account is required to{' '}
                                    {isExtensionEnabled(props.configurationCascade.merged, props.extension.id)
                                        ? ''
                                        : 'enable and'}{' '}
                                    configure extensions.
                                </small>
                            </div>
                        )}
                    </div>
                    <div className="area-header__nav mt-3">
                        <div className="area-header__nav-links">
                            {props.navItems.map(
                                ({ to, label, exact, icon: Icon, condition = () => true }) =>
                                    condition(props) && (
                                        <NavLink
                                            key={label}
                                            to={props.url + to}
                                            className="btn area-header__nav-link"
                                            activeClassName="area-header__nav-link--active"
                                            exact={exact}
                                        >
                                            {Icon && <Icon className="icon-inline" />} {label}
                                        </NavLink>
                                    )
                            )}
                        </div>
                    </div>
                </>
            )}
        </div>
    </div>
)
