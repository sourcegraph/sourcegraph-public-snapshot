import classNames from 'classnames'
import React from 'react'

import { CardBody, Card, Link } from '@sourcegraph/wildcard'

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
        <Card className={classNames(styles.batchChangesListIntroCard, 'h-100')}>
            <CardBody className="d-flex align-items-start">
                {/* d-none d-sm-block ensure that we hide the icon on XS displays. */}
                <SourcegraphIcon className="mr-3 col-2 mt-2 d-none d-sm-block" />
                <div>
                    <h4>Batch changes trial</h4>
                    <p>
                        Batch changes is a paid feature of Sourcegraph. All users can create sample batch changes with
                        up to five changesets without a license.
                    </p>
                    <p className="mb-0">
                        <Link to="https://about.sourcegraph.com/contact/sales/">Contact sales</Link> to obtain a trial
                        license.
                    </p>
                </div>
            </CardBody>
        </Card>
    </div>
)
