import React from 'react'

import classNames from 'classnames'
import ArchiveIcon from 'mdi-react/ArchiveIcon'
import BeakerQuestionIcon from 'mdi-react/BeakerQuestionIcon'
import BlankCircleIcon from 'mdi-react/CheckboxBlankCircleOutlineIcon'
import CloseCircleOutlineIcon from 'mdi-react/CloseCircleOutlineIcon'
import ImportIcon from 'mdi-react/ImportIcon'
import SourceBranchCheckIcon from 'mdi-react/SourceBranchCheckIcon'
import SourceBranchRefreshIcon from 'mdi-react/SourceBranchRefreshIcon'
import SourceBranchSyncIcon from 'mdi-react/SourceBranchSyncIcon'
import TrashIcon from 'mdi-react/TrashIcon'
import UploadIcon from 'mdi-react/UploadIcon'
import UploadNetworkIcon from 'mdi-react/UploadNetworkIcon'

import { Icon } from '@sourcegraph/wildcard'

import { ChangesetApplyPreviewFields, ChangesetSpecOperation } from '../../../../graphql-operations'

export interface PreviewActionsProps {
    node: ChangesetApplyPreviewFields
    className?: string
}

export const PreviewActions: React.FunctionComponent<React.PropsWithChildren<PreviewActionsProps>> = ({
    node,
    className,
}) => (
    <div className={classNames('d-flex flex-column align-items-left justify-content-center', className)}>
        <PreviewActionsContent node={node} />
    </div>
)

const PreviewActionsContent: React.FunctionComponent<React.PropsWithChildren<Pick<PreviewActionsProps, 'node'>>> = ({
    node,
}) => {
    if (node.__typename === 'HiddenChangesetApplyPreview') {
        return <PreviewActionNoAction reason={NoActionReasonStrings[NoActionReason.NO_ACCESS]} />
    }
    if (node.operations.length === 0) {
        return <PreviewActionNoAction />
    }
    return (
        <>
            {node.operations.map((operation, index) => (
                <PreviewAction
                    operation={operation}
                    operations={node.operations}
                    key={operation}
                    className={classNames(index !== node.operations.length - 1 && 'mb-1')}
                />
            ))}
        </>
    )
}

interface PreviewActionProps {
    operation: ChangesetSpecOperation
    operations: ChangesetSpecOperation[]
    className?: string
}

const PreviewAction: React.FunctionComponent<React.PropsWithChildren<PreviewActionProps>> = ({
    operation,
    operations,
    className,
}) => {
    switch (operation) {
        case ChangesetSpecOperation.IMPORT:
            return <PreviewActionImport className={className} />
        case ChangesetSpecOperation.PUBLISH:
            return <PreviewActionPublish className={className} />
        case ChangesetSpecOperation.PUBLISH_DRAFT:
            return <PreviewActionPublishDraft className={className} />
        case ChangesetSpecOperation.CLOSE:
            return <PreviewActionClose className={className} />
        case ChangesetSpecOperation.REOPEN:
            return <PreviewActionReopen className={className} />
        case ChangesetSpecOperation.UNDRAFT:
            return <PreviewActionUndraft className={className} />
        case ChangesetSpecOperation.UPDATE:
            return <PreviewActionUpdate className={className} />
        case ChangesetSpecOperation.PUSH:
            return <PreviewActionPush className={className} />
        case ChangesetSpecOperation.DETACH:
            return <PreviewActionDetach className={className} />
        case ChangesetSpecOperation.ARCHIVE:
            return <PreviewActionArchive className={className} />
        case ChangesetSpecOperation.SYNC:
        case ChangesetSpecOperation.SLEEP:
            // We don't want to expose these states.
            return null
        default:
            return <PreviewActionUnknown operations={operations.join(' => ')} className={className} />
    }
}

const iconClassNames = 'm-0 text-nowrap'

export const PreviewActionPublish: React.FunctionComponent<
    React.PropsWithChildren<{ label?: string; className?: string }>
> = ({ label = 'Publish', className }) => (
    <div className={classNames(className, iconClassNames)}>
        <Icon
            className="mr-1"
            data-tooltip="This changeset will be published to its code host"
            as={UploadIcon}
            role="presentation"
        />
        <span>{label}</span>
    </div>
)

export const PreviewActionPublishDraft: React.FunctionComponent<
    React.PropsWithChildren<{ label?: string; className?: string }>
> = ({ label = 'Publish draft', className }) => (
    <div className={classNames(className, iconClassNames)}>
        <Icon
            className="text-muted mr-1"
            data-tooltip="This changeset will be published as a draft to its code host"
            as={UploadIcon}
            role="presentation"
        />
        <span>{label}</span>
    </div>
)

