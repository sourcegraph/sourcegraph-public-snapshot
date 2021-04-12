import React from 'react'

import { SourcegraphIcon } from '../../../auth/icons'
import { DismissibleAlert } from '../../../components/DismissibleAlert'

export interface BatchChangesListIntroProps {
    licensed: boolean | undefined
}

export const BatchChangesListIntro: React.FunctionComponent<BatchChangesListIntroProps> = ({ licensed }) => (
    <>
        <div className="row">
            <div className="col-12">
                <BatchChangesRenameAlert />
            </div>
        </div>
        <div className="row mb-2">
            {licensed === true ? (
                <div className="col-12">
                    <BatchChangesChangelogAlert />
                </div>
            ) : (
                <>
                    {licensed === false && (
                        <>
                            <div className="col-12 col-md-6">
                                <BatchChangesUnlicensedAlert />
                            </div>
                            <div className="col-12 col-md-6">
                                <BatchChangesChangelogAlert />
                            </div>
                        </>
                    )}
                </>
            )}
        </div>
    </>
)

const BatchChangesChangelogAlert: React.FunctionComponent = () => (
    <DismissibleAlert
        className="batch-changes-list-intro__alert"
        partialStorageKey="batch-changes-list-intro-changelog-3.26"
    >
        <div className="batch-changes-list-intro__card card h-100 p-2">
            <div className="card-body">
                <h4>New Batch Changes features in version 3.26</h4>
                <ul className="text-muted mb-0 pl-3">
                    <li>Batch Changes now supports SSH cloned repos. Users can configure SSH access in settings.</li>
                </ul>
                <ul className="text-muted mb-0 pl-3">
                    <li>
                        Burndown charts have been improved: changeset progress is now shown with greater resolution
                        across the entire lifespan of the batch change.
                    </li>
                </ul>
            </div>
        </div>
    </DismissibleAlert>
)

const BatchChangesRenameAlert: React.FunctionComponent = () => (
    // Unlike the other alerts in this file, the spacing below the alert is
    // handled as padding within the element here to avoid extra margin being
    // included once this alert is dismissed. (The other alerts in this file
    // have margin that is used for more than just the alerts. Structural
    // margin?)
    <DismissibleAlert
        className="batch-changes-list-intro__alert pb-4"
        partialStorageKey="batch-changes-list-intro-rename"
    >
        <div className="batch-changes-list-intro__card card h-100 p-2">
            <div className="card-body">
                <h4>Campaigns is now Batch Changes</h4>
                <p className="text-muted mb-0">
                    Campaigns was renamed to Sourcegraph Batch Changes in version 3.26. If you were already using it
                    under the previous name (campaigns), backwards compatibility has been preserved.{' '}
                    <a href="https://docs.sourcegraph.com/batch_changes/references/name-change">Read more.</a>
                </p>
            </div>
        </div>
    </DismissibleAlert>
)

const BatchChangesUnlicensedAlert: React.FunctionComponent = () => (
    <div className="batch-change-list-intro__alert">
        <div className="batch-change-list-intro__card card p-2 h-100">
            <div className="card-body d-flex align-items-start">
                {/* d-none d-sm-block ensure that we hide the icon on XS displays. */}
                <SourcegraphIcon className="mr-3 col-2 mt-2 d-none d-sm-block" />
                <div>
                    <h4>Batch changes trial</h4>
                    <p className="text-muted">
                        Batch changes is a paid feature of Sourcegraph. All users can create sample batch changes with
                        up to five changesets without a license.
                    </p>
                    <p className="text-muted mb-0">
                        <a href="https://about.sourcegraph.com/contact/sales/">Contact sales</a> to obtain a trial
                        license.
                    </p>
                </div>
            </div>
        </div>
    </div>
)
