import { FunctionComponent } from 'react'

import classNames from 'classnames'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { LoadingSpinner, Alert } from '@sourcegraph/wildcard'

import { usePreviewRepositoryFilter } from '../hooks/usePreviewRepositoryFilter'

import styles from './RepositoryPreview.module.scss'

interface RepositoryPreviewProps {
    patterns: string[]
}

export const RepositoryPreview: FunctionComponent<React.PropsWithChildren<RepositoryPreviewProps>> = ({ patterns }) => {
    const { previewResult: preview, isLoadingPreview: previewLoading, previewError } = usePreviewRepositoryFilter(
        patterns
    )

    return (
        <div>
            {previewError && <ErrorAlert prefix="Error fetching matching repositories" error={previewError} />}

            {previewLoading ? (
                <LoadingSpinner inline={false} className={styles.loading} />
            ) : (
                preview && (
                    <>
                        <p>
                            <small>
                                {preview.repositoryNames.length === 0
                                    ? 'Configuration policy does not match any known repositories.'
                                    : preview.repositoryNames.length === 1
                                    ? 'Configuration policy will be applied to the following repository.'
                                    : preview.repositoryNames.length < preview.totalCount
                                    ? `Configuration policy will be applied to ${preview.totalCount} repositories (showing ${preview.repositoryNames.length} below).`
                                    : `Configuration policy will be applied to the following ${preview.repositoryNames.length} repositories.`}
                            </small>
                        </p>

                        <div>
                            {preview.repositoryNames.length === 0 ? (
                                <div className="mt-2 pt-2">
                                    <div className={styles.empty}>
                                        <p className="text-monospace">N/A</p>
                                    </div>
                                </div>
                            ) : (
                                <>
                                    <div className="mt-2 pt-2">
                                        <div className={classNames('bg-dark text-light p-2', styles.container)}>
                                            {preview.repositoryNames.map(name => (
                                                <p key={`${name}`} className="text-monospace p-0 m-0">
                                                    <span className="search-filter-keyword">repo:</span>
                                                    <span>{name}</span>
                                                </p>
                                            ))}
                                        </div>
                                    </div>
                                </>
                            )}
                        </div>

                        {preview.totalMatches > preview.totalCount && (
                            <Alert variant="danger">
                                Each policy pattern can match a maximum of {preview.limit} repositories. There are{' '}
                                {preview.totalMatches - preview.totalCount} additional repositories that match the
                                filter not covered by this policy. Create a more constrained policy or increase the
                                system limit.
                            </Alert>
                        )}
                    </>
                )
            )}
        </div>
    )
}