export const PreviewActionImport: React.FunctionComponent<
    React.PropsWithChildren<{ label?: string; className?: string }>
> = ({ label = 'Import', className }) => (
    <div className={classNames(className, iconClassNames)}>
        <Icon
            className="mr-1"
            data-tooltip="This changeset will be imported and tracked in this batch change"
            as={ImportIcon}
            role="presentation"
        />
        <span>{label}</span>
    </div>
)

export const PreviewActionClose: React.FunctionComponent<
    React.PropsWithChildren<{ label?: string; className?: string }>
> = ({ label = 'Close', className }) => (
    <div className={classNames(className, iconClassNames)}>
        <Icon
            className="text-danger mr-1"
            data-tooltip="This changeset will be closed on the code host"
            as={CloseCircleOutlineIcon}
            role="presentation"
        />
        <span>{label}</span>
    </div>
)

export const PreviewActionDetach: React.FunctionComponent<
    React.PropsWithChildren<{ label?: string; className?: string }>
> = ({ label = 'Detach', className }) => (
    <div className={classNames(className, iconClassNames)}>
        <Icon
            className="text-danger mr-1"
            data-tooltip="This changeset will be removed from the batch change"
            as={TrashIcon}
            role="presentation"
        />
        <span>{label}</span>
    </div>
)

export const PreviewActionReopen: React.FunctionComponent<
    React.PropsWithChildren<{ label?: string; className?: string }>
> = ({ label = 'Reopen', className }) => (
    <div className={classNames(className, iconClassNames)}>
        <Icon
            className="text-success mr-1"
            data-tooltip="This changeset will be reopened on the code host"
            as={SourceBranchRefreshIcon}
            role="presentation"
        />
        <span>{label}</span>
    </div>
)

export const PreviewActionUndraft: React.FunctionComponent<
    React.PropsWithChildren<{ label?: string; className?: string }>
> = ({ label = 'Undraft', className }) => (
    <div className={classNames(className, iconClassNames)}>
        <Icon
            className="text-success mr-1"
            data-tooltip="This changeset will be marked as ready for review on the code host"
            as={SourceBranchCheckIcon}
            role="presentation"
        />
        <span>{label}</span>
    </div>
)

export const PreviewActionUpdate: React.FunctionComponent<
    React.PropsWithChildren<{ label?: string; className?: string }>
> = ({ label = 'Update', className }) => (
    <div className={classNames(className, iconClassNames)}>
        <Icon
            className="mr-1"
            data-tooltip="This changeset will be updated on the code host"
            as={SourceBranchSyncIcon}
            role="presentation"
        />
        <span>{label}</span>
    </div>
)

export const PreviewActionPush: React.FunctionComponent<
    React.PropsWithChildren<{ label?: string; className?: string }>
> = ({ label = 'Push', className }) => (
    <div className={classNames(className, iconClassNames)}>
        <Icon
            className="mr-1"
            data-tooltip="A new commit will be pushed to the code host"
            as={UploadNetworkIcon}
            role="presentation"
        />
        <span>{label}</span>
    </div>
)

export const PreviewActionUnknown: React.FunctionComponent<
    React.PropsWithChildren<{ className?: string; operations: string }>
> = ({ operations, className }) => (
    <div className={classNames(className, iconClassNames)}>
        <Icon
            className="mr-1"
            data-tooltip={`The operation ${operations} can't yet be displayed.`}
            as={BeakerQuestionIcon}
            role="presentation"
        />
        <span>Unknown</span>
    </div>
)

export const PreviewActionArchive: React.FunctionComponent<
    React.PropsWithChildren<{ label?: string; className?: string }>
> = ({ label = 'Archive', className }) => (
    <div className={classNames(className, iconClassNames)}>
        <Icon
            className="text-muted mr-1"
            data-tooltip="This changeset will be kept and marked as archived in this batch change"
            as={ArchiveIcon}
            role="presentation"
        />
        <span>{label}</span>
    </div>
)

export enum NoActionReason {
    NO_ACCESS = 'no-access',
}

export const NoActionReasonStrings: Record<NoActionReason, string> = {
    [NoActionReason.NO_ACCESS]: "You don't have access to the repository this changeset spec targets.",
}

export const PreviewActionNoAction: React.FunctionComponent<
    React.PropsWithChildren<{ className?: string; reason?: string }>
> = ({ className, reason }) => (
    <div className={classNames(className, iconClassNames, 'text-muted')}>
        <Icon
            className="mr-1"
            data-tooltip={reason}
            as={BlankCircleIcon}
            role="presentation"
        />
        <span>No action</span>
    </div>
)
