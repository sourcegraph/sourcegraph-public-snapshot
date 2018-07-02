import CheckmarkIcon from '@sourcegraph/icons/lib/Checkmark'
import CodeArrayIcon from '@sourcegraph/icons/lib/CodeArray'
import CodeTagsIcon from '@sourcegraph/icons/lib/CodeTags'
import PencilIcon from '@sourcegraph/icons/lib/Pencil'
import PuzzleIcon from '@sourcegraph/icons/lib/Puzzle'
import SharingIcon from '@sourcegraph/icons/lib/Sharing'
import * as React from 'react'
import { NavLink, RouteComponentProps } from 'react-router-dom'
import { RegistryExtensionAreaPageProps } from './RegistryExtensionArea'
import { RegistryExtensionConfigureButton } from './RegistryExtensionConfigureButton'
import { RegistryExtensionSourceBadge } from './RegistryExtensionSourceBadge'

interface Props extends RegistryExtensionAreaPageProps, RouteComponentProps<{}> {
    className: string
}

/**
 * Header for the registry extension area.
 */
export const RegistryExtensionHeader: React.SFC<Props> = (props: Props) => (
    <div className={`registry-extension-header area-header ${props.className}`}>
        <div className={`${props.className}-inner`}>
            {props.extension && (
                <>
                    <div className="mb-3">
                        <div className="d-flex align-items-start">
                            <PuzzleIcon className="registry-extension-header__icon mr-3 icon-inline" />
                            <div>
                                <div className="d-flex align-items-center">
                                    <h2 className="mb-0">
                                        {(props.extension.manifest && props.extension.manifest.title) ||
                                            props.extension.extensionID}
                                    </h2>
                                    <RegistryExtensionSourceBadge
                                        extension={props.extension}
                                        showIcon={true}
                                        showText={true}
                                        showRegistryName={true}
                                        className="ml-2 mt-1"
                                    />
                                </div>
                                {props.extension.manifest &&
                                    props.extension.manifest.title && (
                                        <div className="text-muted">{props.extension.extensionID}</div>
                                    )}
                            </div>
                        </div>
                    </div>
                    <div className="area-header__nav">
                        <div className="area-header__nav-links">
                            <NavLink
                                to={props.url}
                                exact={true}
                                className="btn area-header__nav-link"
                                activeClassName="area-header__nav-link--active"
                            >
                                <PuzzleIcon className="icon-inline" /> Extension
                            </NavLink>
                            <NavLink
                                to={`${props.url}/-/usage`}
                                className="btn area-header__nav-link"
                                activeClassName="area-header__nav-link--active"
                            >
                                <SharingIcon className="icon-inline" /> Usage
                            </NavLink>
                            <NavLink
                                to={`${props.url}/-/contributions`}
                                exact={true}
                                className="btn area-header__nav-link"
                                activeClassName="area-header__nav-link--active"
                            >
                                <CodeArrayIcon className="icon-inline" /> Contributions
                            </NavLink>
                            <NavLink
                                to={`${props.url}/-/manifest`}
                                exact={true}
                                className="btn area-header__nav-link"
                                activeClassName="area-header__nav-link--active"
                            >
                                <CodeTagsIcon className="icon-inline" /> Manifest
                            </NavLink>
                            {props.extension.viewerCanAdminister && (
                                <NavLink
                                    to={`${props.url}/-/edit`}
                                    className="btn area-header__nav-link"
                                    activeClassName="area-header__nav-link--active"
                                >
                                    <PencilIcon className="icon-inline" /> Edit
                                </NavLink>
                            )}
                        </div>
                        {props.extension.viewerCanConfigure &&
                            props.authenticatedUser && (
                                <div className="area-header__nav-actions">
                                    {props.extension.viewerHasEnabled && (
                                        <small className="text-success mr-1">
                                            <strong>
                                                <CheckmarkIcon className="icon-inline" /> Enabled
                                            </strong>
                                        </small>
                                    )}
                                    <RegistryExtensionConfigureButton
                                        extensionGQLID={props.extension.id}
                                        subject={props.authenticatedUser.id}
                                        viewerCanConfigure={props.extension.viewerCanConfigure}
                                        isEnabled={props.extension.viewerHasEnabled}
                                        onDidUpdate={props.onDidUpdateExtension}
                                    />
                                </div>
                            )}
                    </div>
                </>
            )}
        </div>
    </div>
)
