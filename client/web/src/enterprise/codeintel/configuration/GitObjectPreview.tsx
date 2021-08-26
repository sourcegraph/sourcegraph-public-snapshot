import React, { FunctionComponent, useEffect, useState } from 'react'

import { ErrorAlert } from '@sourcegraph/web/src/components/alerts'
import { LoadingSpinner } from '@sourcegraph/wildcard'

import { GitObjectType } from '../../../graphql-operations'

import {
    repoName as defaultRepoName,
    searchGitBranches as defaultSearchGitBranches,
    searchGitTags as defaultSearchGitTags,
} from './backend'

export interface GitObjectPreviewProps {
    repoId: string
    type: GitObjectType
    pattern: string
    repoName: typeof defaultRepoName
    searchGitTags: typeof defaultSearchGitTags
    searchGitBranches: typeof defaultSearchGitBranches
}

enum PreviewState {
    Idle,
    LoadingTags,
}

export const GitObjectPreview: FunctionComponent<GitObjectPreviewProps> = ({
    repoId,
    type,
    pattern,
    repoName,
    searchGitTags,
    searchGitBranches,
}) => {
    const [state, setState] = useState(() => PreviewState.Idle)
    const [commitPreview, setCommitPreview] = useState<GitObjectPreviewResult>()
    const [commitPreviewFetchError, setCommitPreviewFetchError] = useState<Error>()

    useEffect(() => {
        async function inner(): Promise<void> {
            setState(PreviewState.LoadingTags)
            setCommitPreviewFetchError(undefined)

            const resultFactories = [
                { type: GitObjectType.GIT_COMMIT, factory: () => resultFromCommit(repoId, pattern, repoName) },
                { type: GitObjectType.GIT_TAG, factory: () => resultFromTag(repoId, pattern, searchGitTags) },
                { type: GitObjectType.GIT_TREE, factory: () => resultFromBranch(repoId, pattern, searchGitBranches) },
            ]

            try {
                for (const { type: match, factory } of resultFactories) {
                    if (type === match) {
                        setCommitPreview(await factory())
                        break
                    }
                }
            } catch (error) {
                setCommitPreviewFetchError(error)
            } finally {
                setState(PreviewState.Idle)
            }
        }

        inner().catch(console.error)
    }, [repoId, type, pattern, repoName, searchGitTags, searchGitBranches])

    return (
        <>
            <h3>Preview of Git object filter</h3>

            {type ? (
                <>
                    <small>
                        {commitPreview?.preview.length === 0 ? (
                            <>Configuration policy does not match any known commits.</>
                        ) : (
                            <>
                                Configuration policy will be applied to the following
                                {type === GitObjectType.GIT_COMMIT
                                    ? ' commit'
                                    : type === GitObjectType.GIT_TAG
                                    ? ' tags'
                                    : type === GitObjectType.GIT_TREE
                                    ? ' branches'
                                    : ''}
                                .
                            </>
                        )}
                    </small>

                    {commitPreviewFetchError ? (
                        <ErrorAlert
                            prefix="Error fetching matching repository objects"
                            error={commitPreviewFetchError}
                        />
                    ) : (
                        <>
                            {commitPreview !== undefined && commitPreview.preview.length !== 0 && (
                                <div className="mt-2 p-2">
                                    <div className="bg-dark text-light p-2">
                                        {commitPreview.preview.map(tag => (
                                            <p key={tag.revlike} className="text-monospace p-0 m-0">
                                                <span className="search-filter-keyword">repo:</span>
                                                <span>{tag.name}</span>
                                                <span className="search-filter-keyword">@</span>
                                                <span>{tag.revlike}</span>
                                            </p>
                                        ))}
                                    </div>

                                    {commitPreview.preview.length < commitPreview.totalCount && (
                                        <p className="pt-2">
                                            ...and {commitPreview.totalCount - commitPreview.preview.length} other
                                            matches
                                        </p>
                                    )}
                                </div>
                            )}
                            {state === PreviewState.LoadingTags && <LoadingSpinner />}
                        </>
                    )}
                </>
            ) : (
                <small>Select a Git object type to preview matching commits.</small>
            )}
        </>
    )
}

interface GitObjectPreviewResult {
    preview: { name: string; revlike: string }[]
    totalCount: number
}

const resultFromCommit = async (
    repoId: string,
    pattern: string,
    repoName: typeof defaultRepoName
): Promise<GitObjectPreviewResult> => {
    const result = await repoName(repoId).toPromise()
    if (!result) {
        return { preview: [], totalCount: 0 }
    }

    return { preview: [{ name: result.name, revlike: pattern }], totalCount: 1 }
}

const resultFromTag = async (
    repoId: string,
    pattern: string,
    searchGitTags: typeof defaultSearchGitTags
): Promise<GitObjectPreviewResult> => {
    const result = await searchGitTags(repoId, pattern).toPromise()
    if (!result) {
        return { preview: [], totalCount: 0 }
    }

    const { nodes, totalCount } = result.tags

    return {
        preview: nodes.map(node => ({ name: result.name, revlike: node.displayName })),
        totalCount,
    }
}

const resultFromBranch = async (
    repoId: string,
    pattern: string,
    searchGitBranches: typeof defaultSearchGitBranches
): Promise<GitObjectPreviewResult> => {
    const result = await searchGitBranches(repoId, pattern).toPromise()
    if (!result) {
        return { preview: [], totalCount: 0 }
    }

    const { nodes, totalCount } = result.branches

    return {
        preview: nodes.map(node => ({ name: result.name, revlike: node.displayName })),
        totalCount,
    }
}
