import React from 'react'
import { ChangesetSpecType, ChangesetSpecFields, ChangesetSpecOperation } from '../../../graphql-operations'
import BlankCircleIcon from 'mdi-react/CheckboxBlankCircleOutlineIcon'
import ImportIcon from 'mdi-react/ImportIcon'
import UploadIcon from 'mdi-react/UploadIcon'
import classNames from 'classnames'
import CloseCircleOutlineIcon from 'mdi-react/CloseCircleOutlineIcon'

export interface ChangesetSpecActionProps {
    spec: ChangesetSpecFields
    className?: string
}

export const ChangesetSpecAction: React.FunctionComponent<ChangesetSpecActionProps> = ({ spec, className }) => {
    if (spec.operations.length === 0) {
        // If no operations, we don't need to do anything. If the target changeset is null, it will remain unpublished.
        return <ChangesetSpecActionNoAction unpublished={spec.changeset === null} className={className} />
    }
    if (spec.operations[0] === ChangesetSpecOperation.PUBLISH_DRAFT) {
        return <ChangesetSpecActionPublishDraft className={className} />
    }
    if (spec.operations[0] === ChangesetSpecOperation.PUBLISH) {
        return <ChangesetSpecActionPublish className={className} />
    }
    if (spec.operations[0] === ChangesetSpecOperation.CLOSE) {
        return <ChangesetSpecActionClose className={className} />
    }
    // TODO: More states.

    // This should never be reached.
    return <></>
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
        <UploadIcon
            className="text-muted"
            data-tooltip="This changeset will be published as a draft to its code host"
        />
        <span>Publish draft</span>
    </div>
)
export const ChangesetSpecActionClose: React.FunctionComponent<{ className?: string }> = ({ className }) => (
    <div className={classNames(className, iconClassNames, 'text-muted')}>
        <CloseCircleOutlineIcon />
        <span>Close</span>
    </div>
)
export const ChangesetSpecActionNoAction: React.FunctionComponent<{ unpublished: boolean; className?: string }> = ({
    unpublished,
    className,
}) => (
    <div className={classNames(className, iconClassNames, 'text-muted')}>
        <BlankCircleIcon
            data-tooltip={
                unpublished ? 'Set publish: true in the changeset template to publish to the code host' : undefined
            }
        />
        <span>No action</span>
    </div>
)
export const ChangesetSpecActionImport: React.FunctionComponent<{ className?: string }> = ({ className }) => (
    <div className={classNames(className, iconClassNames)}>
        <ImportIcon data-tooltip="This changeset will be imported and tracked in this campaign" />
        <span>Import</span>
    </div>
)
