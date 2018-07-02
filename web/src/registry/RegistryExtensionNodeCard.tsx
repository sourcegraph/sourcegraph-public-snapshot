import * as React from 'react'
import { Link } from 'react-router-dom'
import { Timestamp } from '../components/time/Timestamp'
import { pluralize } from '../util/strings'
import { RegistryExtensionConfigureButton } from './RegistryExtensionConfigureButton'
import { RegistryExtensionIDLink } from './RegistryExtensionIDLink'
import { RegistryExtensionNodeProps } from './RegistryExtensionsPage'

/** Displays a registry extension as a card. */
export const RegistryExtensionNodeCard: React.SFC<RegistryExtensionNodeProps> = ({ node, ...props }) => {
    const title = node.manifest && node.manifest.title
    const extensionIDLink = (
        <RegistryExtensionIDLink
            extension={node}
            showExtensionID={props.showExtensionID}
            showRegistryMuted={props.showSource && node.isLocal}
            className="registry-extension-node-card__title-link"
        />
    )
    return (
        <div className="registry-extension-node-card col-sm-6 col-md-6 col-lg-4 pb-3">
            <div className="registry-extension-node-card__card card">
                <div className="card-body registry-extension-node-card__card-body">
                    <h4 className="card-title registry-extension-node-card__title mb-0">
                        {title ? <Link to={node.url}>{title}</Link> : extensionIDLink}
                    </h4>
                    {title && <small>{extensionIDLink}</small>}
                    {props.showTimestamp &&
                        node.updatedAt && (
                            <p className="card-text">
                                <small className="text-muted">
                                    Updated <Timestamp date={node.updatedAt} />
                                </small>
                            </p>
                        )}
                    <div className="d-flex align-items-center justify-content-between">
                        <div className="d-flex align-items-center mr-2">
                            <Link
                                to={node.url}
                                className="btn btn-sm d-flex align-items-center pl-0 py-0 text-muted"
                                title={`${node.users.totalCount} ${pluralize(
                                    'user',
                                    node.users.totalCount
                                )} on this Sourcegraph site`}
                            >
                                {node.users.totalCount} {pluralize('user', node.users.totalCount)}
                            </Link>
                            {node.viewerCanAdminister &&
                                props.showEditAction && (
                                    <Link
                                        to={`${node.url}/-/edit`}
                                        className="btn btn-link btn-sm py-0 pl-0"
                                        title="Edit extension"
                                    >
                                        Edit
                                    </Link>
                                )}
                        </div>
                        <div>
                            {props.subject &&
                                props.subjectIsViewer &&
                                props.subject.viewerCanAdminister &&
                                props.showUserActions && (
                                    <RegistryExtensionConfigureButton
                                        extensionGQLID={node.id}
                                        subject={props.subject.id}
                                        viewerCanConfigure={props.subject.viewerCanAdminister}
                                        isEnabled={node.viewerHasEnabled}
                                        onDidUpdate={props.onDidUpdate}
                                        compact={true}
                                    />
                                )}
                        </div>
                    </div>
                </div>
            </div>
        </div>
    )
}
