import { FunctionComponent, useCallback } from 'react'

import { useApolloClient } from '@apollo/client'
import { mdiInformationOutline, mdiMapSearch } from '@mdi/js'
import classNames from 'classnames'
import * as H from 'history'
import { Observable } from 'rxjs'

import { isErrorLike, pluralize } from '@sourcegraph/common'
import { H3, Icon, Link, Text, Tooltip } from '@sourcegraph/wildcard'

import {
    Connection,
    FilteredConnection,
    FilteredConnectionQueryArguments,
} from '../../../../components/FilteredConnection'
import { PreciseIndexFields } from '../../../../graphql-operations'
import {
    NormalizedUploadRetentionMatch,
    queryPreciseIndexRetention as defaultQueryPreciseIndexRetention,
    RetentionPolicyMatch,
    UploadReferenceMatch,
} from '../hooks/queryPreciseIndexRetention'

import styles from './Retention.module.scss'

export interface RetentionPanelProps {
    index: PreciseIndexFields
    history: H.History
    location: H.Location
    queryPreciseIndexRetention?: typeof defaultQueryPreciseIndexRetention
}

export const RetentionPanel: FunctionComponent<RetentionPanelProps> = ({
    index,
    history,
    location,
    queryPreciseIndexRetention = defaultQueryPreciseIndexRetention,
}) => {
    const apolloClient = useApolloClient()

    const queryRetentionPoliciesCallback = useCallback(
        (args: FilteredConnectionQueryArguments): Observable<Connection<NormalizedUploadRetentionMatch>> => {
            if (index && !isErrorLike(index)) {
                return queryPreciseIndexRetention(apolloClient, index.id, {
                    matchesOnly: false,
                    ...args,
                })
            }

            throw new Error('unreachable: queryRetentionPolicies referenced with invalid upload')
        },
        [index, apolloClient, queryPreciseIndexRetention]
    )

    return (
        <FilteredConnection
            listComponent="table"
            headComponent={() => (
                <thead>
                    <tr>
                        <th>Policy</th>
                        <th>Matches</th>
                    </tr>
                </thead>
            )}
            listClassName={classNames(styles.grid, 'mb-3')}
            noun="match"
            pluralNoun="matches"
            nodeComponent={RetentionMatchNode}
            queryConnection={queryRetentionPoliciesCallback}
            history={history}
            location={location}
            cursorPaging={true}
            useURLQuery={false}
            hideSearch={true}
            emptyElement={<EmptyUploadRetentionMatchStatus />}
        />
    )
}

interface RetentionMatchNodeProps {
    node: NormalizedUploadRetentionMatch
}

const retentionByUploadTitle = 'Retention by reference'
const retentionByBranchTipTitle = 'Retention by tip of default branch'

const RetentionMatchNode: FunctionComponent<React.PropsWithChildren<RetentionMatchNodeProps>> = ({ node }) => {
    if (node.matchType === 'RetentionPolicy') {
        return <RetentionPolicyRetentionMatchNode match={node} />
    }
    if (node.matchType === 'UploadReference') {
        return <UploadReferenceRetentionMatchNode match={node} />
    }

    throw new Error(`invalid node type ${JSON.stringify(node as object)}`)
}

const RetentionPolicyRetentionMatchNode: FunctionComponent<
    React.PropsWithChildren<{ match: RetentionPolicyMatch }>
> = ({ match }) => (
    <tr>
        <td>
            {match.configurationPolicy ? (
                <Link to={`../configuration/${match.configurationPolicy.id}`} className="p-0">
                    <H3 className="m-0 d-block d-md-inline">{match.configurationPolicy.name}</H3>
                </Link>
            ) : (
                <H3 className="m-0 d-block d-md-inline">{retentionByBranchTipTitle}</H3>
            )}
            {match.protectingCommits.length !== 0 && (
                <>
                    , by {match.protectingCommits.length} visible {pluralize('commit', match.protectingCommits.length)},
                    including{' '}
                    {match.protectingCommits
                        .slice(0, 4)
                        .map(hash => hash.slice(0, 7))
                        .join(', ')}
                    <Tooltip content="This upload is retained to service code-intel queries for commit(s) with applicable retention policies.">
                        <Icon
                            aria-label="This upload is retained to service code-intel queries for commit(s) with applicable retention policies."
                            className="ml-1"
                            svgPath={mdiInformationOutline}
                        />
                    </Tooltip>
                </>
            )}
            {!match.configurationPolicy && (
                <Tooltip content="Uploads at the tip of the default branch are always retained indefinitely.">
                    <Icon
                        aria-label="Uploads at the tip of the default branch are always retained indefinitely."
                        className="ml-1"
                        svgPath={mdiInformationOutline}
                    />
                </Tooltip>
            )}
        </td>

        <td>
            <div className="mr-2 d-block d-mdinline-block">Retained: {match.matches ? 'yes' : 'no'}</div>
        </td>
    </tr>
)

const UploadReferenceRetentionMatchNode: FunctionComponent<
    React.PropsWithChildren<{ match: UploadReferenceMatch }>
> = ({ match }) => (
    <tr>
        <td>{retentionByUploadTitle}</td>

        <td>
            Referenced by {match.total} {pluralize('upload', match.total, 'uploads')}, including{' '}
            {match.uploadSlice
                .slice(0, 3)
                .map<React.ReactNode>(upload => (
                    <Link key={upload.id} to={`/site-admin/code-graph/uploads/${upload.id}`}>
                        {upload.projectRoot?.repository.name ?? 'unknown'}
                    </Link>
                ))
                .reduce((previous, current) => [previous, ', ', current])}
            <Tooltip content="Uploads that are dependencies of other upload(s) are retained to service cross-repository code-intel queries.">
                <Icon
                    aria-label="Uploads that are dependencies of other upload(s) are retained to service cross-repository code-intel queries."
                    className="ml-1"
                    svgPath={mdiInformationOutline}
                />
            </Tooltip>
        </td>
    </tr>
)

const EmptyUploadRetentionMatchStatus: React.FunctionComponent<React.PropsWithChildren<unknown>> = () => (
    <Text alignment="center" className="text-muted w-100 mb-0 mt-1">
        <Icon className="mb-2" svgPath={mdiMapSearch} inline={false} aria-hidden={true} />
        <br />
        No retention policies matched.
    </Text>
)
