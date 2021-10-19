import classNames from 'classnames'
import React from 'react'

import { SourcegraphIcon } from '../../../auth/icons'

import { BatchChangesChangelogAlert } from './BatchChangesChangelogAlert'
import styles from './BatchChangesListIntro.module.scss'

export interface BatchChangesListIntroProps {
    licensed: boolean | undefined
}

export const BatchChangesListIntro: React.FunctionComponent<BatchChangesListIntroProps> = ({ licensed }) => {
    if (licensed === undefined) {
        return null
    }
    return (
        <div className="row">
            {licensed === true ? (
                <div className="col-12">
                    <BatchChangesChangelogAlert />
                </div>
            ) : (
                <>
                    <div className="col-12 col-md-6 mb-3">
                        <BatchChangesUnlicensedAlert />
                    </div>
                    <div className="col-12 col-md-6 mb-3">
                        <BatchChangesChangelogAlert />
                    </div>
                </>
            )}
        </div>
    )
}

const BatchChangesUnlicensedAlert: React.FunctionComponent = () => (
    <div className={classNames(styles.batchChangesListIntroAlert, 'h-100')}>
        <div className={classNames(styles.batchChangesListIntroCard, 'card p-2 h-100')}>
            <div className="card-body d-flex align-items-start">
                {/* d-none d-sm-block ensure that we hide the icon on XS displays. */}
                <SourcegraphIcon className="mr-3 col-2 mt-2 d-none d-sm-block" />
                <div>
                    <h4>Batch changes trial</h4>
                    <p>
                        Batch changes is a paid feature of Sourcegraph. All users can create sample batch changes with
                        up to five changesets without a license.
                    </p>
                    <p className="mb-0">
                        <a href="https://about.sourcegraph.com/contact/sales/">Contact sales</a> to obtain a trial
                        license.
                    </p>
                </div>
            </div>
        </div>
    </div>
)
