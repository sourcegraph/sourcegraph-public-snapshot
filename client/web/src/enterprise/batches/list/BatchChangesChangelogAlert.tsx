import classNames from 'classnames'
import React from 'react'

import { DismissibleAlert } from '@sourcegraph/web/src/components/DismissibleAlert'

import styles from './BatchChangesListIntro.module.scss'

export const BatchChangesChangelogAlert: React.FunctionComponent = () => (
    <DismissibleAlert
        className={styles.batchChangesListIntroAlert}
        partialStorageKey="batch-changes-list-intro-changelog-3.33"
    >
        <div className={classNames(styles.batchChangesListIntroCard, 'card h-100 p-2')}>
            <div className="card-body">
                <h4>Batch Changes updates in version 3.33</h4>
                <ul className="mb-0 pl-3">
                    <li>
                        <a
                            href="https://docs.sourcegraph.com/batch_changes/references/name-change"
                            rel="noopener"
                            target="_blank"
                        >
                            The deprecated campaigns APIs will be removed in the next release.
                        </a>
                    </li>
                    <li>
                        <a
                            href="https://docs.sourcegraph.com/batch_changes/references/name-change"
                            rel="noopener"
                            target="_blank"
                        >
                            The deprecated <code>campaigns.enabled</code> and <code>campaigns.restrictToAdmins</code>{' '}
                            site settings will be non-functional in the next release.
                        </a>
                    </li>
                    <li>
                        <a
                            href="https://docs.sourcegraph.com/batch_changes/references/name-change"
                            rel="noopener"
                            target="_blank"
                        >
                            The deprecated <code>src campaign</code> command will be removed in the next release of{' '}
                            <code>src-cli</code>.
                        </a>
                    </li>
                    <li>
                        <a
                            href="https://docs.sourcegraph.com/batch_changes/references/name-change"
                            rel="noopener"
                            target="_blank"
                        >
                            The deprecated campaigns URLs will be removed in the next release.
                        </a>
                    </li>
                </ul>
            </div>
        </div>
    </DismissibleAlert>
)
