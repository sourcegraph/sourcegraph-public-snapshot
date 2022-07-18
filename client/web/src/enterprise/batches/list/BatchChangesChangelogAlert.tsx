import React from 'react'

import classNames from 'classnames'

import { CardBody, Card, Link, Code, H3, H4 } from '@sourcegraph/wildcard'

import { DismissibleAlert } from '../../../components/DismissibleAlert'

import styles from './BatchChangesListIntro.module.scss'

export const BatchChangesChangelogAlert: React.FunctionComponent<React.PropsWithChildren<unknown>> = () => (
    <DismissibleAlert
        className={styles.batchChangesListIntroAlert}
        partialStorageKey="batch-changes-list-intro-changelog-3.41"
    >
        <Card className={classNames(styles.batchChangesListIntroCard, 'h-100')}>
            <CardBody>
                <H4 as={H3}>Batch Changes updates in version 3.41</H4>
                <ul className="mb-0 pl-3">
                    <li>
                        <Link to="/help/batch_changes/explanations/server_side" rel="noopener" target="_blank">
                            🚀 Running batch changes server-side
                        </Link>{' '}
                        is now in beta! In addition to using src-cli to run batch changes locally, you can now run them
                        server-side as well. This requires{' '}
                        <Link to="/help/admin/deploy_executors" rel="noopener" target="_blank">
                            installing executors.
                        </Link>
                        While running server-side unlocks a new and improved UI experience, you can still use src-cli
                        just like before.
                    </li>
                    <li>
                        <Link
                            to="/help/batch_changes/references/batch_spec_yaml_reference#steps-mount"
                            rel="noopener"
                            target="_blank"
                        >
                            src-cli now allows to mount local files into step containers.
                        </Link>
                    </li>
                    <li>It's now possible to start a batch change from search results.</li>
                    <li>
                        <Link to="/help/batch_changes/references/batch_spec_templating" rel="noopener" target="_blank">
                            You can now control where the link to the batch change in the changeset body is rendered
                            using {/* eslint-disable-next-line no-template-curly-in-string */}
                            <Code>{'${{ batch_change_link }}'}</Code>.
                        </Link>
                    </li>
                    <li>Viewing large batch changes is now significantly faster.</li>
                </ul>
            </CardBody>
        </Card>
    </DismissibleAlert>
)
