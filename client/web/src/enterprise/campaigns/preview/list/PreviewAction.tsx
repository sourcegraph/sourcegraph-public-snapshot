import React from 'react'
import { ChangesetApplyPreviewFields, ChangesetSpecOperation } from '../../../../graphql-operations'
import BlankCircleIcon from 'mdi-react/CheckboxBlankCircleOutlineIcon'
import ImportIcon from 'mdi-react/ImportIcon'
import UploadIcon from 'mdi-react/UploadIcon'
import TrashIcon from 'mdi-react/TrashIcon'
import SourceBranchRefreshIcon from 'mdi-react/SourceBranchRefreshIcon'
import SourceBranchSyncIcon from 'mdi-react/SourceBranchSyncIcon'
import SourceBranchCheckIcon from 'mdi-react/SourceBranchCheckIcon'
import BeakerQuestionIcon from 'mdi-react/BeakerQuestionIcon'
import classNames from 'classnames'
import CloseCircleOutlineIcon from 'mdi-react/CloseCircleOutlineIcon'

export interface PreviewActionProps {
    node: ChangesetApplyPreviewFields
    className?: string
}

export const PreviewAction: React.FunctionComponent<PreviewActionProps> = ({ node, className }) => {
    if (node.__typename === 'HiddenChangesetApplyPreview') {
        return <PreviewActionNoAction reason={NoActionReasonStrings[NoActionReason.NO_ACCESS]} className={className} />
    }
    if (node.operations.length === 0) {
        if (node.targets.__typename === 'VisibleApplyPreviewTargetsDetach') {
            return <PreviewActionDetach className={className} />
        }
        return <PreviewActionNoAction className={className} />
    }
    if (node.operations.includes(ChangesetSpecOperation.IMPORT)) {
        return <PreviewActionImport className={className} />
    }
    if (node.operations.includes(ChangesetSpecOperation.PUBLISH)) {
        return <PreviewActionPublish className={className} />
    }
    if (node.operations.includes(ChangesetSpecOperation.PUBLISH_DRAFT)) {
        return <PreviewActionPublishDraft className={className} />
    }
    if (node.operations.includes(ChangesetSpecOperation.CLOSE)) {
        return <PreviewActionClose className={className} />
    }
    if (node.operations.includes(ChangesetSpecOperation.REOPEN)) {
        return <PreviewActionReopen className={className} />
    }
    if (node.operations.includes(ChangesetSpecOperation.UNDRAFT)) {
        return <PreviewActionUndraft className={className} />
    }
    if (
        node.operations.includes(ChangesetSpecOperation.UPDATE) ||
        node.operations.includes(ChangesetSpecOperation.PUSH)
    ) {
        return <PreviewActionUpdate className={className} />
    }
    return <PreviewActionUnknown operations={node.operations.join(' => ')} className={className} />
}

const iconClassNames = 'm-0 text-nowrap d-block d-sm-flex flex-column align-items-center justify-content-center'

export const PreviewActionPublish: React.FunctionComponent<{ className?: string }> = ({ className }) => (
    <div className={classNames(className, iconClassNames)}>
        <UploadIcon data-tooltip="This changeset will be published to its code host" />
        <span>Publish</span>
    </div>
)
export const PreviewActionPublishDraft: React.FunctionComponent<{ className?: string }> = ({ className }) => (
    <div className={classNames(className, iconClassNames)}>
        <UploadIcon
            className="text-muted"
            data-tooltip="This changeset will be published as a draft to its code host"
        />
        <span>Publish draft</span>
    </div>
)
export const PreviewActionImport: React.FunctionComponent<{ className?: string }> = ({ className }) => (
    <div className={classNames(className, iconClassNames)}>
        <ImportIcon data-tooltip="This changeset will be imported and tracked in this campaign" />
        <span>Import</span>
    </div>
)
// TODO: This is currently correct, but as soon as we have a detach reconciler operation, that should be taken into account.
export const PreviewActionClose: React.FunctionComponent<{ className?: string }> = ({ className }) => (
    <div className={classNames(className, iconClassNames)}>
        <CloseCircleOutlineIcon className="text-danger" data-tooltip="This changeset will be closed on the code host" />
        <span>Close &amp; Detach</span>
    </div>
)
export const PreviewActionDetach: React.FunctionComponent<{ className?: string }> = ({ className }) => (
    <div className={classNames(className, iconClassNames)}>
        <TrashIcon className="text-danger" data-tooltip="This changeset will be removed from the campaign" />
        <span>Detach</span>
    </div>
)
export const PreviewActionReopen: React.FunctionComponent<{ className?: string }> = ({ className }) => (
    <div className={classNames(className, iconClassNames)}>
        <SourceBranchRefreshIcon data-tooltip="This changeset will be reopened on the code host" />
        <span>Reopen</span>
    </div>
)
export const PreviewActionUndraft: React.FunctionComponent<{ className?: string }> = ({ className }) => (
    <div className={classNames(className, iconClassNames)}>
        <SourceBranchCheckIcon data-tooltip="This changeset will be marked as ready for review on the code host" />
        <span>Undraft</span>
    </div>
)
export const PreviewActionUpdate: React.FunctionComponent<{ className?: string }> = ({ className }) => (
    <div className={classNames(className, iconClassNames)}>
        <SourceBranchSyncIcon data-tooltip="This changeset will be updated on the code host" />
        <span>Update</span>
    </div>
)
export const PreviewActionUnknown: React.FunctionComponent<{ className?: string; operations: string }> = ({
    operations,
    className,
}) => (
    <div className={classNames(className, iconClassNames)}>
        <BeakerQuestionIcon data-tooltip={`The operation ${operations} can't yet be displayed.`} />
        <span>Unknown</span>
    </div>
)
export enum NoActionReason {
    NO_ACCESS = 'no-access',
}
export const NoActionReasonStrings: Record<NoActionReason, string> = {
    [NoActionReason.NO_ACCESS]: "You don't have access to the repository this changeset spec targets.",
}
export const PreviewActionNoAction: React.FunctionComponent<{ className?: string; reason?: string }> = ({
    className,
    reason,
}) => (
    <div className={classNames(className, iconClassNames, 'text-muted')}>
        <BlankCircleIcon data-tooltip={reason} />
        <span>No action</span>
    </div>
)
