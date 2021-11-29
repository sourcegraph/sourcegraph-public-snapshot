import classNames from 'classnames'
import React from 'react'

import { DismissibleAlert } from '@sourcegraph/web/src/components/DismissibleAlert'

import styles from './BatchChangesListIntro.module.scss'

export const BatchChangesChangelogAlert: React.FunctionComponent = () => (
    <DismissibleAlert
        className={styles.batchChangesListIntroAlert}
        partialStorageKey="batch-changes-list-intro-changelog-3.34"
    >
        <div className={classNames(styles.batchChangesListIntroCard, 'card h-100 p-2')}>
            <div className="card-body">
                <h4>Batch Changes updates in version 3.34</h4>
                <ul className="mb-0 pl-3">
                    <li>
                        <a
                            href="https://docs.sourcegraph.com/batch_changes/references/name-change"
                            rel="noopener"
                            target="_blank"
                        >
                            The deprecated campaigns APIs have been removed.
                        </a>
                    </li>
                    <li>
                        <a
                            href="https://docs.sourcegraph.com/batch_changes/references/name-change"
                            rel="noopener"
                            target="_blank"
                        >
                            The deprecated <code>campaigns.enabled</code> and <code>campaigns.restrictToAdmins</code>{' '}
                            site settings no longer have any effect.
                        </a>
                    </li>
                    <li>
                        <a
                            href="https://docs.sourcegraph.com/batch_changes/references/name-change"
                            rel="noopener"
                            target="_blank"
                        >
                            The deprecated <code>src campaign</code> command has been removed from <code>src-cli</code>.
                        </a>
                    </li>
                    <li>
                        <a
                            href="https://docs.sourcegraph.com/batch_changes/references/name-change"
                            rel="noopener"
                            target="_blank"
                        >
                            The deprecated campaigns URLs have been removed.
                        </a>
                    </li>
                </ul>
            </div>
        </div>
    </DismissibleAlert>
)
