import classNames from 'classnames'
import InformationOutlineIcon from 'mdi-react/InformationOutlineIcon'
import React, { FunctionComponent } from 'react'

import { pluralize } from '@sourcegraph/common'
import { Link } from '@sourcegraph/wildcard/src/components/Link'

import {
    NormalizedUploadRetentionMatch,
    RetentionPolicyMatch,
    UploadReferenceMatch,
} from '../hooks/queryUploadRetentionMatches'

import styles from './DependencyOrDependentNode.module.scss'

export interface RetentionMatchNodeProps {
    node: NormalizedUploadRetentionMatch
}

export const retentionByUploadTitle = 'Retention by reference'
export const retentionByBranchTipTitle = 'Retention by tip of default branch'

export const RetentionMatchNode: FunctionComponent<RetentionMatchNodeProps> = ({ node }) => {
    if (node.matchType === 'RetentionPolicy') {
        return <RetentionPolicyRetentionMatchNode match={node} />
    }
    if (node.matchType === 'UploadReference') {
        return <UploadReferenceRetentionMatchNode match={node} />
    }

    throw new Error(`invalid node type ${JSON.stringify(node as object)}`)
}

const RetentionPolicyRetentionMatchNode: FunctionComponent<{ match: RetentionPolicyMatch }> = ({ match }) => (
    <>
        <span className={styles.separator} />

        <div className={classNames(styles.information, 'd-flex flex-column')}>
            <div className="m-0">
                {match.configurationPolicy ? (
                    <Link to={`../configuration/${match.configurationPolicy.id}`} className="p-0">
                        <h3 className="m-0 d-block d-md-inline">{match.configurationPolicy.name}</h3>
                    </Link>
                ) : (
                    <h3 className="m-0 d-block d-md-inline">{retentionByBranchTipTitle}</h3>
                )}
                <div className="mr-2 d-block d-mdinline-block">
                    Retained: {match.matches ? 'yes' : 'no'}
                    {match.protectingCommits.length !== 0 && (
                        <>
                            , by {match.protectingCommits.length} visible{' '}
                            {pluralize('commit', match.protectingCommits.length)}, including{' '}
                            {match.protectingCommits
                                .slice(0, 4)
                                .map(hash => hash.slice(0, 9))
                                .join(', ')}
                            <InformationOutlineIcon
                                className="ml-1 icon-inline"
                                data-tooltip="This upload is retained to service code-intel queries for commit(s) with applicable retention policies."
                            />
                        </>
                    )}
                    {!match.configurationPolicy && (
                        <InformationOutlineIcon
                            className="ml-1 icon-inline"
                            data-tooltip="Uploads at the tip of the default branch are always retained indefinitely."
                        />
                    )}
                </div>
            </div>
        </div>
    </>
)

const UploadReferenceRetentionMatchNode: FunctionComponent<{ match: UploadReferenceMatch }> = ({ match }) => (
    <>
        <span className={styles.separator} />

        <div className={classNames(styles.information, 'd-flex flex-column')}>
            <div className="m-0">
                <h3 className="m-0 d-block d-md-inline">{retentionByUploadTitle}</h3>
                <div className="mr-2 d-block d-mdinline-block">
                    Referenced by {match.total} {pluralize('upload', match.total, 'uploads')}, including{' '}
                    {match.uploadSlice
                        .slice(0, 3)
                        .map<React.ReactNode>(upload => (
                            <Link key={upload.id} to={`/site-admin/code-intelligence/uploads/${upload.id}`}>
                                {upload.projectRoot?.repository.name ?? 'unknown'}
                            </Link>
                        ))
                        .reduce((previous, current) => [previous, ', ', current])}
                    <InformationOutlineIcon
                        className="ml-1 icon-inline"
                        data-tooltip="Uploads that are dependencies of other upload(s) are retained to service cross-repository code-intel queries."
                    />
                </div>
            </div>
        </div>
    </>
)
