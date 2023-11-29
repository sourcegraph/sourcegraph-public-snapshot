import React from 'react'

import {
    mdiArchive,
    mdiBeakerQuestion,
    mdiCheckboxBlankCircleOutline,
    mdiCloseCircleOutline,
    mdiDelete,
    mdiImport,
    mdiPaperclip,
    mdiSourceBranchCheck,
    mdiSourceBranchRefresh,
    mdiSourceBranchSync,
    mdiUpload,
    mdiUploadNetwork,
} from '@mdi/js'
import classNames from 'classnames'

import { pluralize } from '@sourcegraph/common'
import { H3, H4, Icon, Tooltip } from '@sourcegraph/wildcard'

import { type ChangesetApplyPreviewFields, ChangesetSpecOperation } from '../../../../graphql-operations'

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
        case ChangesetSpecOperation.IMPORT: {
            return <PreviewActionImport className={className} />
        }
        case ChangesetSpecOperation.PUBLISH: {
            return <PreviewActionPublish className={className} />
        }
        case ChangesetSpecOperation.PUBLISH_DRAFT: {
            return <PreviewActionPublishDraft className={className} />
        }
        case ChangesetSpecOperation.CLOSE: {
            return <PreviewActionClose className={className} />
        }
        case ChangesetSpecOperation.REOPEN: {
            return <PreviewActionReopen className={className} />
        }
        case ChangesetSpecOperation.UNDRAFT: {
            return <PreviewActionUndraft className={className} />
        }
        case ChangesetSpecOperation.UPDATE: {
            return <PreviewActionUpdate className={className} />
        }
        case ChangesetSpecOperation.PUSH: {
            return <PreviewActionPush className={className} />
        }
        case ChangesetSpecOperation.DETACH: {
            return <PreviewActionDetach className={className} />
        }
        case ChangesetSpecOperation.ARCHIVE: {
            return <PreviewActionArchive className={className} />
        }
        case ChangesetSpecOperation.REATTACH: {
            return <PreviewActionReattach className={className} />
        }
        case ChangesetSpecOperation.SYNC:
        case ChangesetSpecOperation.SLEEP: {
            // We don't want to expose these states.
            return null
        }
        default: {
            return <PreviewActionUnknown operations={operations.join(' => ')} className={className} />
        }
    }
}

const iconClassNames = 'm-0 text-nowrap'

export const PreviewPublishStat: React.FunctionComponent<
    React.PropsWithChildren<{ count: number; className?: string }>
> = ({ count, className }) => (
    <div className={classNames(className, iconClassNames)}>
        <Icon aria-hidden={true} svgPath={mdiUpload} />
        <H4
            as={H3}
            className="font-weight-normal text-muted m-0"
            aria-label={`${count} ${pluralize('changeset', count)} will be published`}
        >
            {count} publish
        </H4>
    </div>
)

export const PreviewActionPublish: React.FunctionComponent<React.PropsWithChildren<{ className?: string }>> = ({
    className,
}) => (
    <div className={classNames(className, iconClassNames)}>
        <Tooltip content="This changeset will be published to its code host">
            <Icon aria-label="This changeset will be published to its code host" className="mr-1" svgPath={mdiUpload} />
        </Tooltip>
        <span aria-hidden={true}>Publish</span>
    </div>
)

export const PreviewActionPublishDraft: React.FunctionComponent<React.PropsWithChildren<{ className?: string }>> = ({
    className,
}) => (
    <div className={classNames(className, iconClassNames)}>
        <Tooltip content="This changeset will be published as a draft to its code host">
            <Icon
                aria-label="This changeset will be published as a draft to its code host"
                className="text-muted mr-1"
                svgPath={mdiUpload}
            />
        </Tooltip>
        <span aria-hidden={true}>Publish draft</span>
    </div>
)

export const PreviewImportStat: React.FunctionComponent<
    React.PropsWithChildren<{ count: number; className?: string }>
> = ({ count, className }) => (
    <div className={classNames(className, iconClassNames)}>
        <Icon aria-hidden={true} svgPath={mdiImport} />
        <H4
            as={H3}
            className="font-weight-normal text-muted m-0"
            aria-label={`${count} ${pluralize('changeset', count)} will be imported`}
        >
            {count} import
        </H4>
    </div>
)

