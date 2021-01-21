import React, { useCallback, useState } from 'react'
import { VisibleChangesetApplyPreviewFields } from '../../../../graphql-operations'
import { formatPersonName, PersonLink } from '../../../../person/PersonLink'
import { UserAvatar } from '../../../../user/UserAvatar'

const dotDotDot = '\u22EF'

interface Props {
    node: VisibleChangesetApplyPreviewFields

    /** For testing only. */
    isExpandedInitially?: boolean
}

export const GitBranchChangesetDescriptionInfo: React.FunctionComponent<Props> = ({
    node,
    isExpandedInitially = false,
}) => {
    const [showCommitMessageBody, setShowCommitMessageBody] = useState<boolean>(isExpandedInitially)

    const toggleShowCommitMessageBody = useCallback((): void => {
        setShowCommitMessageBody(previousState => !previousState)
    }, [])

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
        <>
            {(node.delta.authorEmailChanged || node.delta.authorNameChanged || node.delta.commitMessageChanged) &&
                previousCommit && (
                    <>
                        <div className="w-100">
                            <div className="d-flex align-items-center flex-grow-1">
                                <DeletedEntry deleted={node.delta.authorEmailChanged || node.delta.authorNameChanged}>
                                    <small className="mr-2">
                                        <UserAvatar
                                            className="icon-inline mr-1"
                                            user={previousCommit.author}
                                            data-tooltip={formatPersonName(previousCommit.author)}
                                        />{' '}
                                        <PersonLink person={previousCommit.author} className="font-weight-bold" />
                                    </small>
                                </DeletedEntry>
                                <DeletedEntry deleted={node.delta.commitMessageChanged}>
                                    <span
                                        className="overflow-hidden font-weight-bold text-nowrap text-truncate pr-2"
                                        data-tooltip={previousCommit.subject}
                                    >
                                        {previousCommit.subject}
                                    </span>
                                </DeletedEntry>
                                {previousCommit.body && (
                                    <button
                                        type="button"
                                        className="btn btn-secondary btn-sm px-1 py-0 font-weight-bold align-item-center mr-2"
                                        onClick={toggleShowCommitMessageBody}
                                    >
                                        {dotDotDot}
                                    </button>
                                )}
                            </div>
                        </div>
                        {showCommitMessageBody && (
                            <DeletedEntry deleted={node.delta.commitMessageChanged}>
                                <div className="w-100 mt-1">
                                    <small>
                                        <pre className="text-wrap mb-0">{previousCommit.body}</pre>
                                    </small>
                                </div>
                            </DeletedEntry>
                        )}
                        <hr className="mb-3" />
                    </>
                )}
            <div className="w-100">
                <div className="d-flex align-items-center flex-grow-1">
                    <small className="mr-2">
                        <UserAvatar
                            className="icon-inline mr-1"
                            user={commit.author}
                            data-tooltip={formatPersonName(commit.author)}
                        />{' '}
                        <PersonLink person={commit.author} className="font-weight-bold" />
                    </small>
                    <span
                        className="overflow-hidden font-weight-bold text-nowrap text-truncate pr-2"
                        data-tooltip={commit.subject}
                    >
                        {commit.subject}
                    </span>
                    {commit.body && (
                        <button
                            type="button"
                            className="btn btn-secondary btn-sm px-1 py-0 font-weight-bold align-item-center mr-2"
                            onClick={toggleShowCommitMessageBody}
                        >
                            {dotDotDot}
                        </button>
                    )}
                </div>
            </div>
            {showCommitMessageBody && (
                <div className="w-100 mt-1">
                    <small>
                        <pre className="text-wrap mb-0">{commit.body}</pre>
                    </small>
                </div>
            )}
            <hr className="mb-3" />
        </>
    )
}

const DeletedEntry: React.FunctionComponent<{ children: React.ReactNode; deleted: boolean }> = ({
    children,
    deleted,
}) => {
    if (deleted) {
        return <del>{children}</del>
    }
    return <>{children}</>
}
