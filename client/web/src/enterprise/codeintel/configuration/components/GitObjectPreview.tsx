import { FunctionComponent } from 'react'

import { ApolloError } from '@apollo/client'
import classNames from 'classnames'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { Badge, LoadingSpinner, Typography } from '@sourcegraph/wildcard'

import { GitObjectType } from '../../../../graphql-operations'
import { GitObjectPreviewResult, usePreviewGitObjectFilter } from '../hooks/usePreviewGitObjectFilter'

import styles from './GitObjectPreview.module.scss'

export interface GitObjectPreviewWrapperProps {
    repoId: string
    type: GitObjectType
    pattern: string
}

const GitObjectHeader = <Typography.H3>Preview of Git object filter</Typography.H3>

export const GitObjectPreviewWrapper: FunctionComponent<React.PropsWithChildren<GitObjectPreviewWrapperProps>> = ({
    repoId,
    type,
    pattern,
}) => (
    <div className="form-group">
        {pattern === '' ? (
            <>
                {GitObjectHeader}
                <small>Enter a pattern to preview matching commits.</small>{' '}
            </>
        ) : type === GitObjectType.GIT_COMMIT ? (
            <GitCommitPreview repoId={repoId} pattern={pattern} typeText=" commit." />
        ) : type === GitObjectType.GIT_TAG ? (
            <GitTagPreview repoId={repoId} pattern={pattern} typeText=" tags." />
        ) : type === GitObjectType.GIT_TREE ? (
            <GitBranchesPreview repoId={repoId} pattern={pattern} typeText=" branches." />
        ) : (
            <>
                {GitObjectHeader}
                <small>Select a Git object type to preview matching commits.</small>
            </>
        )}
    </div>
)

export interface GitPreviewProps {
    repoId: string
    pattern: string
    typeText: string
}

const createGitCommitPreview = (type: GitObjectType): FunctionComponent<React.PropsWithChildren<GitPreviewProps>> => ({
    repoId,
    pattern,
    typeText,
}) => {
    const { previewResult, isLoadingPreview, previewError } = usePreviewGitObjectFilter(repoId, type, pattern)

    return (
        <GitObjectPreview
            typeText={typeText}
            preview={previewResult}
            previewLoading={isLoadingPreview}
            previewError={previewError}
        />
    )
}

const GitTagPreview: FunctionComponent<React.PropsWithChildren<GitPreviewProps>> = createGitCommitPreview(
    GitObjectType.GIT_TAG
)
const GitBranchesPreview: FunctionComponent<React.PropsWithChildren<GitPreviewProps>> = createGitCommitPreview(
    GitObjectType.GIT_TREE
)
const GitCommitPreview: FunctionComponent<React.PropsWithChildren<GitPreviewProps>> = createGitCommitPreview(
    GitObjectType.GIT_COMMIT
)

interface GitObjectPreviewProps {
    typeText: string
    preview: GitObjectPreviewResult
    previewError: ApolloError | undefined
    previewLoading: boolean
}

const GitObjectPreview: FunctionComponent<React.PropsWithChildren<GitObjectPreviewProps>> = ({
    typeText,
    preview,
    previewError,
    previewLoading,
}) => (
    <div>
        {GitObjectHeader}
        <small>
            {preview.preview.length === 0 ? (
                <>Configuration policy does not match any known commits.</>
            ) : (
                <>
                    Configuration policy will be applied to the following
                    {typeText}.
                </>
            )}
        </small>

        {previewError && <ErrorAlert prefix="Error fetching matching git objects" error={previewError} />}

        {previewLoading ? (
            <LoadingSpinner className={styles.loading} />
        ) : (
            <>
                {preview.preview.length === 0 ? (
                    <div className="mt-2 pt-2">
                        <div className={styles.empty}>
                            <p className="text-monospace">N/A</p>
                        </div>
                    </div>
                ) : (
                    <div className="mt-2 pt-2">
                        <div className={classNames('bg-dark text-light p-2', styles.container)}>
                            {preview.preview.map(tag => (
                                <p key={`${tag.repoName}@${tag.name}`} className="text-monospace p-0 m-0">
                                    <span className="search-filter-keyword">repo:</span>
                                    <span>{tag.repoName}</span>
                                    <span className="search-filter-keyword">@</span>
                                    <span>{tag.name}</span>
                                    <Badge variant="info" className="ml-4">
                                        {tag.rev.slice(0, 7)}
                                    </Badge>
                                </p>
                            ))}
                        </div>
                    </div>
                )}
            </>
        )}
    </div>
)
