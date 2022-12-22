import React from 'react'

import classNames from 'classnames'

import { CardBody, Card, H3, H4, Link } from '@sourcegraph/wildcard'

import { DismissibleAlert } from '../../../components/DismissibleAlert'

import styles from './BatchChangesListIntro.module.scss'

export const BatchChangesChangelogAlert: React.FunctionComponent<React.PropsWithChildren<unknown>> = () => (
    <DismissibleAlert
        className={styles.batchChangesListIntroAlert}
        partialStorageKey="batch-changes-list-intro-changelog-4.4"
    >
        <Card className={classNames(styles.batchChangesListIntroCard, 'h-100')}>
            <CardBody>
                <H4 as={H3}>Batch Changes updates in version 4.4</H4>
                <ul className="mb-0 pl-3">
                    <li>
                        <Link to="/help/admin/deploy_executors#using-private-registries" rel="noopener" target="_blank">
                            Using private container registries
                        </Link>{' '}
                        is now supported in server-side batch changes.
                    </li>
                </ul>
            </CardBody>
        </Card>
    </DismissibleAlert>
)
