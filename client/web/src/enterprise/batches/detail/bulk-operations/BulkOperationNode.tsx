import classNames from 'classnames'
import React from 'react'

import { BulkOperationState } from '@sourcegraph/shared/src/graphql-operations'

import { ErrorMessage } from '../../../../components/alerts'
import { Timestamp } from '../../../../components/time/Timestamp'
import { BulkOperationFields } from '../../../../graphql-operations'
import ExternalLinkIcon from 'mdi-react/ExternalLinkIcon'

export interface BulkOperationNodeProps {
    node: BulkOperationFields
}

export const BulkOperationNode: React.FunctionComponent<BulkOperationNodeProps> = ({ node }) => (
    <div>
        Ran type <span>{node.type}</span> <Timestamp date={node.createdAt} />
        <p
            className={classNames(
                node.state === BulkOperationState.PROCESSING && 'text-info',
                node.state === BulkOperationState.FAILED && 'text-danger',
                node.state === BulkOperationState.COMPLETED && 'text-success'
            )}
        >
            {node.state}
        </p>
        <div>
            <progress value={node.progress} max={1} /> {Math.ceil(node.progress * 100)}%
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
)
