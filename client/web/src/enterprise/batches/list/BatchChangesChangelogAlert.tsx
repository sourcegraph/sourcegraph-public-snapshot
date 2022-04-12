import React from 'react'

import classNames from 'classnames'

import { CardBody, Card } from '@sourcegraph/wildcard'

import { DismissibleAlert } from '../../../components/DismissibleAlert'

import styles from './BatchChangesListIntro.module.scss'

export const BatchChangesChangelogAlert: React.FunctionComponent = () => (
    <DismissibleAlert
        className={styles.batchChangesListIntroAlert}
        partialStorageKey="batch-changes-list-intro-changelog-3.37"
    >
        <Card className={classNames(styles.batchChangesListIntroCard, 'h-100')}>
            <CardBody>
                <h4>Batch Changes updates in version 3.37</h4>
                <ul className="mb-0 pl-3">
                    <li>Nothing noteworthy this time.</li>
                </ul>
            </CardBody>
        </Card>
    </DismissibleAlert>
)
