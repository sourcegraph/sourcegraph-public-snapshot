import classNames from 'classnames'
import React from 'react'

import { DismissibleAlert } from '@sourcegraph/web/src/components/DismissibleAlert'
import { CardBody, Card, Link } from '@sourcegraph/wildcard'

import styles from './BatchChangesListIntro.module.scss'

export const BatchChangesChangelogAlert: React.FunctionComponent = () => (
    <DismissibleAlert
        className={styles.batchChangesListIntroAlert}
        partialStorageKey="batch-changes-list-intro-changelog-3.36"
    >
        <Card className={classNames(styles.batchChangesListIntroCard, 'h-100')}>
            <CardBody>
                <h4>Batch Changes updates in version 3.36</h4>
                <ul className="mb-0 pl-3">
                    <li>
                        <Link
                            to="https://docs.sourcegraph.com/admin/config/batch_changes#forks"
                            rel="noopener"
                            target="_blank"
                        >
                            Batch Changes now supports pushing changesets to forked repositories.
                        </Link>
                    </li>
                </ul>
            </CardBody>
        </Card>
    </DismissibleAlert>
)
