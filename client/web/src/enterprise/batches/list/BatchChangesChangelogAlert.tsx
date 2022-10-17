import React from 'react'

import classNames from 'classnames'

import { CardBody, Card, H3, H4, Link } from '@sourcegraph/wildcard'

import { DismissibleAlert } from '../../../components/DismissibleAlert'

import styles from './BatchChangesListIntro.module.scss'

export const BatchChangesChangelogAlert: React.FunctionComponent<React.PropsWithChildren<unknown>> = () => (
    <DismissibleAlert
        className={styles.batchChangesListIntroAlert}
        partialStorageKey="batch-changes-list-intro-changelog-4.1"
    >
        <Card className={classNames(styles.batchChangesListIntroCard, 'h-100')}>
            <CardBody>
                <H4 as={H3}>Batch Changes updates in version 4.1</H4>
                <ul className="mb-0 pl-3">
                    <li>
                        Batch changes run on the server can now be created within{' '}
                        <Link
                            to="/help/batch_changes/how-tos/creating_a_batch_change#creating-a-batch-change-in-a-different-namespace"
                            rel="noopener"
                            target="_blank"
                        >
                            organizations
                        </Link>
                        .
                    </li>
                    <li>
                        Batch changes run on the server now support{' '}
                        <Link to="/help/batch_changes/how-tos/server_side_file_mounts" rel="noopener" target="_blank">
                            mounting files
                        </Link>
                        .
                    </li>
                </ul>
            </CardBody>
        </Card>
    </DismissibleAlert>
)
