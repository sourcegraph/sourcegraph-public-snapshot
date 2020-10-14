import React from 'react'
import { ChangesetSpecType, ChangesetSpecFields } from '../../../graphql-operations'
import BlankCircleIcon from 'mdi-react/CheckboxBlankCircleOutlineIcon'
import ImportIcon from 'mdi-react/ImportIcon'
import UploadIcon from 'mdi-react/UploadIcon'

export interface ChangesetSpecActionProps {
    spec: ChangesetSpecFields
}

export const ChangesetSpecAction: React.FunctionComponent<ChangesetSpecActionProps> = ({ spec }) => {
    if (spec.__typename === 'HiddenChangesetSpec') {
        if (spec.type === ChangesetSpecType.BRANCH) {
            return <ChangesetSpecActionNoPublish />
        }
        return <ChangesetSpecActionImport />
    }
    if (spec.description.__typename === 'ExistingChangesetReference') {
        return <ChangesetSpecActionImport />
    }
    if (spec.description.published) {
        return <ChangesetSpecActionPublish />
    }
    return <ChangesetSpecActionNoPublish />
}

const iconClassNames = 'm-0 text-nowrap d-flex flex-column align-items-center justify-content-center'

export const ChangesetSpecActionPublish: React.FunctionComponent<{}> = () => (
    <div className={iconClassNames}>
        <UploadIcon data-tooltip="This changeset will be published to its code host" />
        <span>Publish</span>
    </div>
)
export const ChangesetSpecActionNoPublish: React.FunctionComponent<{}> = () => (
    <div className={`${iconClassNames} text-muted`}>
        <BlankCircleIcon data-tooltip="Set publish: true in the changeset template to publish to the code host" />
        <span>Won't publish</span>
    </div>
)
export const ChangesetSpecActionImport: React.FunctionComponent<{}> = () => (
    <div className={iconClassNames}>
        <ImportIcon data-tooltip="This changeset will be imported and tracked in this campaign" />
        <span>Import</span>
    </div>
)
