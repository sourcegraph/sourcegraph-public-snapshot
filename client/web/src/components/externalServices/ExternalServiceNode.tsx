import React, { useCallback, useState } from 'react'

import { mdiCircle, mdiCog, mdiConnection, mdiDelete } from '@mdi/js'
import classNames from 'classnames'
import * as H from 'history'

import { Timestamp } from '@sourcegraph/branded/src/components/Timestamp'
import { asError, isErrorLike, pluralize } from '@sourcegraph/common'
import { Button, ButtonProps, Link, LoadingSpinner, Icon, Tooltip, Text, ErrorAlert } from '@sourcegraph/wildcard'

import { ListExternalServiceFields } from '../../graphql-operations'
import { refreshSiteFlags } from '../../site/backend'

import { deleteExternalService, testExternalServiceConnection } from './backend'
import { defaultExternalServices } from './externalServices'

import styles from './ExternalServiceNode.module.scss'


export interface ExternalServiceNodeProps {
    node: ListExternalServiceFields
    onDidUpdate: () => void
    history: H.History
    routingPrefix: string
    afterDeleteRoute: string
    editingDisabled: boolean
}

export const ExternalServiceNode: React.FunctionComponent<React.PropsWithChildren<ExternalServiceNodeProps>> = ({
    node,
    onDidUpdate,
    history,
    routingPrefix,
    afterDeleteRoute,
    editingDisabled,
}) => {
    const [isDeleting, setIsDeleting] = useState<boolean | Error>(false)
    const onDelete = useCallback<React.MouseEventHandler>(async () => {
        if (!window.confirm(`Delete the external service ${node.displayName}?`)) {
            return
        }
        setIsDeleting(true)
        try {
            await deleteExternalService(node.id)
            setIsDeleting(false)
            onDidUpdate()
            // eslint-disable-next-line rxjs/no-ignored-subscription
            refreshSiteFlags().subscribe()
            history.push(afterDeleteRoute)
        } catch (error) {
            setIsDeleting(asError(error))
        }
    }, [afterDeleteRoute, history, node.displayName, node.id, onDidUpdate])

    const [isTestInProgress, setIsTestInProgress] = useState<boolean | Error>(false)
    const onTestConnection = useCallback<React.MouseEventHandler>(async () => {
        setIsTestInProgress(true)
        try {
            await testExternalServiceConnection(node.id)
            setIsTestInProgress(false)
            // FIXME: add on update callback
        } catch (error) {
            setIsTestInProgress(asError(error))
        }
    }, [])

    const IconComponent = defaultExternalServices[node.kind].icon

    return (
        <li
            className={classNames(styles.listNode, 'external-service-node list-group-item')}
            data-test-external-service-name={node.displayName}
        >
            <div className="d-flex align-items-center justify-content-between">
                <div className="align-self-start">
                    {node.lastSyncError === null && (
                        <Tooltip content="All good, no errors!">
                            <Icon
                                svgPath={mdiCircle}
                                aria-label="Code host integration is healthy"
                                className="text-success mr-2"
                            />
                        </Tooltip>
                    )}
                    {node.lastSyncError !== null && (
                        <Tooltip content="Syncing failed, check the error message for details!">
                            <Icon
                                svgPath={mdiCircle}
                                aria-label="Code host integration is unhealthy"
                                className="text-danger mr-2"
                            />
                        </Tooltip>
                    )}
                </div>
                <div className="flex-grow-1">
                    <div>
                        <Icon as={IconComponent} aria-label="Code host logo" className="mr-2" />
                        <strong>
                            {node.displayName}{' '}
                            <small className="text-muted">
                                ({node.repoCount} {pluralize('repository', node.repoCount, 'repositories')})
                            </small>
                        </strong>
                        <br />
                        <Text className="mb-0 text-muted">
                            <small>
                                {node.lastSyncAt === null ? (
                                    <>Never synced.</>
                                ) : (
                                    <>
                                        Last synced <Timestamp date={node.lastSyncAt} />.
                                    </>
                                )}{' '}
                                {node.nextSyncAt !== null && (
                                    <>
                                        Next sync scheduled <Timestamp date={node.nextSyncAt} />.
                                    </>
                                )}
                                {node.nextSyncAt === null && <>No next sync scheduled.</>}
                            </small>
                        </Text>
                    </div>
                </div>
                <div className="flex-shrink-0 ml-3">
                    <TestConnectionButton
                        className="test-code-host-connection-button"
                        title="Test"
                        buttonVariant="secondary"
                        // FIXME: Maybe update button if check in progress
                        buttonLabel="Test"
                        buttonSubtitle="Check code host connection"
                        buttonDisabled={!node.hasConnectionCheck}
                        flashText="Checking..."
                        run={async () => {
                            await testExternalServiceConnection(node.id)
                        }}
                    // history={}
                    />
                    <Tooltip content={`${editingDisabled ? 'View' : 'Edit'} code host connection settings`}>
                        <Button
                            className="test-edit-external-service-button"
                            to={`${routingPrefix}/external-services/${node.id}`}
                            variant="secondary"
                            size="sm"
                            as={Link}
                        >
                            <Icon aria-hidden={true} svgPath={mdiCog} /> {editingDisabled ? 'View' : 'Edit'}
                        </Button>
                    </Tooltip>{' '}
                    <Tooltip content="Delete code host connection">
                        <Button
                            aria-label="Delete"
                            className="test-delete-external-service-button"
                            onClick={onDelete}
                            disabled={isDeleting === true}
                            variant="danger"
                            size="sm"
                        >
                            <Icon aria-hidden={true} svgPath={mdiDelete} />
                        </Button>
                    </Tooltip>
                </div>
            </div>
            {node.lastSyncError !== null && (
                <ErrorAlert error={node.lastSyncError} variant="danger" className="mt-2 mb-0" />
            )}
            {isErrorLike(isDeleting) && <ErrorAlert className="mt-2" error={isDeleting} />}
        </li>
    )
}

