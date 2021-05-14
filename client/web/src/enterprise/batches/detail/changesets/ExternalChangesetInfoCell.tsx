import classNames from 'classnames'
import React from 'react'

import { Link } from '@sourcegraph/shared/src/components/Link'

import {
    ExternalChangesetFields,
    GitBranchChangesetDescriptionFields,
    ChangesetState,
} from '../../../../graphql-operations'

import { ChangesetLabel } from './ChangesetLabel'
import { ChangesetLastSynced } from './ChangesetLastSynced'
import { ExternalChangesetTitle } from './ExternalChangesetTitle'

export interface ExternalChangesetInfoCellProps {
    node: ExternalChangesetFields
    viewerCanAdminister: boolean
    className?: string
}

export const ExternalChangesetInfoCell: React.FunctionComponent<ExternalChangesetInfoCellProps> = ({
    node,
    viewerCanAdminister,
    className,
}) => {
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
                <h3 className={classNames('m-0 d-md-inline-block', { 'mr-2': node.labels.length > 0 })}>
                    {changesetTitle}
                </h3>
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
                        <div className="d-block d-sm-inline-block">
                            <span className="badge badge-secondary text-monospace">{headReference(node)}</span>
                        </div>
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

function headReference(node: ExternalChangesetFields): string | undefined {
    if (hasHeadReference(node)) {
        return node.currentSpec?.description.headRef
    }
    return undefined
}

function hasHeadReference(
    node: ExternalChangesetFields
): node is ExternalChangesetFields & {
    currentSpec: ExternalChangesetFields & { description: GitBranchChangesetDescriptionFields }
} {
    return node.currentSpec?.description.__typename === 'GitBranchChangesetDescription'
}
