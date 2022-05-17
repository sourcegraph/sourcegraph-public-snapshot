import { FunctionComponent } from 'react'

import classNames from 'classnames'
import ChevronRightIcon from 'mdi-react/ChevronRightIcon'

import { Link, Typography } from '@sourcegraph/wildcard'

import { LsifUploadFields } from '../../../../graphql-operations'
import { CodeIntelState } from '../../shared/components/CodeIntelState'
import { CodeIntelUploadOrIndexCommit } from '../../shared/components/CodeIntelUploadOrIndexCommit'
import { CodeIntelUploadOrIndexRepository } from '../../shared/components/CodeIntelUploadOrIndexerRepository'
import { CodeIntelUploadOrIndexIndexer } from '../../shared/components/CodeIntelUploadOrIndexIndexer'
import { CodeIntelUploadOrIndexLastActivity } from '../../shared/components/CodeIntelUploadOrIndexLastActivity'
import { CodeIntelUploadOrIndexRoot } from '../../shared/components/CodeIntelUploadOrIndexRoot'

import styles from './CodeIntelUploadNode.module.scss'

export interface CodeIntelUploadNodeProps {
    node: LsifUploadFields
    now?: () => Date
}

export const CodeIntelUploadNode: FunctionComponent<React.PropsWithChildren<CodeIntelUploadNodeProps>> = ({
    node,
    now,
}) => (
    <>
        <span className={styles.separator} />

        <div className={classNames(styles.information, 'd-flex flex-column')}>
            <div className="m-0">
                <Typography.H3 className="m-0 d-block d-md-inline">
                    <CodeIntelUploadOrIndexRepository node={node} />
                </Typography.H3>
            </div>

            <div>
                <span className="mr-2 d-block d-mdinline-block">
                    Directory <CodeIntelUploadOrIndexRoot node={node} /> indexed at commit{' '}
                    <CodeIntelUploadOrIndexCommit node={node} /> by <CodeIntelUploadOrIndexIndexer node={node} />
                </span>

                <small className="text-mute">
                    <CodeIntelUploadOrIndexLastActivity node={{ ...node, queuedAt: null }} now={now} />
                </small>
            </div>
        </div>

        <span className={classNames(styles.state, 'd-none d-md-inline')}>
            <CodeIntelState node={node} className="d-flex flex-column align-items-center" />
        </span>
        <span>
            <Link to={`./uploads/${node.id}`}>
                <ChevronRightIcon />
            </Link>
        </span>
    </>
)
