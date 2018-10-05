import * as React from 'react'
import { Link } from 'react-router-dom'
import { ExtensionsProps } from '../../context'
import { isErrorLike } from '../../errors'
import { ExtensionManifest } from '../../schema/extension.schema'
import { ConfigurationCascadeProps, ConfigurationSubject, Settings } from '../../settings'
import { LinkOrSpan } from '../../ui/generic/LinkOrSpan'
import { ConfiguredExtension, isExtensionAdded, isExtensionEnabled } from '../extension'
import { ExtensionConfigurationState } from '../ExtensionConfigurationState'
import { ExtensionToggle } from '../ExtensionToggle'

interface Props<S extends ConfigurationSubject, C extends Settings>
    extends ConfigurationCascadeProps<S, C>,
        ExtensionsProps<S, C> {
    node: ConfiguredExtension
    subject: Pick<ConfigurationSubject, 'id' | 'viewerCanAdminister'>
    onDidUpdate: () => void
}

/** Displays an extension as a card. */
export class ExtensionCard<S extends ConfigurationSubject, C extends Settings> extends React.PureComponent<
    Props<S, C>
> {
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
                        className="card-body extension-card-body d-flex flex-column"
                    >
                        <div className="d-flex">
                            {manifest &&
                                manifest.icon &&
                                iconURL &&
                                iconURL.protocol === 'data:' &&
                                /^data:image\/png(;base64)?,/.test(manifest.icon) && (
                                    <img className="extension-card__icon mr-2" src={manifest.icon} />
                                )}
                            <div>
                                <h4 className="card-title extension-card-body-title mb-0">
                                    {manifest && manifest.title ? manifest.title : node.id}
                                </h4>
                                <div className="extension-card-body-text d-inline-block mt-1">
                                    {node.manifest ? (
                                        isErrorLike(node.manifest) ? (
                                            <span className="text-danger small" title={node.manifest.message}>
                                                <props.extensions.context.icons.Warning className="icon-inline" />{' '}
                                                Invalid manifest
                                            </span>
                                        ) : (
                                            node.manifest.description && (
                                                <span className="text-muted">{node.manifest.description}</span>
                                            )
                                        )
                                    ) : (
                                        <span className="text-warning small">
                                            <props.extensions.context.icons.Warning className="icon-inline" /> No
                                            manifest
                                        </span>
                                    )}
                                </div>
                            </div>
                        </div>
                    </LinkOrSpan>
                    <div className="card-footer extension-card-footer py-0 pl-0">
                        <ul className="nav align-items-center">
                            {node.registryExtension &&
                                node.registryExtension.url && (
                                    <li className="nav-item">
                                        <Link to={node.registryExtension.url} className="nav-link px-2" tabIndex={-1}>
                                            Details
                                        </Link>
                                    </li>
                                )}
                            <li className="extension-card-spacer" />
                            {props.subject &&
                                (props.subject.viewerCanAdminister ? (
                                    <ExtensionToggle
                                        extension={node}
                                        configurationCascade={this.props.configurationCascade}
                                        onUpdate={props.onDidUpdate}
                                        className="btn-sm btn-secondary"
                                        extensions={this.props.extensions}
                                    />
                                ) : (
                                    <li className="nav-item">
                                        <ExtensionConfigurationState
                                            isAdded={isExtensionAdded(props.configurationCascade.merged, node.id)}
                                            isEnabled={isExtensionEnabled(props.configurationCascade.merged, node.id)}
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
