import DeleteIcon from '@sourcegraph/icons/lib/Delete'
import PencilIcon from '@sourcegraph/icons/lib/Pencil'
import ShareIcon from '@sourcegraph/icons/lib/Share'
import UserIcon from '@sourcegraph/icons/lib/User'
import { upperFirst } from 'lodash'
import * as React from 'react'
import { Link } from 'react-router-dom'
import { Subject, Subscription } from 'rxjs'
import { catchError, map, mapTo, startWith, switchMap, tap } from 'rxjs/operators'
import { Timestamp } from '../components/time/Timestamp'
import { asError, ErrorLike, isErrorLike } from '../util/errors'
import { pluralize } from '../util/strings'
import { deleteRegistryExtensionWithConfirmation } from './backend'
import { RegistryExtensionConfigureButton } from './RegistryExtensionConfigureButton'
import { RegistryExtensionIDLink } from './RegistryExtensionIDLink'
import { RegistryExtensionSourceBadge } from './RegistryExtensionSourceBadge'
import { RegistryExtensionNodeProps } from './RegistryExtensionsPage'

interface RegistryExtensionNodeState {
    /** Undefined means in progress, null means done or not started. */
    deletionOrError?: null | ErrorLike
}

/** Displays an extension in a row. */
export class RegistryExtensionNodeRow extends React.PureComponent<
    RegistryExtensionNodeProps,
    RegistryExtensionNodeState
> {
    public state: RegistryExtensionNodeState = {
        deletionOrError: null,
    }

    private deletes = new Subject<void>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        this.subscriptions.add(
            this.deletes
                .pipe(
                    switchMap(() =>
                        deleteRegistryExtensionWithConfirmation(this.props.node.id).pipe(
                            mapTo(null),
                            catchError(error => [asError(error)]),
                            map(c => ({ deletionOrError: c })),
                            tap(() => {
                                if (this.props.onDidUpdate) {
                                    this.props.onDidUpdate()
                                }
                            }),
                            startWith<Pick<RegistryExtensionNodeState, 'deletionOrError'>>({
                                deletionOrError: undefined,
                            })
                        )
                    )
                )
                .subscribe(stateUpdate => this.setState(stateUpdate), error => console.error(error))
        )
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        const loading = this.state.deletionOrError === undefined

        const extensionIDLink = (
            <RegistryExtensionIDLink
                extension={this.props.node}
                showExtensionID={this.props.showExtensionID}
                showRegistryMuted={this.props.showSource && this.props.node.isLocal}
            />
        )

        return (
            <li className="registry-extension-node-row list-group-item d-block">
                <div className="d-flex w-100 justify-content-between align-items-center">
                    <div className="mr-2">
                        <strong>
                            {(this.props.node.manifest &&
                                this.props.node.manifest.title && (
                                    <Link to={this.props.node.url}>{this.props.node.manifest.title}</Link>
                                )) ||
                                extensionIDLink}
                        </strong>
                        {this.props.node.manifest && this.props.node.manifest.title && <> &mdash; {extensionIDLink}</>}
                        <small className="text-muted">
                            {this.props.showSource && (
                                <>
                                    {' '}
                                    - <RegistryExtensionSourceBadge extension={this.props.node} showText={true} />
                                </>
                            )}
                            {this.props.showTimestamp &&
                                this.props.node.updatedAt &&
                                (this.props.showSource ? ', ' : <> - </>)}
                            {this.props.showTimestamp &&
                                this.props.node.updatedAt && (
                                    <>
                                        updated <Timestamp date={this.props.node.updatedAt} />{' '}
                                    </>
                                )}
                        </small>
                    </div>
                    <div className="d-flex align-items-center">
                        {this.props.node.viewerCanAdminister && (
                            <Link
                                to={`${this.props.node.url}/-/edit`}
                                className="btn btn-link btn-sm"
                                title="Edit extension"
                            >
                                <PencilIcon className="icon-inline" />
                            </Link>
                        )}
                        <Link
                            to={this.props.node.url}
                            className="btn btn-sm d-flex align-items-center"
                            title={`${this.props.node.users.totalCount} ${pluralize(
                                'user',
                                this.props.node.users.totalCount
                            )}`}
                        >
                            <span className="registry-extension-node__users">{this.props.node.users.totalCount}</span>{' '}
                            <UserIcon className="icon-inline registry-extension-node__users-icon" />
                        </Link>
                        {!this.props.node.isLocal &&
                            this.props.node.remoteURL &&
                            this.props.node.registryName && (
                                <Link
                                    to={this.props.node.remoteURL}
                                    className="btn btn-link text-info btn-sm py-0"
                                    title={`View extension on ${this.props.node.registryName}`}
                                >
                                    <ShareIcon className="icon-inline" />
                                </Link>
                            )}
                        {this.props.node.viewerCanConfigure &&
                            this.props.authenticatedUserID &&
                            this.props.showUserActions && (
                                <RegistryExtensionConfigureButton
                                    className="ml-1"
                                    extensionGQLID={this.props.node.id}
                                    subject={this.props.authenticatedUserID}
                                    viewerCanConfigure={this.props.node.viewerCanConfigure}
                                    isEnabled={this.props.node.viewerHasEnabled}
                                    disabled={loading}
                                    onDidUpdate={this.props.onDidUpdate}
                                    compact={true}
                                />
                            )}
                        {this.props.showDeleteAction &&
                            this.props.node.viewerCanAdminister && (
                                <button
                                    className="btn btn-outline-danger btn-sm py-0 ml-1"
                                    onClick={this.deleteExtension}
                                    disabled={loading}
                                    title="Delete extension"
                                >
                                    <DeleteIcon className="icon-inline" />
                                </button>
                            )}
                    </div>
                </div>
                {isErrorLike(this.state.deletionOrError) && (
                    <div className="alert alert-danger mt-2">
                        Error: {upperFirst(this.state.deletionOrError.message)}
                    </div>
                )}
            </li>
        )
    }

    private deleteExtension = () => this.deletes.next()
}
