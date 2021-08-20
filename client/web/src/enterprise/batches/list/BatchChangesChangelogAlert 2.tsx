import classNames from 'classnames'
import React from 'react'

import { DismissibleAlert } from '@sourcegraph/web/src/components/DismissibleAlert'

import styles from './BatchChangesListIntro.module.scss'

export const BatchChangesChangelogAlert: React.FunctionComponent = () => (
    <DismissibleAlert
        className={styles.batchChangesListIntroAlert}
        partialStorageKey="batch-changes-list-intro-changelog-3.30"
    >
        <div className={classNames(styles.batchChangesListIntroCard, 'card h-100 p-2')}>
            <div className="card-body">
                <h4>New Batch Changes features in version 3.30</h4>
                <ul className="mb-0 pl-3">
                    <li>
                        Changesets can now be published from the UI!{' '}
                        <a
                            href="https://docs.sourcegraph.com/batch_changes/how-tos/publishing_changesets#within-the-ui"
                            rel="noopener"
                            target="_blank"
                        >
                            Learn more.
                        </a>
                    </li>
                    <li>
                        You can now view changesets created in a specific repository from that repository's landing
                        page.
                    </li>
                </ul>
            </div>
        </div>
    </DismissibleAlert>
)
