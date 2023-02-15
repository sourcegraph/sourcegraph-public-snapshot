import React, { useMemo, useCallback, useEffect } from 'react'

import classNames from 'classnames'
import { useParams } from 'react-router-dom-v5-compat'
import { Observable } from 'rxjs'

import { gql, useQuery } from '@sourcegraph/http-client'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary/useTemporarySetting'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { LoadingSpinner, ErrorAlert } from '@sourcegraph/wildcard'

import { FileDiffNode, FileDiffNodeProps } from '../../components/diff/FileDiffNode'
import { FilteredConnection, FilteredConnectionQueryArguments } from '../../components/FilteredConnection'
import { PageTitle } from '../../components/PageTitle'
import {
    ExternalLinkFields,
    GitCommitFields,
    RepositoryCommitResult,
    RepositoryCommitVariables,
    RepositoryFields,
    FileDiffFields,
} from '../../graphql-operations'
import { GitCommitNode } from '../commits/GitCommitNode'
import { gitCommitFragment } from '../commits/RepositoryCommitsPage'
import { queryRepositoryComparisonFileDiffs, RepositoryComparisonDiff } from '../compare/RepositoryCompareDiffPage'

import styles from './RepositoryCommitPage.module.scss'

const COMMIT_QUERY = gql`
    query RepositoryCommit($repo: ID!, $revspec: String!) {
        node(id: $repo) {
            __typename
            ... on Repository {
                commit(rev: $revspec) {
                    __typename # Necessary for error handling to check if commit exists
                    ...GitCommitFields
                }
            }
        }
    }
    ${gitCommitFragment}
`

interface RepositoryCommitPageProps extends TelemetryProps, PlatformContextProps, ThemeProps, SettingsCascadeProps {
    repo: RepositoryFields
    onDidUpdateExternalLinks: (externalLinks: ExternalLinkFields[] | undefined) => void
}

export type { DiffMode } from '@sourcegraph/shared/src/settings/temporary/diffMode'

/** Displays a commit. */
export const RepositoryCommitPage: React.FunctionComponent<RepositoryCommitPageProps> = props => {
    const params = useParams<{ revspec: string }>()

    if (!params.revspec) {
        throw new Error('Missing `revspec` param!')
    }

    const { data, error, loading } = useQuery<RepositoryCommitResult, RepositoryCommitVariables>(COMMIT_QUERY, {
        variables: {
            repo: props.repo.id,
            revspec: params.revspec,
        },
    })

    const commit = useMemo(
        () => (data?.node && data?.node?.__typename === 'Repository' ? data?.node?.commit : undefined),
        [data]
    )

    const [diffMode, setDiffMode] = useTemporarySetting('repo.commitPage.diffMode', 'unified')

    useEffect(() => {
        props.telemetryService.logViewEvent('RepositoryCommit')
    }, [props.telemetryService])

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
                repo: props.repo.id,
                base: commitParentOrEmpty(commit!),
                head: commit!.oid,
            }),
        [commit, props.repo.id]
    )

    return (
        <div data-testid="repository-commit-page" className={classNames('p-3', styles.repositoryCommitPage)}>
            <PageTitle title={commit ? commit.subject : `Commit ${params.revspec}`} />
            {loading ? (
                <LoadingSpinner className="mt-2" />
            ) : error || !commit ? (
                <ErrorAlert className="mt-2" error={error ?? new Error('Commit not found')} />
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
                        updateOnChange={`${props.repo.id}:${commit.oid}`}
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
