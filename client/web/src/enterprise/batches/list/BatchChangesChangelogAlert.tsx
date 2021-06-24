import classNames from 'classnames'
import React from 'react'

import { DismissibleAlert } from '@sourcegraph/web/src/components/DismissibleAlert'

import styles from './BatchChangesListIntro.module.scss'

export const BatchChangesChangelogAlert: React.FunctionComponent = () => (
    <DismissibleAlert
        className={styles.batchChangesListIntroAlert}
        partialStorageKey="batch-changes-list-intro-changelog-3.29"
    >
        <div className={classNames(styles.batchChangesListIntroCard, 'card h-100 p-2')}>
            <div className="card-body">
                <h4>New Batch Changes features in version 3.29</h4>
                <ul className="mb-0 pl-3">
                    <li>New bulk operations have been added to retry or merge multiple changesets at once. &#x2705;</li>
                    <li>
                        <a target="_blank" rel="noopener noreferrer" href="https://github.com/sourcegraph/src-cli">
                            <code>src</code>
                        </a>{' '}
                        can now cache the result of each step when executing a batch spec, rather than only caching the
                        final result. &#x1f680;
                    </li>
                    <li>
                        Changeset specs on batch changes that have been previewed but not applied now expire
                        consistently after one week. &#x1f5d3;&#xfe0f;
                    </li>
                </ul>
            </div>
        </div>
    </DismissibleAlert>
)
