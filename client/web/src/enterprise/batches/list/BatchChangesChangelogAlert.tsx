import classNames from 'classnames'
import React from 'react'

import { DismissibleAlert } from '@sourcegraph/web/src/components/DismissibleAlert'

import styles from './BatchChangesListIntro.module.scss'

export const BatchChangesChangelogAlert: React.FunctionComponent = () => (
    <DismissibleAlert
        className={styles.batchChangesListIntroAlert}
        partialStorageKey="batch-changes-list-intro-changelog-3.35"
    >
        <div className={classNames(styles.batchChangesListIntroCard, 'card h-100 p-2')}>
            <div className="card-body">
                <h4>Batch Changes updates in version 3.35</h4>
                <ul className="mb-0 pl-3">
                    <li>
                        <a
                            href="https://docs.sourcegraph.com/batch_changes/references/batch_spec_yaml_reference#on-repository"
                            rel="noopener"
                            target="_blank"
                        >
                            Batch Changes now supports multiple base branches per repository.
                        </a>
                    </li>
                    <li>Moving a changeset from draft to open works again on GitLab.</li>
                </ul>
            </div>
        </div>
    </DismissibleAlert>
)