export const PreviewActionImport: React.FunctionComponent<React.PropsWithChildren<{ className?: string }>> = ({
    className,
}) => (
    <div className={classNames(className, iconClassNames)}>
        <Tooltip content="This changeset will be imported and tracked in this batch change">
            <Icon
                aria-label="This changeset will be imported and tracked in this batch change"
                className="mr-1"
                svgPath={mdiImport}
            />
        </Tooltip>
        <span aria-hidden={true}>Import</span>
    </div>
)

export const PreviewCloseStat: React.FunctionComponent<
    React.PropsWithChildren<{ count: number; className?: string }>
> = ({ count, className }) => (
    <div className={classNames(className, iconClassNames, 'text-danger')}>
        <Icon aria-hidden={true} svgPath={mdiCloseCircleOutline} />
        <H4
            as={H3}
            className="font-weight-normal text-muted m-0"
            aria-label={`${count} ${pluralize('changeset', count)} will be closed`}
        >
            {count} close
        </H4>
    </div>
)

export const PreviewActionClose: React.FunctionComponent<React.PropsWithChildren<{ className?: string }>> = ({
    className,
}) => (
    <div className={classNames(className, iconClassNames)}>
        <Tooltip content="This changeset will be closed on the code host">
            <Icon
                aria-label="This changeset will be closed on the code host"
                className="text-danger mr-1"
                svgPath={mdiCloseCircleOutline}
            />
        </Tooltip>
        <span aria-hidden={true}>Close</span>
    </div>
)

export const PreviewActionDetach: React.FunctionComponent<React.PropsWithChildren<{ className?: string }>> = ({
    className,
}) => (
    <div className={classNames(className, iconClassNames)}>
        <Tooltip content="This changeset will be removed from the batch change">
            <Icon
                aria-label="This changeset will be removed from the batch change"
                className="text-danger mr-1"
                svgPath={mdiDelete}
            />
        </Tooltip>
        <span aria-hidden={true}>Detach</span>
    </div>
)

export const PreviewReopenStat: React.FunctionComponent<
    React.PropsWithChildren<{ count: number; className?: string }>
> = ({ count, className }) => (
    <div className={classNames(className, iconClassNames, 'text-success')}>
        <Icon aria-hidden={true} svgPath={mdiSourceBranchRefresh} />
        <H4
            as={H3}
            className="font-weight-normal text-muted m-0"
            aria-label={`${count} ${pluralize('changeset', count)} will be reopened`}
        >
            {count} reopen
        </H4>
    </div>
)

export const PreviewActionReopen: React.FunctionComponent<React.PropsWithChildren<{ className?: string }>> = ({
    className,
}) => (
    <div className={classNames(className, iconClassNames)}>
        <Tooltip content="This changeset will be reopened on the code host">
            <Icon
                aria-label="This changeset will be reopened on the code host"
                className="text-success mr-1"
                svgPath={mdiSourceBranchRefresh}
            />
        </Tooltip>
        <span aria-hidden={true}>Reopen</span>
    </div>
)

export const PreviewUndraftStat: React.FunctionComponent<
    React.PropsWithChildren<{ count: number; className?: string }>
> = ({ count, className }) => (
    <div className={classNames(className, iconClassNames, 'text-success')}>
        <Icon aria-hidden={true} svgPath={mdiSourceBranchCheck} />
        <H4
            as={H3}
            className="font-weight-normal text-muted m-0"
            aria-label={`${count} ${pluralize('changeset', count)} will be promoted from ${pluralize('draft', count)}`}
        >
            {count} undraft
        </H4>
    </div>
)

export const PreviewActionUndraft: React.FunctionComponent<React.PropsWithChildren<{ className?: string }>> = ({
    className,
}) => (
    <div className={classNames(className, iconClassNames)}>
        <Tooltip content="This changeset will be marked as ready for review on the code host">
            <Icon
                aria-label="This changeset will be marked as ready for review on the code host"
                className="text-success mr-1"
                svgPath={mdiSourceBranchCheck}
            />
        </Tooltip>
        <span aria-hidden={true}>Undraft</span>
    </div>
)

export const PreviewUpdateStat: React.FunctionComponent<
    React.PropsWithChildren<{ count: number; className?: string }>
