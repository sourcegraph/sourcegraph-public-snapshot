import GearIcon from '@sourcegraph/icons/lib/Gear'
import * as React from 'react'
import { Link } from 'react-router-dom'
import { gql } from '../backend/graphql'
import * as GQL from '../backend/graphqlschema'
import { FilteredConnection } from '../components/FilteredConnection'
import { RegistryExtensionConfigureButton } from './RegistryExtensionConfigureButton'

export const configuredExtensionFragment = gql`
    fragment ConfiguredExtensionFields on ConfiguredExtension {
        extension {
            id
            extensionID
            manifest {
                title
            }
            url
        }
        subject {
            id
        }
        extensionID
        isEnabled
        viewerCanConfigure
    }
`

export interface ConfiguredExtensionNodeDisplayProps {
    /** Whether to show the action to configure settings for (enable/disable/remove) an extension. */
    showUserActions?: boolean
}

export interface ConfiguredExtensionNodeProps extends ConfiguredExtensionNodeDisplayProps {
    node: GQL.IConfiguredExtension
    settingsURL?: string
    onDidUpdate: () => void
}

export class ConfiguredExtensionNode extends React.PureComponent<ConfiguredExtensionNodeProps> {
    public render(): JSX.Element | null {
        const extensionSpec = this.props.node.extension
            ? { extensionGQLID: this.props.node.extension.id }
            : { extensionID: this.props.node.extensionID }

        return (
            <li className="list-group-item d-block">
                <div className="d-flex w-100 justify-content-between align-items-center">
                    <div className="mr-2 d-flex align-items-center">
                        {this.props.node.extension ? (
                            <>
                                <Link to={this.props.node.extension.url}>
                                    <strong>
                                        {(this.props.node.extension.manifest &&
                                            this.props.node.extension.manifest.title) ||
                                            this.props.node.extensionID}
                                    </strong>
                                </Link>
                                {this.props.node.extension.manifest &&
                                    this.props.node.extension.manifest.title && (
                                        <span className="text-muted ml-1">&mdash; {this.props.node.extensionID}</span>
                                    )}
                            </>
                        ) : (
                            <>
                                <strong>{this.props.node.extensionID}</strong>{' '}
                                <span className="badge badge-danger ml-2">No extension found</span>
                            </>
                        )}
                    </div>
                    {this.props.showUserActions && (
                        <div>
                            {this.props.node.viewerCanConfigure &&
                                this.props.node.subject && (
                                    <RegistryExtensionConfigureButton
                                        {...extensionSpec}
                                        showRemove={!this.props.node.extension || !this.props.node.isEnabled}
                                        subject={this.props.node.subject.id}
                                        viewerCanConfigure={this.props.node.viewerCanConfigure}
                                        isEnabled={this.props.node.isEnabled}
                                        onDidUpdate={this.props.onDidUpdate}
                                        compact={true}
                                    />
                                )}
                            {this.props.settingsURL &&
                                this.props.node.viewerCanConfigure && (
                                    <Link
                                        to={this.props.settingsURL}
                                        className="btn btn-link btn-sm pr-0"
                                        title="Extension settings"
                                    >
                                        <GearIcon className="icon-inline" />
                                    </Link>
                                )}
                        </div>
                    )}
                </div>
            </li>
        )
    }
}

export class FilteredConfiguredExtensionConnection extends FilteredConnection<
    GQL.IConfiguredExtension,
    Pick<ConfiguredExtensionNodeProps, 'onDidUpdate' | 'settingsURL'>
> {}
