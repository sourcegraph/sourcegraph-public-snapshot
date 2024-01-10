import React from 'react'

import {
    mdiCommentOutline,
    mdiLinkVariantRemove,
    mdiSync,
    mdiSourceBranch,
    mdiUpload,
    mdiOpenInNew,
    mdiDownload,
} from '@mdi/js'
import classNames from 'classnames'

import { Timestamp } from '@sourcegraph/branded/src/components/Timestamp'
import { pluralize } from '@sourcegraph/common'
import { BulkOperationState, type BulkOperationType } from '@sourcegraph/shared/src/graphql-operations'
import { Badge, AlertLink, Link, Alert, Icon, H4, Text, ErrorMessage } from '@sourcegraph/wildcard'

import { Collapsible } from '../../../../components/Collapsible'
import type { BulkOperationFields } from '../../../../graphql-operations'

import styles from './BulkOperationNode.module.scss'

const OPERATION_TITLES: Record<BulkOperationType, JSX.Element> = {
    COMMENT: (
        <>
            <Icon aria-hidden={true} className="text-muted" svgPath={mdiCommentOutline} /> Comment on changesets
        </>
    ),
    DETACH: (
        <>
            <Icon aria-hidden={true} className="text-muted" svgPath={mdiLinkVariantRemove} /> Detach changesets
        </>
    ),
    REENQUEUE: (
        <>
            <Icon aria-hidden={true} className="text-muted" svgPath={mdiSync} /> Retry changesets
        </>
    ),
    MERGE: (
        <>
            <Icon aria-hidden={true} className="text-muted" svgPath={mdiSourceBranch} /> Merge changesets
        </>
    ),
    CLOSE: (
        <>
            <Icon aria-hidden={true} className="text-danger" svgPath={mdiSourceBranch} /> Close changesets
        </>
    ),
    PUBLISH: (
        <>
            <Icon aria-hidden={true} className="text-muted" svgPath={mdiUpload} /> Publish changesets
        </>
    ),
    EXPORT: (
        <>
            <Icon aria-hidden={true} className="text-muted" svgPath={mdiDownload} /> Export changesets
        </>
    ),
}

export interface BulkOperationNodeProps {
    node: BulkOperationFields
}

export const BulkOperationNode: React.FunctionComponent<React.PropsWithChildren<BulkOperationNodeProps>> = ({
    node,
}) => (
    <>
        <div
            className={classNames(
                styles.bulkOperationNodeContainer,
                'd-flex justify-content-between align-items-center'
            )}
        >
            <div className={classNames(styles.bulkOperationNodeChangesetCounts, 'text-center')}>
                <Badge variant="secondary" className="mb-2" as="p">
                    {node.changesetCount}
                </Badge>
                <Text className="mb-0">{pluralize('changeset', node.changesetCount)}</Text>
            </div>
            <div className={styles.bulkOperationNodeDivider} />
            <div className="flex-grow-1 ml-3">
                <H4>{OPERATION_TITLES[node.type]}</H4>
                <Text className="mb-0">
                    <Link to={node.initiator.url}>{node.initiator.username}</Link> <Timestamp date={node.createdAt} />
                </Text>
            </div>
            {node.state === BulkOperationState.PROCESSING && (
                <div className={classNames(styles.bulkOperationNodeProgressBar, 'flex-grow-1 ml-3')}>
                    <meter value={node.progress} className="w-100" min={0} max={1} />
                    <Text alignment="center" className="mb-0">
                        {Math.ceil(node.progress * 100)}%
                    </Text>
                </div>
            )}
            {node.state === BulkOperationState.FAILED && (
                <Badge variant="danger" className="text-uppercase">
                    failed
                </Badge>
            )}
            {node.state === BulkOperationState.COMPLETED && (
                <Badge variant="success" className="text-uppercase">
                    complete
                </Badge>
            )}
        </div>
        {node.errors.length > 0 && (
            <div className={classNames(styles.bulkOperationNodeErrors, 'px-4')}>
                <Collapsible
                    titleClassName="flex-grow-1 p-3"
                    title={<H4 className="mb-0">The following errors occured while running this task:</H4>}
                >
                    {node.errors.map((error, index) => (
                        <Alert className="mt-2" key={index} variant="danger">
                            <Text>
                                {error.changeset.__typename === 'HiddenExternalChangeset' ? (
                                    <span className="text-muted">On hidden repository</span>
                                ) : (
                                    <>
                                        <AlertLink to={error.changeset.externalURL?.url ?? ''}>
                                            {error.changeset.title} <Icon aria-hidden={true} svgPath={mdiOpenInNew} />
                                        </AlertLink>{' '}
                                        on{' '}
                                        <AlertLink to={error.changeset.repository.url}>
                                            repository {error.changeset.repository.name}
                                        </AlertLink>
                                        .
                                    </>
                                )}
                            </Text>
                            {error.error && <ErrorMessage error={'```\n' + error.error + '\n```'} />}
                        </Alert>
                    ))}
                </Collapsible>
            </div>
        )}
    </>
)