> = ({ count, className }) => (
    <div className={classNames(className, iconClassNames)}>
        <Icon aria-hidden={true} svgPath={mdiSourceBranchSync} />
        <H4
            as={H3}
            className="font-weight-normal text-muted m-0"
            aria-label={`${count} ${pluralize('changeset', count)} will be updated`}
        >
            {count} update
        </H4>
    </div>
)

export const PreviewActionUpdate: React.FunctionComponent<React.PropsWithChildren<{ className?: string }>> = ({
    className,
}) => (
    <div className={classNames(className, iconClassNames)}>
        <Tooltip content="This changeset will be updated on the code host">
            <Icon
                aria-label="This changeset will be updated on the code host"
                className="mr-1"
                svgPath={mdiSourceBranchSync}
            />
        </Tooltip>
        <span aria-hidden={true}>Update</span>
    </div>
)

export const PreviewActionPush: React.FunctionComponent<React.PropsWithChildren<{ className?: string }>> = ({
    className,
}) => (
    <div className={classNames(className, iconClassNames)}>
        <Tooltip content="A new commit will be pushed to the code host">
            <Icon
                aria-label="A new commit will be pushed to the code host"
                className="mr-1"
                svgPath={mdiUploadNetwork}
            />
        </Tooltip>
        <span aria-hidden={true}>Push</span>
    </div>
)

export const PreviewActionUnknown: React.FunctionComponent<
    React.PropsWithChildren<{ className?: string; operations: string }>
> = ({ operations, className }) => (
    <div className={classNames(className, iconClassNames)}>
        <Tooltip content={`The operation ${operations} can't yet be displayed.`}>
            <Icon
                aria-label={`The operation ${operations} can't yet be displayed.`}
                className="mr-1"
                svgPath={mdiBeakerQuestion}
            />
        </Tooltip>
        <span aria-hidden={true}>Unknown</span>
    </div>
)

export const PreviewArchiveStat: React.FunctionComponent<
    React.PropsWithChildren<{ count: number; className?: string }>
> = ({ count, className }) => (
    <div className={classNames(className, iconClassNames)}>
        <Icon aria-hidden={true} svgPath={mdiArchive} />
        <H4
            as={H3}
            className="font-weight-normal text-muted m-0"
            aria-label={`${count} ${pluralize('changeset', count)} will be archived`}
        >
            {count} archive
        </H4>
    </div>
)

export const PreviewActionArchive: React.FunctionComponent<React.PropsWithChildren<{ className?: string }>> = ({
    className,
}) => (
    <div className={classNames(className, iconClassNames)}>
        <Tooltip content="This changeset will be kept and marked as archived in this batch change">
            <Icon
                aria-label="This changeset will be kept and marked as archived in this batch change"
                className="text-muted mr-1"
                svgPath={mdiArchive}
            />
        </Tooltip>
        <span aria-hidden={true}>Archive</span>
    </div>
)

export const PreviewReattachStat: React.FunctionComponent<
    React.PropsWithChildren<{ count: number; className?: string }>
> = ({ count, className }) => (
    <div className={classNames(className, iconClassNames)}>
        <Icon aria-hidden={true} svgPath={mdiPaperclip} />
        <H4
            as={H3}
            className="font-weight-normal text-muted m-0"
            aria-label={`${count} ${pluralize('changeset', count)} will be re-added`}
        >
            {count} reattach
        </H4>
    </div>
)

export const PreviewActionReattach: React.FunctionComponent<React.PropsWithChildren<{ className?: string }>> = ({
    className,
}) => (
    <div className={classNames(className, iconClassNames)}>
        <Tooltip content="This changeset will be re-added to the batch change">
            <Icon
                aria-label="This changeset will be re-added to the batch change"
                className="text-muted mr-1"
                svgPath={mdiPaperclip}
            />
        </Tooltip>
        <span aria-hidden={true}>Reattach</span>
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
        <Tooltip content={reason ?? 'The state of this changeset will not change.'}>
            <Icon
                aria-label={reason ?? 'The state of this changeset will not change'}
                className="mr-1"
                svgPath={mdiCheckboxBlankCircleOutline}
            />
        </Tooltip>
        <span aria-hidden={true}>No action</span>
    </div>
)