interface Props {
    title: React.ReactNode
    buttonVariant?: ButtonProps['variant']
    buttonLabel: React.ReactNode
    buttonSubtitle?: string
    buttonDisabled?: boolean
    info?: React.ReactNode
    className?: string

    /** The message to briefly display below the button when the action is successful. */
    flashText?: string

    run: () => Promise<void>
    history: H.History
}

interface State {
    loading: boolean
    flash: boolean
    error?: string
}


class TestConnectionButton extends React.PureComponent<Props, State> {
    public state: State = {
        loading: false,
        flash: false,
    }

    private timeoutHandle?: number

    public componentWillUnmount(): void {
        if (this.timeoutHandle) {
            window.clearTimeout(this.timeoutHandle)
        }
    }

    private onClick = (): void => {
        this.setState({
            error: undefined,
            loading: true,
        })

        this.props.run().then(
            () => {
                this.setState({ loading: false, flash: true })
                if (typeof this.timeoutHandle === 'number') {
                    window.clearTimeout(this.timeoutHandle)
                }
                this.timeoutHandle = window.setTimeout(() => this.setState({ flash: false }), 1000)
            },
            error => this.setState({ loading: false, error: asError(error).message })
        )
    }

    public render(): JSX.Element | null {
        let content
        if (!this.props.buttonDisabled) {
            content = this.props.buttonSubtitle
        } else {
            content = "Connectivity check is not implemented for this code host"
        }

        let buttonLabelElement
        if (this.state.loading) {
            buttonLabelElement = <><LoadingSpinner /> Checking</>
        } else {
            buttonLabelElement=<><Icon aria-hidden={true} svgPath={mdiConnection} /> Test</>
        }

        return (
            <>
                <Tooltip content={content}>
                    <Button
                        className={this.props.className}
                        variant={this.props.buttonVariant || "primary"}
                        onClick={this.onClick}
                        disabled={this.props.buttonDisabled || this.state.loading}
                        size="sm"
                    >
                {buttonLabelElement}
                </Button>
                </Tooltip>{' '}

                {this.props.flashText && this.state.flash && (
                    <div className={classNames("flash")}>
                        <small>{this.props.flashText}</small>
                    </div>
                )}
            </>
        )
    }
}
