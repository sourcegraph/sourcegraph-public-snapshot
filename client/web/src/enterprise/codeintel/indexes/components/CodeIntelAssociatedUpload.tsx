import { FunctionComponent } from 'react'

import classNames from 'classnames'
import ChevronRightIcon from 'mdi-react/ChevronRightIcon'

import { Link, Typography } from '@sourcegraph/wildcard'

import { Timestamp } from '../../../../components/time/Timestamp'
import { LsifIndexFields } from '../../../../graphql-operations'
import { CodeIntelState } from '../../shared/components/CodeIntelState'
import { CodeIntelUploadOrIndexLastActivity } from '../../shared/components/CodeIntelUploadOrIndexLastActivity'

import styles from './CodeIntelAssociatedUpload.module.scss'

export interface CodeIntelAssociatedUploadProps {
    node: LsifIndexFields
    now?: () => Date
}

export const CodeIntelAssociatedUpload: FunctionComponent<React.PropsWithChildren<CodeIntelAssociatedUploadProps>> = ({
    node,
    now,
}) =>
    node.associatedUpload && node.projectRoot ? (
        <>
            <div className="list-group position-relative">
                <div className={styles.grid}>
                    <span className={styles.separator} />

                    <div className={classNames(styles.information, 'd-flex flex-column')}>
                        <div className="m-0">
                            <Typography.H3 className="m-0 d-block d-md-inline">
                                This job performed an upload{' '}
                                <Timestamp date={node.associatedUpload.uploadedAt} now={now} />
                            </Typography.H3>
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
                            to={`/${node.projectRoot.repository.name}/-/code-intelligence/uploads/${node.associatedUpload.id}`}
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
