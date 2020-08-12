import React from 'react'
import { ChangesetSpecType, ChangesetSpecFields } from '../../../graphql-operations'
import ClipboardCheckOutlineIcon from 'mdi-react/ClipboardCheckOutlineIcon'
import ClipboardAlertOutlineIcon from 'mdi-react/ClipboardAlertOutlineIcon'
import ClipboardArrowUpOutlineIcon from 'mdi-react/ClipboardArrowUpOutlineIcon'

export const ChangesetSpecAction: React.FunctionComponent<{ spec: ChangesetSpecFields }> = ({ spec }) => {
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

export const ChangesetSpecActionPublish: React.FunctionComponent<{}> = () => (
    <div className="m-0 mx-4 text-nowrap d-flex flex-column align-items-center justify-content-center">
        <ClipboardCheckOutlineIcon data-tooltip="This changeset will be published on the code host when the spec is applied." />
        <span className="text-muted">Publish</span>
    </div>
)
export const ChangesetSpecActionNoPublish: React.FunctionComponent<{}> = () => (
    <div className="m-0 mx-4 text-nowrap d-flex flex-column align-items-center justify-content-center">
        <ClipboardAlertOutlineIcon data-tooltip="This changeset will NOT be published on the code host when the spec is applied." />
        <span className="text-muted">Not publish</span>
    </div>
)
export const ChangesetSpecActionImport: React.FunctionComponent<{}> = () => (
    <div className="m-0 mx-4 text-nowrap d-flex flex-column align-items-center justify-content-center">
        <ClipboardArrowUpOutlineIcon data-tooltip="This changeset will be imported from the codehost." />
        <span className="text-muted">Import</span>
    </div>
)
