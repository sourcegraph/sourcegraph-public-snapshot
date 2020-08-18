import React from 'react'
import {
    ChangesetFields,
    ChangesetPublicationState,
    ChangesetReconcilerState,
    ChangesetExternalState,
} from '../../../../graphql-operations'
import SourceBranchIcon from 'mdi-react/SourceBranchIcon'
import SourcePullIcon from 'mdi-react/SourcePullIcon'
import SourceMergeIcon from 'mdi-react/SourceMergeIcon'
import DeleteIcon from 'mdi-react/DeleteIcon'
import ErrorIcon from 'mdi-react/ErrorIcon'
import TimerSandIcon from 'mdi-react/TimerSandIcon'
import classNames from 'classnames'

export interface ChangesetStatusCellProps {
    className?: string
    changeset: Pick<ChangesetFields, 'publicationState' | 'externalState' | 'reconcilerState'>
}

export const ChangesetStatusCell: React.FunctionComponent<ChangesetStatusCellProps> = ({ changeset }) => {
    if (changeset.reconcilerState === ChangesetReconcilerState.ERRORED) {
        return <ChangesetStatusError />
    }
    if (changeset.reconcilerState !== ChangesetReconcilerState.COMPLETED) {
        return <ChangesetStatusProcessing />
    }
    if (changeset.publicationState === ChangesetPublicationState.UNPUBLISHED) {
        return <ChangesetStatusUnpublished />
    }
    // Must be set, because changesetPublicationState !== UNPUBLISHED.
    const externalState = changeset.externalState!
    switch (externalState) {
        case ChangesetExternalState.OPEN:
            return <ChangesetStatusOpen />
        case ChangesetExternalState.CLOSED:
            return <ChangesetStatusClosed />
        case ChangesetExternalState.MERGED:
            return <ChangesetStatusMerged />
        case ChangesetExternalState.DELETED:
            return <ChangesetStatusDeleted />
    }
}

const iconClassNames = 'm-0 text-nowrap d-flex flex-column align-items-center justify-content-center'

export const ChangesetStatusUnpublished: React.FunctionComponent<{ label?: JSX.Element; className?: string }> = ({
    label = <span className="text-muted">Unpublished</span>,
    className,
}) => (
    <div className={classNames(iconClassNames, 'text-muted', className)}>
        <SourceBranchIcon />
        {label}
    </div>
)
export const ChangesetStatusClosed: React.FunctionComponent<{ label?: JSX.Element; className?: string }> = ({
    label = <span className="text-muted">Closed</span>,
    className,
}) => (
    <div className={classNames(iconClassNames, 'text-danger', className)}>
        <SourcePullIcon />
        {label}
    </div>
)
export const ChangesetStatusMerged: React.FunctionComponent<{ label?: JSX.Element; className?: string }> = ({
    label = <span className="text-muted">Merged</span>,
    className,
}) => (
    <div className={classNames(iconClassNames, 'text-merged', className)}>
        <SourceMergeIcon />
        {label}
    </div>
)
export const ChangesetStatusOpen: React.FunctionComponent<{ label?: JSX.Element; className?: string }> = ({
    label = <span className="text-muted">Open</span>,
    className,
}) => (
    <div className={classNames(iconClassNames, 'text-success', className)}>
        <SourcePullIcon />
        {label}
    </div>
)
export const ChangesetStatusDeleted: React.FunctionComponent<{ label?: JSX.Element; className?: string }> = ({
    label = <span className="text-muted">Deleted</span>,
    className,
}) => (
    <div className={classNames(iconClassNames, 'text-muted', className)}>
        <DeleteIcon />
        {label}
    </div>
)
export const ChangesetStatusError: React.FunctionComponent<{ label?: JSX.Element; className?: string }> = ({
    label = <span className="text-danger">Error</span>,
    className,
}) => (
    <div className={classNames(iconClassNames, 'text-danger', className)}>
        <ErrorIcon />
        {label}
    </div>
)
export const ChangesetStatusProcessing: React.FunctionComponent<{ label?: JSX.Element; className?: string }> = ({
    label = <span className="text-muted">Processing</span>,
    className,
}) => (
    <div className={classNames(iconClassNames, className)}>
        <TimerSandIcon className="changeset-status-cell__processing-icon" />
        {label}
    </div>
)
