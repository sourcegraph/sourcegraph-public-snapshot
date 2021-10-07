import classNames from 'classnames'
import React from 'react'

import { DismissibleAlert } from '@sourcegraph/web/src/components/DismissibleAlert'

import styles from './BatchChangesListIntro.module.scss'

export const BatchChangesChangelogAlert: React.FunctionComponent = () => (
    <DismissibleAlert
        className={styles.batchChangesListIntroAlert}
        partialStorageKey="batch-changes-list-intro-changelog-3.32"
    >
        <div className={classNames(styles.batchChangesListIntroCard, 'card h-100 p-2')}>
            <div className="card-body">
                <h4>New Batch Changes features in version 3.32</h4>
                <ul className="mb-0 pl-3">
                    <li>No changes.</li>
                </ul>
            </div>
        </div>
    </DismissibleAlert>
)
