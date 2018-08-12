import * as React from 'react'
import { Link } from 'react-router-dom'
import { ExtensionsProps } from '../../context'
import { isErrorLike } from '../../errors'
import { ConfigurationCascadeProps, ConfigurationSubject, ID } from '../../settings'
import { LinkOrSpan } from '../../ui/generic/LinkOrSpan'
import { ConfiguredExtension, isExtensionAdded, isExtensionEnabled } from '../extension'
import {
    ADDED_AND_CAN_ADMINISTER,
    ALL_CAN_ADMINISTER,
    ExtensionConfigureButton,
    ExtensionConfiguredSubjectItemForAdd,
    ExtensionConfiguredSubjectItemForRemove,
} from '../ExtensionConfigureButton'
import { ExtensionEnablementToggle } from '../ExtensionEnablementToggle'

interface Props<S extends ConfigurationSubject, C> extends ConfigurationCascadeProps<S, C>, ExtensionsProps<S, C> {
    node: ConfiguredExtension
    subject: ID
    onDidUpdate: () => void
}

/** Displays an extension as a card. */
export class ExtensionCard<S extends ConfigurationSubject, C> extends React.PureComponent<Props<S, C>> {
    public render(): JSX.Element | null {
        const { node, ...props } = this.props
        return (
            <div className="col-sm-6 col-md-6 col-lg-4 pb-4">
                <div className="extension-card card">
                    <LinkOrSpan
                        to={node.registryExtension && node.registryExtension.url}
                        className="card-body extension-card-body d-flex flex-column"
                    >
                        <h4 className="card-title extension-card-body-title mb-0">
                            {node.manifest && !isErrorLike(node.manifest) && node.manifest.title
                                ? node.manifest.title
                                : node.id}
                        </h4>
                        <div className="extension-card-body-text d-inline-block mt-1">
                            {node.manifest ? (
                                isErrorLike(node.manifest) ? (
                                    <span className="text-danger small" title={node.manifest.message}>
                                        <props.extensions.context.icons.Warning className="icon-inline" /> Invalid
                                        manifest
                                    </span>
                                ) : (
                                    node.manifest.description && (
                                        <span className="text-muted">{node.manifest.description}</span>
                                    )
                                )
                            ) : (
                                <span className="text-warning small">
                                    <props.extensions.context.icons.Warning className="icon-inline" /> No manifest
                                </span>
                            )}
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
                            {props.subject && (
                                <>
                                    {isExtensionAdded(props.configurationCascade.merged, node.id) &&
                                        !isExtensionEnabled(props.configurationCascade.merged, node.id) && (
                                            <li className="nav-item">
                                                <ExtensionConfigureButton
                                                    extension={node}
                                                    onUpdate={props.onDidUpdate}
                                                    header="Remove extension for..."
                                                    itemFilter={ADDED_AND_CAN_ADMINISTER}
                                                    itemComponent={ExtensionConfiguredSubjectItemForRemove}
                                                    buttonClassName="btn-outline-link btn-sm py-0 mr-1"
                                                    caret={false}
                                                    configurationCascade={this.props.configurationCascade}
                                                    extensions={this.props.extensions}
                                                >
                                                    Remove
                                                </ExtensionConfigureButton>
                                            </li>
                                        )}
                                    {isExtensionAdded(props.configurationCascade.merged, node.id) && (
                                        <li className="nav-item">
                                            <ExtensionEnablementToggle
                                                extension={node}
                                                subject={props.subject}
                                                onChange={props.onDidUpdate}
                                                tabIndex={-1}
                                                configurationCascade={this.props.configurationCascade}
                                                extensions={this.props.extensions}
                                            />
                                        </li>
                                    )}
                                    {!isExtensionAdded(props.configurationCascade.merged, node.id) && (
                                        <li className="nav-item">
                                            <ExtensionConfigureButton
                                                extension={node}
                                                onUpdate={props.onDidUpdate}
                                                header="Add extension for..."
                                                itemFilter={ALL_CAN_ADMINISTER}
                                                itemComponent={ExtensionConfiguredSubjectItemForAdd}
                                                buttonClassName="btn-primary btn-sm"
                                                caret={false}
                                                configurationCascade={this.props.configurationCascade}
                                                extensions={this.props.extensions}
                                            >
                                                Add
                                            </ExtensionConfigureButton>
                                        </li>
                                    )}
                                </>
                            )}
                        </ul>
                    </div>
                </div>
            </div>
        )
    }
}
