import WarningIcon from 'mdi-react/WarningIcon'
import * as React from 'react'
import { Link } from 'react-router-dom'
import { LinkOrSpan } from '../../../shared/src/components/LinkOrSpan'
import { Path } from '../../../shared/src/components/Path'
import { ConfiguredRegistryExtension, isExtensionEnabled } from '../../../shared/src/extensions/extension'
import * as GQL from '../../../shared/src/graphql/schema'
import { PlatformContextProps } from '../../../shared/src/platform/context'
import { ExtensionManifest } from '../../../shared/src/schema/extensionSchema'
import { SettingsCascadeProps } from '../../../shared/src/settings/settings'
import { isErrorLike } from '../../../shared/src/util/errors'
import { isExtensionAdded } from './extension/extension'
import { ExtensionConfigurationState } from './extension/ExtensionConfigurationState'
import { WorkInProgressBadge } from './extension/WorkInProgressBadge'
import { ExtensionToggle } from './ExtensionToggle'

interface Props extends SettingsCascadeProps, PlatformContextProps<'updateSettings'> {
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
}

/** Displays an extension as a card. */
export class ExtensionCard extends React.PureComponent<Props> {
    public render(): JSX.Element | null {
        const { node, ...props } = this.props
        const manifest: ExtensionManifest | undefined =
            node.manifest && !isErrorLike(node.manifest) ? node.manifest : undefined
        let iconURL: URL | undefined
        try {
            if (manifest && manifest.icon) {
                iconURL = new URL(manifest.icon)
            }
        } catch (e) {
            // noop
        }

        return (
            <div className="d-flex col-sm-6 col-md-6 col-lg-4 pb-4">
                <div className="extension-card card">
                    <LinkOrSpan
                        to={node.registryExtension && node.registryExtension.url}
                        className="card-body extension-card__body d-flex flex-column"
                    >
                        <div className="d-flex">
                            {manifest &&
                                manifest.icon &&
                                iconURL &&
                                iconURL.protocol === 'data:' &&
                                /^data:image\/png(;base64)?,/.test(manifest.icon) && (
                                    <img className="extension-card__icon mr-2" src={manifest.icon} />
                                )}
                            <div className="text-truncate">
                                <div className="d-flex align-items-center">
                                    <h4 className="card-title extension-card__body-title mb-0 mr-1 text-truncate font-weight-normal">
                                        <Path
                                            path={
                                                node.registryExtension
                                                    ? node.registryExtension.extensionIDWithoutRegistry
                                                    : node.id
                                            }
                                        />
                                    </h4>
                                    {node.registryExtension && node.registryExtension.isWorkInProgress && (
                                        <WorkInProgressBadge
                                            viewerCanAdminister={node.registryExtension.viewerCanAdminister}
                                        />
                                    )}
                                </div>
                                <div className="mt-1">
                                    {node.manifest ? (
                                        isErrorLike(node.manifest) ? (
                                            <span className="text-danger small" title={node.manifest.message}>
                                                <WarningIcon className="icon-inline" /> Invalid manifest
                                            </span>
                                        ) : (
                                            node.manifest.description && (
                                                <div className="text-muted text-truncate">
                                                    {node.manifest.description}
                                                </div>
                                            )
                                        )
                                    ) : (
                                        <span className="text-warning small">
                                            <WarningIcon className="icon-inline" /> No manifest
                                        </span>
                                    )}
                                </div>
                            </div>
                        </div>
                    </LinkOrSpan>
                    <div className="card-footer extension-card__footer py-0 pl-0">
                        <ul className="nav align-items-center">
                            {node.registryExtension && node.registryExtension.url && (
                                <li className="nav-item">
                                    <Link to={node.registryExtension.url} className="nav-link px-2" tabIndex={-1}>
                                        Details
                                    </Link>
                                </li>
                            )}
                            <li className="extension-card__spacer" />
                            {props.subject &&
                                (props.subject.viewerCanAdminister ? (
                                    <ExtensionToggle
                                        extension={node}
                                        settingsCascade={this.props.settingsCascade}
                                        platformContext={this.props.platformContext}
                                    />
                                ) : (
                                    <li className="nav-item">
                                        <ExtensionConfigurationState
                                            isAdded={isExtensionAdded(props.settingsCascade.final, node.id)}
                                            isEnabled={isExtensionEnabled(props.settingsCascade.final, node.id)}
                                        />
                                    </li>
                                ))}
                        </ul>
                    </div>
                </div>
            </div>
        )
    }
}
