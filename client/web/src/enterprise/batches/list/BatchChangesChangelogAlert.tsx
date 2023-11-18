import React from 'react'

import classNames from 'classnames'

import { CardBody, Card, H3, H4, Link, Code, ProductStatusBadge } from '@sourcegraph/wildcard'

import { DismissibleAlert } from '../../../components/DismissibleAlert'

import styles from './BatchChangesListIntro.module.scss'

/**
 * CURRENT_VERSION is meant to be updated by release tooling to the current in-progress version.
 * Ie. After 5.0 is cut, this should be bumped to 5.1, and so on. This ensures we
 * always render the right changelog.
 */
const CURRENT_VERSION = '5.2'
/**
 * SHOW_CHANGELOG has to be set to true when a changelog entry is added for a release.
 * After every release, this will be set back to `false`. Chromatic will also verify
 * changes to this variable via visual regression testing.
 */
const SHOW_CHANGELOG = false

interface BatchChangesChangelogAlertProps {
    className?: string
    viewerIsAdmin: boolean
}

export const BatchChangesChangelogAlert: React.FunctionComponent<
    React.PropsWithChildren<BatchChangesChangelogAlertProps>
> = ({ className, viewerIsAdmin }) => {
    // IMPORTANT!! If you add an entry, make sure to set SHOW_CHANGELOG to true!
    if (!SHOW_CHANGELOG) {
        return null
    }
    return (
        <DismissibleAlert
            className={classNames(styles.batchChangesListIntroAlert, className)}
            partialStorageKey={`batch-changes-list-intro-changelog-${CURRENT_VERSION}`}
        >
            <Card className={classNames(styles.batchChangesListIntroCard, 'h-100')}>
                <CardBody>
                    <H4 as={H3}>Batch Changes updates in version {CURRENT_VERSION}</H4>
                    <ul className="mb-0 pl-3">
                        <li>
                            <ProductStatusBadge status="beta" className="mr-1" />
                            Batch Changes can now{' '}
                            <Link
                                rel="noopener"
                                to="/help/admin/config/batch_changes#commit-signing-for-github"
                                target="_blank"
                            >
                                sign commits
                            </Link>{' '}
                            created on GitHub via GitHub Apps.{' '}
                            {viewerIsAdmin ? (
                                <>
                                    {' '}
                                    Site admins can{' '}
                                    <Link to="/site-admin/batch-changes" target="_blank">
                                        configure a GitHub App integration
                                    </Link>{' '}
                                    to enable this feature.
                                </>
                            ) : (
                                <>
                                    GitHub App commit signing integrations can be configured by site admins and viewed
                                    from{' '}
                                    <Link to="/user/settings/batch-changes" target="_blank">
                                        your user settings
                                    </Link>
                                    .
                                </>
                            )}
                        </li>
                        <li>
                            {/* TODO: Add link to configuring credentials docs page once it's added. */}
                            Batch Changes now supports Gerrit. To start publishing Gerrit Changes, add your user
                            credentials.
                        </li>
                        <li>
                            Batch Changes now supports per-batch-change control for pushing to a fork of the upstream
                            repository with the batch spec property{' '}
                            <Link
                                rel="noopener"
                                to="/help/batch_changes/references/batch_spec_yaml_reference#changesettemplate-fork"
                                target="_blank"
                            >
                                <Code>changesetTemplate.fork</Code>
                            </Link>
                            .
                        </li>
                        <li>
                            Branches created by Batch Changes can now be automatically deleted on the code host upon
                            merging or closing a changeset by enabling the site config setting{' '}
                            <Link
                                rel="noopener"
                                to="/help/admin/config/batch_changes#automatically-delete-branches-on-merge-close"
                                target="_blank"
                            >
                                <Code>batchChanges.autoDeleteBranch</Code>
                            </Link>
                            .
                        </li>
                    </ul>
                </CardBody>
            </Card>
        </DismissibleAlert>
    )
}
