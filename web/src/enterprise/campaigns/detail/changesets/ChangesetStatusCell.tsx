import React from 'react'
import { ChangesetFields } from '../../../../graphql-operations'
import SourceBranchIcon from 'mdi-react/SourceBranchIcon'
import SourcePullIcon from 'mdi-react/SourcePullIcon'
import SourceMergeIcon from 'mdi-react/SourceMergeIcon'
import DeleteIcon from 'mdi-react/DeleteIcon'
import ErrorIcon from 'mdi-react/ErrorIcon'
import TimerSandIcon from 'mdi-react/TimerSandIcon'
import classNames from 'classnames'
import { computeChangesetUIState, ChangesetUIState } from '../../utils'

export interface ChangesetStatusCellProps {
    className?: string
    changeset: Pick<ChangesetFields, 'publicationState' | 'externalState' | 'reconcilerState'>
}

export const ChangesetStatusCell: React.FunctionComponent<ChangesetStatusCellProps> = React.memo(
    function ChangesetStatusCell({ changeset }) {
        switch (computeChangesetUIState(changeset)) {
            case ChangesetUIState.ERRORED:
                return <ChangesetStatusError />
            case ChangesetUIState.PROCESSING:
                return <ChangesetStatusProcessing />
            case ChangesetUIState.UNPUBLISHED:
                return <ChangesetStatusUnpublished />
            case ChangesetUIState.OPEN:
                return <ChangesetStatusOpen />
            case ChangesetUIState.CLOSED:
                return <ChangesetStatusClosed />
            case ChangesetUIState.MERGED:
                return <ChangesetStatusMerged />
            case ChangesetUIState.DELETED:
                return <ChangesetStatusDeleted />
        }
    },
    ({ changeset: previous }, { changeset: next }) =>
        previous.externalState === next.externalState &&
        previous.publicationState === next.publicationState &&
        previous.reconcilerState === next.reconcilerState
)

const iconClassNames = 'm-0 text-nowrap d-flex flex-column align-items-center justify-content-center'

export const ChangesetStatusUnpublished: React.FunctionComponent<{
    label?: JSX.Element
    className?: string
}> = React.memo(function ChangesetStatusUnpublished({
    label = <span className="text-muted">Unpublished</span>,
    className,
}) {
    return (
        <div className={classNames(iconClassNames, 'text-muted', className)}>
            <SourceBranchIcon />
            {label}
        </div>
    )
})

export const ChangesetStatusClosed: React.FunctionComponent<{ label?: JSX.Element; className?: string }> = React.memo(
    function ChangesetStatusClosed({ label = <span className="text-muted">Closed</span>, className }) {
        return (
            <div className={classNames(iconClassNames, 'text-danger', className)}>
                <SourcePullIcon />
                {label}
            </div>
        )
    }
)

export const ChangesetStatusMerged: React.FunctionComponent<{ label?: JSX.Element; className?: string }> = React.memo(
    function ChangesetStatusMerged({ label = <span className="text-muted">Merged</span>, className }) {
        return (
            <div className={classNames(iconClassNames, 'text-merged', className)}>
                <SourceMergeIcon />
                {label}
            </div>
        )
    }
)

export const ChangesetStatusOpen: React.FunctionComponent<{ label?: JSX.Element; className?: string }> = React.memo(
    function ChangesetStatusOpen({ label = <span className="text-muted">Open</span>, className }) {
        return (
            <div className={classNames(iconClassNames, 'text-success', className)}>
                <SourcePullIcon />
                {label}
            </div>
        )
    }
)

export const ChangesetStatusDeleted: React.FunctionComponent<{ label?: JSX.Element; className?: string }> = React.memo(
    function ChangesetStatusDeleted({ label = <span className="text-muted">Deleted</span>, className }) {
        return (
            <div className={classNames(iconClassNames, 'text-muted', className)}>
                <DeleteIcon />
                {label}
            </div>
        )
    }
)

export const ChangesetStatusError: React.FunctionComponent<{ label?: JSX.Element; className?: string }> = React.memo(
    function ChangesetStatusError({ label = <span className="text-danger">Error</span>, className }) {
        return (
            <div className={classNames(iconClassNames, 'text-danger', className)}>
                <ErrorIcon />
                {label}
            </div>
        )
    }
)

export const ChangesetStatusProcessing: React.FunctionComponent<{
    label?: JSX.Element
    className?: string
}> = React.memo(function ChangesetStatusProcessing({
    label = <span className="text-muted">Processing</span>,
    className,
}) {
    return (
        <div className={classNames(iconClassNames, className)}>
            <TimerSandIcon className="changeset-status-cell__processing-icon" />
            {label}
        </div>
    )
})
