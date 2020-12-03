import React from 'react'
import { ChangesetSpecFields, ChangesetSpecOperation } from '../../../graphql-operations'
import BlankCircleIcon from 'mdi-react/CheckboxBlankCircleOutlineIcon'
import ImportIcon from 'mdi-react/ImportIcon'
import UploadIcon from 'mdi-react/UploadIcon'
import SourceBranchRefreshIcon from 'mdi-react/SourceBranchRefreshIcon'
import SourceBranchSyncIcon from 'mdi-react/SourceBranchSyncIcon'
import SourceBranchCheckIcon from 'mdi-react/SourceBranchCheckIcon'
import BeakerQuestionIcon from 'mdi-react/BeakerQuestionIcon'
import classNames from 'classnames'
import CloseCircleOutlineIcon from 'mdi-react/CloseCircleOutlineIcon'

export interface ChangesetSpecActionProps {
    spec: ChangesetSpecFields
    className?: string
}

export const ChangesetSpecAction: React.FunctionComponent<ChangesetSpecActionProps> = ({ spec, className }) => {
    if (spec.operations.length === 0) {
        return (
            <ChangesetSpecActionNoAction
                reason={
                    spec.__typename === 'HiddenChangesetSpec'
                        ? NoActionReasonStrings[NoActionReason.NO_ACCESS]
                        : undefined
                }
                className={className}
            />
        )
    }
    if (spec.operations.includes(ChangesetSpecOperation.IMPORT)) {
        return <ChangesetSpecActionImport className={className} />
    }
    if (spec.operations.includes(ChangesetSpecOperation.PUBLISH)) {
        return <ChangesetSpecActionPublish className={className} />
    }
    if (spec.operations.includes(ChangesetSpecOperation.PUBLISH_DRAFT)) {
        return <ChangesetSpecActionPublishDraft className={className} />
    }
    if (spec.operations.includes(ChangesetSpecOperation.CLOSE)) {
        return <ChangesetSpecActionClose className={className} />
    }
    if (spec.operations.includes(ChangesetSpecOperation.REOPEN)) {
        return <ChangesetSpecActionReopen className={className} />
    }
    if (spec.operations.includes(ChangesetSpecOperation.UNDRAFT)) {
        return <ChangesetSpecActionUndraft className={className} />
    }
    if (
        spec.operations.includes(ChangesetSpecOperation.UPDATE) ||
        spec.operations.includes(ChangesetSpecOperation.PUSH)
    ) {
        return <ChangesetSpecActionUpdate className={className} />
    }
    return <ChangesetSpecActionUnknown operations={spec.operations.join(' => ')} className={className} />
}

const iconClassNames = 'm-0 text-nowrap d-block d-sm-flex flex-column align-items-center justify-content-center'

export const ChangesetSpecActionPublish: React.FunctionComponent<{ className?: string }> = ({ className }) => (
    <div className={classNames(className, iconClassNames)}>
        <UploadIcon data-tooltip="This changeset will be published to its code host" />
        <span>Publish</span>
    </div>
)
export const ChangesetSpecActionPublishDraft: React.FunctionComponent<{ className?: string }> = ({ className }) => (
    <div className={classNames(className, iconClassNames)}>
        <UploadIcon className="text-muted" data-tooltip="This changeset will be published as draft to its code host" />
        <span>Publish draft</span>
    </div>
)
export const ChangesetSpecActionImport: React.FunctionComponent<{ className?: string }> = ({ className }) => (
    <div className={classNames(className, iconClassNames)}>
        <ImportIcon data-tooltip="This changeset will be imported and tracked in this campaign" />
        <span>Import</span>
    </div>
)
export const ChangesetSpecActionClose: React.FunctionComponent<{ className?: string }> = ({ className }) => (
    <div className={classNames(className, iconClassNames)}>
        <CloseCircleOutlineIcon data-tooltip="This changeset will be closed on the codehost" />
        <span>Close</span>
    </div>
)
export const ChangesetSpecActionReopen: React.FunctionComponent<{ className?: string }> = ({ className }) => (
    <div className={classNames(className, iconClassNames)}>
        <SourceBranchRefreshIcon data-tooltip="This changeset will be reopened on the codehost" />
        <span>Reopen</span>
    </div>
)
export const ChangesetSpecActionUndraft: React.FunctionComponent<{ className?: string }> = ({ className }) => (
    <div className={classNames(className, iconClassNames)}>
        <SourceBranchCheckIcon data-tooltip="This changeset will be marked as ready for review on the codehost" />
        <span>Undraft</span>
    </div>
)
export const ChangesetSpecActionUpdate: React.FunctionComponent<{ className?: string }> = ({ className }) => (
    <div className={classNames(className, iconClassNames)}>
        <SourceBranchSyncIcon data-tooltip="This changeset will be updated on the codehost" />
        <span>Update</span>
    </div>
)
export const ChangesetSpecActionUnknown: React.FunctionComponent<{ className?: string; operations: string }> = ({
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
export const ChangesetSpecActionNoAction: React.FunctionComponent<{ className?: string; reason?: string }> = ({
    className,
    reason,
}) => (
    <div className={classNames(className, iconClassNames, 'text-muted')}>
        <BlankCircleIcon data-tooltip={reason} />
        <span>No action</span>
    </div>
)
