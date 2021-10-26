import classNames from 'classnames'
import React, { FunctionComponent } from 'react'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { ErrorAlert } from '@sourcegraph/web/src/components/alerts'

import styles from './RepositoryPreview.module.scss'
import { usePreviewRepositoryFilter } from './useSearchRepositories'

interface RepositoryPreviewProps {
    pattern: string
}

export const RepositoryPreview: FunctionComponent<RepositoryPreviewProps> = ({ pattern }) => {
    const { previewResult: preview, isLoadingPreview: previewLoading, previewError } = usePreviewRepositoryFilter(
        pattern
    )

    return (
        <>
            {pattern === '' ? (
                <small>Enter a pattern to preview matching repositories.</small>
            ) : (
                <div className={styles.wrapper}>
                    <small>
                        {preview.preview.length === 0 ? (
                            <>Pattern does not match any repositories.</>
                        ) : (
                            <>Configuration policy will be applied to the following repositories.</>
                        )}
                    </small>

                    {previewError && (
                        <ErrorAlert prefix="Error fetching matching repository objects" error={previewError} />
                    )}

                    {previewLoading ? (
                        <LoadingSpinner className={styles.loading} />
                    ) : (
                        <>
                            {preview.preview.length !== 0 ? (
                                <div>
                                    <div className={classNames('bg-dark text-light p-2', styles.container)}>
                                        {preview.preview.map(tag => (
                                            <p key={tag.name}>{tag.name}</p>
                                        ))}
                                    </div>
                                </div>
                            ) : (
                                <div>
                                    <div className={styles.empty}>
                                        <p className="text-monospace">N/A</p>
                                    </div>
                                </div>
                            )}
                        </>
                    )}
                </div>
            )}
        </>
    )
}
