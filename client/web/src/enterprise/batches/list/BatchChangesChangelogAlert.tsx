import React from 'react'

import classNames from 'classnames'

import { CardBody, Card, Typography } from '@sourcegraph/wildcard'

import { DismissibleAlert } from '../../../components/DismissibleAlert'

import styles from './BatchChangesListIntro.module.scss'

export const BatchChangesChangelogAlert: React.FunctionComponent<React.PropsWithChildren<unknown>> = () => (
    <DismissibleAlert
        className={styles.batchChangesListIntroAlert}
        partialStorageKey="batch-changes-list-intro-changelog-3.39"
    >
        <Card className={classNames(styles.batchChangesListIntroCard, 'h-100')}>
            <CardBody>
                <Typography.H4 as={Typography.H2}>Batch Changes updates in version 3.39</Typography.H4>
                <ul className="mb-0 pl-3">
                    <li>
                        Bulk actions are now visible regardless of filtering. Previously, you had to filter by status to
                        see them.
                    </li>
                </ul>
            </CardBody>
        </Card>
    </DismissibleAlert>
)
