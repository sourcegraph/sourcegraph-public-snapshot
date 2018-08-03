import LockIcon from '@sourcegraph/icons/lib/Lock'
import PuzzleIcon from '@sourcegraph/icons/lib/Puzzle'
import * as React from 'react'
import { NavLink, RouteComponentProps } from 'react-router-dom'
import { isErrorLike } from '../../util/errors'
import { ExtensionsProps } from '../ExtensionsClientCommonContext'
import { ExtensionAreaPageProps } from './ExtensionArea'
import { ExtensionConfigurationState } from './ExtensionConfigurationState'
import { RegistryExtensionDetailActionButton } from './RegistryExtensionDetailActionButton'

interface Props extends ExtensionAreaPageProps, ExtensionsProps, RouteComponentProps<{}> {}

/**
 * Header for the extension area.
 */
export const ExtensionAreaHeader: React.SFC<Props> = (props: Props) => (
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
                                            props.extension.extensionID}
                                    </h2>
                                </div>
                                {props.extension.manifest &&
                                    !isErrorLike(props.extension.manifest) &&
                                    props.extension.manifest.title && (
                                        <div className="text-muted">{props.extension.extensionID}</div>
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
                                isAdded={props.extension.isAdded}
                                isEnabled={props.extension.isEnabled}
                                enabledIconOnly={true}
                                className="mr-2"
                            />
                            <RegistryExtensionDetailActionButton
                                extension={props.extension}
                                onUpdate={props.onDidUpdateExtension}
                                nonButtonClassName="d-block"
                                extensions={props.extensions}
                            />
                        </div>
                    </div>
                    <div className="area-header__nav mt-3">
                        <div className="area-header__nav-links">
                            <NavLink
                                to={props.url}
                                exact={true}
                                className="btn area-header__nav-link"
                                activeClassName="area-header__nav-link--active"
                            >
                                Extension
                            </NavLink>
                            {props.extension.registryExtension &&
                                props.extension.registryExtension.viewerCanAdminister && (
                                    <NavLink
                                        to={`${props.url}/-/manage`}
                                        className="btn area-header__nav-link"
                                        activeClassName="area-header__nav-link--active"
                                    >
                                        <LockIcon className="icon-inline" /> Manage
                                    </NavLink>
                                )}
                        </div>
                    </div>
                </>
            )}
        </div>
    </div>
)
