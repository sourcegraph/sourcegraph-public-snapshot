import React from 'react'

import classNames from 'classnames'

import { CardBody, Card, H3, H4 } from '@sourcegraph/wildcard'

import { DismissibleAlert } from '../../../components/DismissibleAlert'

import styles from './BatchChangesListIntro.module.scss'

export const BatchChangesChangelogAlert: React.FunctionComponent<React.PropsWithChildren<unknown>> = () => (
    <DismissibleAlert
        className={styles.batchChangesListIntroAlert}
        partialStorageKey="batch-changes-list-intro-changelog-4.0"
    >
        <Card className={classNames(styles.batchChangesListIntroCard, 'h-100')}>
            <CardBody>
                <H4 as={H3}>Batch Changes updates in version 4.0</H4>
                <ul className="mb-0 pl-3">
                    <li>
                        Code host connection tokens are no longer supported as a fallback method for syncing changesets
                        in Batch Changes.
                    </li>
                </ul>
            </CardBody>
        </Card>
    </DismissibleAlert>
)
