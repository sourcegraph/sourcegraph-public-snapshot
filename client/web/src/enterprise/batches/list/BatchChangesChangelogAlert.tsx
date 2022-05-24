import React from 'react'

import classNames from 'classnames'

import { CardBody, Card, Typography, Link } from '@sourcegraph/wildcard'

import { DismissibleAlert } from '../../../components/DismissibleAlert'

import styles from './BatchChangesListIntro.module.scss'

export const BatchChangesChangelogAlert: React.FunctionComponent<React.PropsWithChildren<unknown>> = () => (
    <DismissibleAlert
        className={styles.batchChangesListIntroAlert}
        partialStorageKey="batch-changes-list-intro-changelog-3.40"
    >
        <Card className={classNames(styles.batchChangesListIntroCard, 'h-100')}>
            <CardBody>
                <Typography.H4 as={Typography.H2}>Batch Changes updates in version 3.40</Typography.H4>
                <ul className="mb-0 pl-3">
                    <li>
                        <Link
                            to="https://docs.sourcegraph.com/batch_changes/explanations/introduction_to_batch_changes#supported-code-hosts-and-changeset-types"
                            rel="noopener"
                            target="_blank"
                        >
                            Bitbucket Cloud is now supported with Batch Changes.
                        </Link>
                    </li>
                </ul>
            </CardBody>
        </Card>
    </DismissibleAlert>
)
