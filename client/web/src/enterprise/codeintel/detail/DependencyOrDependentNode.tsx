import classNames from 'classnames'
import ChevronRightIcon from 'mdi-react/ChevronRightIcon'
import React, { FunctionComponent } from 'react'
import { Link } from 'react-router-dom'

import { LsifUploadFields } from '../../../graphql-operations'
import { CodeIntelState } from '../shared/CodeIntelState'
import { CodeIntelUploadOrIndexCommit } from '../shared/CodeIntelUploadOrIndexCommit'
import { CodeIntelUploadOrIndexRepository } from '../shared/CodeIntelUploadOrIndexerRepository'
import { CodeIntelUploadOrIndexIndexer } from '../shared/CodeIntelUploadOrIndexIndexer'
import { CodeIntelUploadOrIndexRoot } from '../shared/CodeIntelUploadOrIndexRoot'

import styles from './DependencyOrDependentNode.module.scss'

export interface DependencyOrDependentNodeProps {
    node: LsifUploadFields
    now?: () => Date
}

export const DependencyOrDependentNode: FunctionComponent<DependencyOrDependentNodeProps> = ({ node }) => (
    <>
        <span className={styles.separator} />

        <div className={classNames(styles.information, 'd-flex flex-column')}>
            <div className="m-0">
                <h3 className="m-0 d-block d-md-inline">
                    <CodeIntelUploadOrIndexRepository node={node} />
                </h3>
            </div>

            <div>
                <span className="mr-2 d-block d-mdinline-block">
                    Directory <CodeIntelUploadOrIndexRoot node={node} /> indexed at commit{' '}
                    <CodeIntelUploadOrIndexCommit node={node} /> by <CodeIntelUploadOrIndexIndexer node={node} />
                </span>
            </div>
        </div>

        <span className={classNames(styles.state, 'd-none d-md-inline')}>
            <CodeIntelState node={node} className="d-flex flex-column align-items-center" />
        </span>
        <span>
            <Link to={`./${node.id}`}>
                <ChevronRightIcon />
            </Link>
        </span>
    </>
)
