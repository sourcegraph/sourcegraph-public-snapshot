import React from 'react'

import classNames from 'classnames'

import { CardBody, Card, Link, H4, Text } from '@sourcegraph/wildcard'

import { SourcegraphIcon } from '../../../auth/icons'

import styles from './BatchChangesListIntro.module.scss'

export interface BatchChangesListIntroProps {
    isLicensed: boolean | undefined
}

export const BatchChangesListIntro: React.FunctionComponent<React.PropsWithChildren<BatchChangesListIntroProps>> = ({
    isLicensed,
}) => {
    if (isLicensed === undefined || isLicensed === true) {
        return null
    }

    return (
        <div className="row">
            {/* {isLicensed === true ? (
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
            )} */}

            <div className="col-12">
                <BatchChangesUnlicensedAlert />
            </div>
        </div>
    )
}

const BatchChangesUnlicensedAlert: React.FunctionComponent<React.PropsWithChildren<unknown>> = () => (
    <div className={classNames(styles.batchChangesListIntroAlert, 'h-100')}>
        <Card className={classNames(styles.batchChangesListIntroCard, 'h-100')}>
            <CardBody className="d-flex align-items-start">
                {/* d-none d-sm-block ensure that we hide the icon on XS displays. */}
                <SourcegraphIcon className="mr-3 col-2 mt-2 d-none d-sm-block" />
                <div>
                    <H4>Batch changes trial</H4>
                    <Text>
                        Batch changes is a paid feature of Sourcegraph. All users can create sample batch changes with
                        up to five changesets without a license.
                    </Text>
                    <Text className="mb-0">
                        <Link to="https://about.sourcegraph.com/contact/sales/">Contact sales</Link> to obtain a trial
                        license.
                    </Text>
                </div>
            </CardBody>
        </Card>
    </div>
)
