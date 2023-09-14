import React from 'react'

import { UserAvatar } from '@sourcegraph/shared/src/components/UserAvatar'
import { H3, Tooltip } from '@sourcegraph/wildcard'

import type { VisibleChangesetApplyPreviewFields } from '../../../../graphql-operations'
import { formatPersonName, PersonLink } from '../../../../person/PersonLink'

import styles from './GitBranchChangesetDescriptionInfo.module.scss'

interface Props {
    node: VisibleChangesetApplyPreviewFields
}

export const GitBranchChangesetDescriptionInfo: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    node,
}) => {
    if (node.targets.__typename === 'VisibleApplyPreviewTargetsDetach') {
        return <></>
    }

    if (node.targets.changesetSpec.description.__typename === 'ExistingChangesetReference') {
        return <></>
    }

    if (node.targets.changesetSpec.description.commits.length !== 1) {
        return <></>
    }

    const commit = node.targets.changesetSpec.description.commits[0]

    const previousCommit =
        node.targets.__typename === 'VisibleApplyPreviewTargetsUpdate' &&
        node.targets.changeset.currentSpec?.description.__typename === 'GitBranchChangesetDescription' &&
        node.targets.changeset.currentSpec.description.commits?.[0]

    return (
        <div className={styles.gitBranchChangesetDescriptionInfoGrid}>
            {(node.delta.authorEmailChanged || node.delta.authorNameChanged || node.delta.commitMessageChanged) &&
                previousCommit && (
                    <>
                        <DeletedEntry
                            deleted={node.delta.authorEmailChanged || node.delta.authorNameChanged}
                            className="text-muted"
                        >
                            <div className="d-flex flex-column align-items-center mr-3">
                                <Tooltip content={formatPersonName(previousCommit.author)}>
                                    <span>
                                        <UserAvatar inline={true} className="mb-1" user={previousCommit.author} />
                                    </span>
                                </Tooltip>{' '}
                                <PersonLink person={previousCommit.author} className="font-weight-bold text-nowrap" />
                            </div>
                        </DeletedEntry>
                        <div className="text-muted">
                            <DeletedEntry deleted={node.delta.commitMessageChanged}>
                                <H3 className="text-muted">{previousCommit.subject}</H3>
                            </DeletedEntry>
                            {previousCommit.body && (
                                <DeletedEntry deleted={node.delta.commitMessageChanged}>
                                    <pre className="text-wrap">{previousCommit.body}</pre>
                                </DeletedEntry>
                            )}
                        </div>
                    </>
                )}
            <div className="d-flex flex-column align-items-center mr-3">
                <Tooltip content={formatPersonName(commit.author)}>
                    <span>
                        <UserAvatar inline={true} className="mb-1" user={commit.author} />
                    </span>
                </Tooltip>{' '}
                <PersonLink person={commit.author} className="font-weight-bold text-nowrap" />
            </div>
            <div>
                <H3>{commit.subject}</H3>
                {commit.body && <pre className="text-wrap">{commit.body}</pre>}
            </div>
        </div>
    )
}

const DeletedEntry: React.FunctionComponent<
    React.PropsWithChildren<{ children: React.ReactNode; deleted: boolean; className?: string }>
> = ({ children, deleted, className }) => {
    if (deleted) {
        return <del className={className}>{children}</del>
    }
    return <div className={className}>{children}</div>
}
