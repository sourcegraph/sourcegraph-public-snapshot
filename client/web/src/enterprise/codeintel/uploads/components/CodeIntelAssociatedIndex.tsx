import { FunctionComponent } from 'react'

import classNames from 'classnames'
import ChevronRightIcon from 'mdi-react/ChevronRightIcon'

import { Link, Typography } from '@sourcegraph/wildcard'

import { LsifUploadFields } from '../../../../graphql-operations'
import { CodeIntelState } from '../../shared/components/CodeIntelState'
import { CodeIntelUploadOrIndexLastActivity } from '../../shared/components/CodeIntelUploadOrIndexLastActivity'

import styles from './CodeIntelAssociatedIndex.module.scss'

export interface CodeIntelAssociatedIndexProps {
    node: LsifUploadFields
    now?: () => Date
}

export const CodeIntelAssociatedIndex: FunctionComponent<React.PropsWithChildren<CodeIntelAssociatedIndexProps>> = ({
    node,
    now,
}) =>
    node.associatedIndex && node.projectRoot ? (
        <>
            <div className="list-group position-relative">
                <div className={classNames(styles.grid, 'mb-3')}>
                    <div className={classNames(styles.information, 'd-flex flex-column')}>
                        <div className="m-0">
                            <Typography.H3 className="m-0 d-block d-md-inline">
                                This upload was created by an auto-indexing job
                            </Typography.H3>
                        </div>

                        <div>
                            <small className="text-mute">
                                <CodeIntelUploadOrIndexLastActivity
                                    node={{ ...node.associatedIndex, uploadedAt: null }}
                                    now={now}
                                />
                            </small>
                        </div>
                    </div>

                    <span className={classNames(styles.state, 'd-none d-md-inline')}>
                        <CodeIntelState node={node.associatedIndex} className="d-flex flex-column align-items-center" />
                    </span>
                    <span>
                        <Link
                            to={`/${node.projectRoot.repository.name}/-/code-intelligence/indexes/${node.associatedIndex.id}`}
                        >
                            <ChevronRightIcon />
                        </Link>
                    </span>

                    <span className={styles.separator} />
                </div>
            </div>
        </>
    ) : (
        <></>
    )
