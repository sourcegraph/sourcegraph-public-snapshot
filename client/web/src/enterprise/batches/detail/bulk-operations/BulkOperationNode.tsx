import classNames from 'classnames'
import ExternalLinkIcon from 'mdi-react/ExternalLinkIcon'
import React from 'react'

import { Link } from '@sourcegraph/shared/src/components/Link'
import { BulkOperationState } from '@sourcegraph/shared/src/graphql-operations'
import { pluralize } from '@sourcegraph/shared/src/util/strings'

import { ErrorMessage } from '../../../../components/alerts'
import { Timestamp } from '../../../../components/time/Timestamp'
import { BulkOperationFields } from '../../../../graphql-operations'

export interface BulkOperationNodeProps {
    node: BulkOperationFields
}

export const BulkOperationNode: React.FunctionComponent<BulkOperationNodeProps> = ({ node }) => (
    <div className="card mb-3">
        <div className="card-body">
            <div className="d-flex justify-content-between">
                <div>
                    <Link to={node.initiator.url}>{node.initiator.username}</Link> ran type <span>{node.type}</span>{' '}
                    over {node.changesetCount} {pluralize('changeset', node.changesetCount)}{' '}
                    <Timestamp date={node.createdAt} />
                    <p
                        className={classNames(
                            node.state === BulkOperationState.PROCESSING && 'text-info',
                            node.state === BulkOperationState.FAILED && 'text-danger',
                            node.state === BulkOperationState.COMPLETED && 'text-success'
                        )}
                    >
                        {node.state}
                    </p>
                </div>
                <div className="flex-grow-1" style={{ maxWidth: '200px' }}>
                    <div>
                        <progress value={node.progress} className="w-100" max={1} />
                    </div>
                    <p className="text-center">{Math.ceil(node.progress * 100)}%</p>
                </div>
            </div>
            {node.errors.map((error, index) => (
                <div className="alert alert-danger" key={index}>
                    <p>
                        Failed to run task for{' '}
                        {error.changeset.__typename === 'HiddenExternalChangeset' ? (
                            <span className="text-muted">hidden repository.</span>
                        ) : (
                            <>
                                <a href={error.changeset.externalURL?.url}>
                                    {error.changeset.title} <ExternalLinkIcon className="icon-inline" />
                                </a>{' '}
                                on repo <a href={error.changeset.repository.url}>{error.changeset.repository.name}</a>.
                            </>
                        )}
                    </p>
                    <ErrorMessage error={'```\n' + error.error! + '\n```'} />
                </div>
            ))}
        </div>
    </div>
)
