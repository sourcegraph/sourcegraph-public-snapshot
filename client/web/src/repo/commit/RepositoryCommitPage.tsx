import React, { useMemo, useCallback, useEffect } from 'react'

import type { ApolloError } from '@apollo/client'
import classNames from 'classnames'
import { useParams } from 'react-router-dom'
import type { Observable } from 'rxjs'

import { useQuery } from '@sourcegraph/http-client'
import type { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import type { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary/useTemporarySetting'
import { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import type { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { LoadingSpinner, ErrorAlert } from '@sourcegraph/wildcard'

import { FileDiffNode, type FileDiffNodeProps } from '../../components/diff/FileDiffNode'
import { FilteredConnection, type FilteredConnectionQueryArguments } from '../../components/FilteredConnection'
import { PageTitle } from '../../components/PageTitle'
import {
    type ExternalLinkFields,
    type GitCommitFields,
    type RepositoryCommitResult,
    type RepositoryCommitVariables,
    type RepositoryFields,
    type FileDiffFields,
    type RepositoryChangelistResult,
    type RepositoryChangelistVariables,
    RepositoryType,
} from '../../graphql-operations'
import { GitCommitNode } from '../commits/GitCommitNode'
import { queryRepositoryComparisonFileDiffs, type RepositoryComparisonDiff } from '../compare/RepositoryCompareDiffPage'

import { CHANGELIST_QUERY, COMMIT_QUERY } from './backend'

import styles from './RepositoryCommitPage.module.scss'

interface RepositoryCommitPageProps
    extends TelemetryProps,
        TelemetryV2Props,
        PlatformContextProps,
        SettingsCascadeProps {
    repo: RepositoryFields
    onDidUpdateExternalLinks: (externalLinks: ExternalLinkFields[] | undefined) => void
}

export type { DiffMode } from '@sourcegraph/shared/src/settings/temporary/diffMode'

/** Displays a commit. */
export const RepositoryCommitPage: React.FunctionComponent<RepositoryCommitPageProps> = props => {
    const params = useParams<{ revspec: string }>()

    if (!params.revspec) {
        throw new Error('Missing `revspec` param.')
    }

    const { data, error, loading } = useQuery<RepositoryCommitResult, RepositoryCommitVariables>(COMMIT_QUERY, {
        variables: {
            repo: props.repo.id,
            revspec: params.revspec,
        },
    })

    const commit = useMemo(
        () => (data?.node && data?.node?.__typename === 'Repository' ? data?.node?.commit || undefined : undefined),
        [data]
    )

    return (
        <RepositoryRevisionNodes
            error={error}
            loading={loading}
            revspec={params.revspec}
            changelistID=""
            commit={commit}
            {...props}
        />
    )
}

/** Displays a changelist. */
export const RepositoryChangelistPage: React.FunctionComponent<RepositoryCommitPageProps> = props => {
    const params = useParams<{ changelistID: string }>()

    if (!params.changelistID) {
        throw new Error('Missing `changelistID` param. It must be set.')
    }

    const { data, error, loading } = useQuery<RepositoryChangelistResult, RepositoryChangelistVariables>(
        CHANGELIST_QUERY,
        {
            variables: {
                repo: props.repo.id,
                changelistID: params.changelistID,
            },
        }
    )

    const commit = useMemo(
        () => (data?.node?.__typename === 'Repository' ? data?.node?.changelist?.commit : undefined),
        [data]
    )

    return (
        <RepositoryRevisionNodes
            error={error}
            loading={loading}
            changelistID={params.changelistID}
            revspec=""
            commit={commit}
            {...props}
        />
    )
}

interface RepositoryRevisionNodesProps
    extends TelemetryProps,
        TelemetryV2Props,
        PlatformContextProps,
        SettingsCascadeProps {
    error?: ApolloError
    loading: boolean

    revspec: string
    changelistID: string

    repo: RepositoryFields
    commit: GitCommitFields | undefined
    onDidUpdateExternalLinks: (externalLinks?: ExternalLinkFields[]) => void
}

const RepositoryRevisionNodes: React.FunctionComponent<RepositoryRevisionNodesProps> = props => {
    const [diffMode, setDiffMode] = useTemporarySetting('repo.commitPage.diffMode', 'unified')

    const { error, loading, commit, repo } = props

    useEffect(() => {
        props.telemetryService.logViewEvent('RepositoryCommit')
        props.telemetryRecorder.recordEvent('RepositoryCommit', 'viewed')
    }, [props.telemetryService, props.telemetryRecorder])

    useEffect(() => {
        if (commit) {
            props.onDidUpdateExternalLinks(commit.externalURLs)
        }

        return () => {
            props.onDidUpdateExternalLinks(undefined)
        }
    }, [commit, props])

    const queryDiffs = useCallback(
        (args: FilteredConnectionQueryArguments): Observable<RepositoryComparisonDiff['comparison']['fileDiffs']> =>
            // Non-null assertions here are safe because the query is only executed if the commit is defined.
            queryRepositoryComparisonFileDiffs({
                first: args.first ?? null,
                after: args.after ?? null,
                paths: [],
                repo: repo.id,
                base: commitParentOrEmpty(commit!),
                head: commit!.oid,
            }),
        [commit, repo.id]
    )

    const pageTitle = commit
        ? commit.subject
        : repo.sourceType === RepositoryType.PERFORCE_DEPOT
        ? `Changelist ${props.changelistID}`
        : `Commit ${props.revspec}`

    const pageError = repo.sourceType === RepositoryType.PERFORCE_DEPOT ? 'Changelist not found' : 'Commit not found'

    return (
        <div data-testid="repository-commit-page" className={classNames('p-3', styles.repositoryCommitPage)}>
            <PageTitle title={pageTitle} />
            {loading ? (
                <LoadingSpinner className="mt-2" />
            ) : error || !commit ? (
                <ErrorAlert className="mt-2" error={error ?? new Error(pageError)} />
            ) : (
                <>
                    <div className="border-bottom pb-2">
                        <div>
                            <GitCommitNode
                                node={commit}
                                expandCommitMessageBody={true}
                                showSHAAndParentsRow={true}
                                diffMode={diffMode}
                                onHandleDiffMode={setDiffMode}
                                className={styles.gitCommitNode}
                            />
                        </div>
                    </div>
                    <FilteredConnection<FileDiffFields, Omit<FileDiffNodeProps, 'node'>>
                        listClassName="list-group list-group-flush"
                        noun="changed file"
                        pluralNoun="changed files"
                        queryConnection={queryDiffs}
                        nodeComponent={FileDiffNode}
                        nodeComponentProps={{
                            ...props,
                            lineNumbers: true,
                            diffMode,
                        }}
                        updateOnChange={`${repo.id}:${commit.oid}`}
                        defaultFirst={15}
                        hideSearch={true}
                        noSummaryIfAllNodesVisible={true}
                        withCenteredSummary={true}
                        cursorPaging={true}
                    />
                </>
            )}
        </div>
    )
}

function commitParentOrEmpty(commit: GitCommitFields): string {
    // 4b825dc642cb6eb9a060e54bf8d69288fbee4904 is `git hash-object -t tree /dev/null`, which is used as the base
    // when computing the `git diff` of the root commit.
    return commit.parents.length > 0 ? commit.parents[0].oid : '4b825dc642cb6eb9a060e54bf8d69288fbee4904'
}
