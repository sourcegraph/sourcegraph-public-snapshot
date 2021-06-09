import classNames from 'classnames'
import WarningIcon from 'mdi-react/WarningIcon'
import React from 'react'
import { Link } from 'react-router-dom'

import { SourcegraphIcon } from '../../../auth/icons'
import { DismissibleAlert } from '../../../components/DismissibleAlert'

import styles from './BatchChangesListIntro.module.scss'

export interface BatchChangesListIntroProps {
    licensed: boolean | undefined
}

export const BatchChangesListIntro: React.FunctionComponent<BatchChangesListIntroProps> = ({ licensed }) => {
    if (licensed === undefined) {
        return null
    }
    return (
        <div className="row">
            {licensed === true ? (
                <div className="col-12">
                    <BatchChangesChangelogAlert />
                </div>
            ) : (
                <>
                    <div className="col-12 col-md-6 mb-3">
                        <BatchChangesUnlicensedAlert />
                    </div>
                    <div className="col-12 col-md-6 mb-3">
                        <BatchChangesChangelogAlert />
                    </div>
                </>
            )}
        </div>
    )
}

const BatchChangesChangelogAlert: React.FunctionComponent = () => (
    <DismissibleAlert
        className={styles.batchChangesListIntroAlert}
        partialStorageKey="batch-changes-list-intro-changelog-3.28"
    >
        <div className={classNames(styles.batchChangesListIntroCard, 'card h-100 p-2')}>
            <div className="card-body">
                <h4>New Batch Changes features in version 3.28</h4>
                <ul className="mb-0 pl-3">
                    <li>
                        <WarningIcon className="icon-inline text-warning" /> <strong>Deprecation:</strong> Starting with
                        Sourcegraph 3.29, we will stop using code host connection tokens for creating changesets. If a
                        site-admin on your instance relied on the global configuration, please ask them to go add global
                        credentials for Batch Changes in the <Link to="/site-admin/batch-changes">admin UI</Link> for
                        uninterrupted Batch Changes usage.
                    </li>
                </ul>
                <ul className="mb-0 pl-3">
                    <li>
                        Commenting on changesets{' '}
                        <a
                            target="_blank"
                            rel="noopener noreferrer"
                            href="https://docs.sourcegraph.com/batch_changes/how-tos/bulk_operations_on_changesets"
                        >
                            is now supported
                        </a>
                        .
                    </li>
                </ul>
                <ul className="mb-0 pl-3">
                    <li>
                        Steps in batch specs can be run conditionally using{' '}
                        <a
                            target="_blank"
                            rel="noopener noreferrer"
                            href="https://docs.sourcegraph.com/batch_changes/references/batch_spec_yaml_reference#steps-if"
                        >
                            the `if:` property
                        </a>
                        .
                    </li>
                </ul>
                <ul className="mb-0 pl-3">
                    <li>
                        User and site credentials can be encrypted in the database by adding a key to{' '}
                        <a
                            target="_blank"
                            rel="noopener noreferrer"
                            href="https://docs.sourcegraph.com/admin/config/encryption"
                        >
                            the `batchChangesCredentialKey` property
                        </a>{' '}
                        of `encryption.keys` in the site configuration.
                    </li>
                </ul>
            </div>
        </div>
    </DismissibleAlert>
)

const BatchChangesUnlicensedAlert: React.FunctionComponent = () => (
    <div className={classNames(styles.batchChangesListIntroAlert, 'h-100')}>
        <div className={classNames(styles.batchChangesListIntroCard, 'card p-2 h-100')}>
            <div className="card-body d-flex align-items-start">
                {/* d-none d-sm-block ensure that we hide the icon on XS displays. */}
                <SourcegraphIcon className="mr-3 col-2 mt-2 d-none d-sm-block" />
                <div>
                    <h4>Batch changes trial</h4>
                    <p>
                        Batch changes is a paid feature of Sourcegraph. All users can create sample batch changes with
                        up to five changesets without a license.
                    </p>
                    <p className="mb-0">
                        <a href="https://about.sourcegraph.com/contact/sales/">Contact sales</a> to obtain a trial
                        license.
                    </p>
                </div>
            </div>
        </div>
    </div>
)
