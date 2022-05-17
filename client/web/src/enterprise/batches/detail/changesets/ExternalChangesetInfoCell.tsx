import React from 'react'

import classNames from 'classnames'

import { Link, Typography } from '@sourcegraph/wildcard'

import { ExternalChangesetFields, ChangesetState } from '../../../../graphql-operations'
import { BranchMerge } from '../../Branch'

import { ChangesetLabel } from './ChangesetLabel'
import { ChangesetLastSynced } from './ChangesetLastSynced'
import { ExternalChangesetTitle } from './ExternalChangesetTitle'

export interface ExternalChangesetInfoCellProps {
    node: ExternalChangesetFields
    viewerCanAdminister: boolean
    className?: string
}

export const ExternalChangesetInfoCell: React.FunctionComponent<
    React.PropsWithChildren<ExternalChangesetInfoCellProps>
> = ({ node, viewerCanAdminister, className }) => {
    const changesetTitle =
        isImporting(node) || importingFailed(node) ? (
            `Importing changeset ${node.externalID ? `#${node.externalID}` : ''}`
        ) : (
            <ExternalChangesetTitle
                className="m-0 d-block d-md-inline"
                externalID={node.externalID}
                externalURL={node.externalURL}
                title={node.title}
            />
        )

    return (
        <div className={classNames('d-flex flex-column', className)}>
            <div className="m-0">
                <Typography.H3 className={classNames('m-0 d-md-inline-block', { 'mr-2': node.labels.length > 0 })}>
                    {changesetTitle}
                </Typography.H3>
                {node.labels.length > 0 && (
                    <span className="d-block d-md-inline-block mr-2">
                        {node.labels.map(label => (
                            <ChangesetLabel label={label} key={label.text} />
                        ))}
                    </span>
                )}
            </div>
            <div>
                <span className="mr-2 d-block">
                    <Link to={node.repository.url} target="_blank" rel="noopener noreferrer">
                        {node.repository.name}
                    </Link>{' '}
                    {hasHeadReference(node) && (
                        <BranchMerge
                            baseRef={node.currentSpec.description.baseRef}
                            forkTarget={
                                node.forkNamespace
                                    ? { pushUser: false, namespace: node.forkNamespace }
                                    : node.currentSpec.forkTarget
                            }
                            headRef={node.currentSpec.description.headRef}
                        />
                    )}
                </span>
                {![
                    ChangesetState.FAILED,
                    ChangesetState.PROCESSING,
                    ChangesetState.RETRYING,
                    ChangesetState.UNPUBLISHED,
                    ChangesetState.SCHEDULED,
                ].includes(node.state) && (
                    <ChangesetLastSynced changeset={node} viewerCanAdminister={viewerCanAdminister} />
                )}
            </div>
        </div>
    )
}

function isImporting(node: ExternalChangesetFields): boolean {
    return node.state === ChangesetState.PROCESSING && !hasHeadReference(node)
}

function importingFailed(node: ExternalChangesetFields): boolean {
    return node.state === ChangesetState.FAILED && !hasHeadReference(node)
}

function hasHeadReference(
    node: ExternalChangesetFields
): node is ExternalChangesetFields & {
    currentSpec: typeof node.currentSpec & {
        description: { __typename: 'GitBranchChangesetDescription' }
    }
} {
    return node.currentSpec?.description.__typename === 'GitBranchChangesetDescription'
}
