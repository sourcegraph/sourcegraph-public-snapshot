import classNames from 'classnames'
import InformationOutlineIcon from 'mdi-react/InformationOutlineIcon'
import React, { FunctionComponent } from 'react'

import { pluralize } from '@sourcegraph/common'
import { Link } from '@sourcegraph/wildcard/src/components/Link'

import { NormalizedUploadRetentionMatch } from '../hooks/queryUploadRetentionMatches'

import styles from './DependencyOrDependentNode.module.scss'

export interface RetentionMatchNodeProps {
    node: NormalizedUploadRetentionMatch
}

export const retentionByUploadTitle = 'Retention by reference'

export const RetentionMatchNode: FunctionComponent<RetentionMatchNodeProps> = ({ node }) => {
    if (node.matchType === 'RetentionPolicy') {
        const joinedRevhashes =
            node.protectingCommits.length !== 0 ? (
                <>
                    , by visible {pluralize('commit', node.protectingCommits.length)}{' '}
                    {node.protectingCommits.map(hash => hash.slice(0, 9)).join(', ')}
                    <InformationOutlineIcon
                        className="ml-1 icon-inline"
                        data-tooltip="This upload is retained to service code-intel queries for commit(s) with applicable retention policies."
                    />
                </>
            ) : (
                <></>
            )
        return (
            <>
                <span className={styles.separator} />

                <div className={classNames(styles.information, 'd-flex flex-column')}>
                    <div className="m-0">
                        <Link to={`../configuration/${node.configurationPolicy.id}`} className="p-0">
                            <h3 className="m-0 d-block d-md-inline">{node.configurationPolicy.name}</h3>
                        </Link>
                        <div className="mr-2 d-block d-mdinline-block">
                            Matched: {node.matches ? 'yes' : 'no'}
                            {joinedRevhashes}
                        </div>
                    </div>
                </div>
            </>
        )
    }
    if (node.matchType === 'UploadReference') {
        return (
            <>
                <span className={styles.separator} />

                <div className={classNames(styles.information, 'd-flex flex-column')}>
                    <div className="m-0">
                        <h3 className="m-0 d-block d-md-inline">{retentionByUploadTitle}</h3>
                        <div className="mr-2 d-block d-mdinline-block">
                            Referenced by {node.total} {pluralize('upload', node.total, 'uploads')}, including{' '}
                            {node.uploadSlice
                                .slice(0, 3)
                                .map<React.ReactNode>(upload => (
                                    <Link key={upload.id} to={`/site-admin/code-intelligence/uploads/${upload.id}`}>
                                        {upload.projectRoot?.repository.name ?? 'unknown'}
                                    </Link>
                                ))
                                .reduce((previous, current) => [previous, ', ', current])}
                            <InformationOutlineIcon
                                className='ml-1 icon-inline'
                                data-tooltip="Uploads that are dependencies of other upload(s) are retained to service cross-repository code-intel queries."
                            />
                        </div>
                    </div>
                </div>
            </>
        )
    }

    throw new Error(`invalid node type ${JSON.stringify(node as object)}`)
}
