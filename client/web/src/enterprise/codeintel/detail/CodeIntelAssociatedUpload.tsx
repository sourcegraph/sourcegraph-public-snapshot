import classNames from 'classnames'
import ChevronRightIcon from 'mdi-react/ChevronRightIcon'
import React, { FunctionComponent } from 'react'

import { Link } from '@sourcegraph/shared/src/components/Link'

import { Timestamp } from '../../../components/time/Timestamp'
import { LsifIndexFields } from '../../../graphql-operations'
import { CodeIntelState } from '../shared/CodeIntelState'
import { CodeIntelUploadOrIndexLastActivity } from '../shared/CodeIntelUploadOrIndexLastActivity'

import styles from './CodeIntelAssociatedUpload.module.scss'

export interface CodeIntelAssociatedUploadProps {
    node: LsifIndexFields
    now?: () => Date
}

export const CodeIntelAssociatedUpload: FunctionComponent<CodeIntelAssociatedUploadProps> = ({ node, now }) =>
    node.associatedUpload && node.projectRoot ? (
        <>
            <div className="list-group position-relative">
                <div className={styles.grid}>
                    <span className={styles.separator} />

                    <div className={classNames(styles.information, 'd-flex flex-column')}>
                        <div className="m-0">
                            <h3 className="m-0 d-block d-md-inline">
                                This job uploaded an index{' '}
                                <Timestamp date={node.associatedUpload.uploadedAt} now={now} />
                            </h3>
                        </div>

                        <div>
                            <small className="text-mute">
                                <CodeIntelUploadOrIndexLastActivity
                                    node={{ ...node.associatedUpload, queuedAt: null }}
                                    now={now}
                                />
                            </small>
                        </div>
                    </div>

                    <span className={classNames(styles.state, 'd-none d-md-inline')}>
                        <CodeIntelState
                            node={node.associatedUpload}
                            className="d-flex flex-column align-items-center"
                        />
                    </span>
                    <span>
                        <Link
                            to={`/${node.projectRoot.repository.name}/-/settings/code-intelligence/uploads/${node.associatedUpload.id}`}
                        >
                            <ChevronRightIcon />
                        </Link>
                    </span>
                </div>
            </div>
        </>
    ) : (
        <></>
    )
