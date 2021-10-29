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
        <div>
            <small>
                {preview.preview.length === 0 ? (
                    <>Configuration policy does not match any known repositories.</>
                ) : (
                    <>Configuration policy will be applied to the following repositories. </>
                )}
            </small>

            {previewError && <ErrorAlert prefix="Error fetching matching repositories" error={previewError} />}

            {previewLoading ? (
                <LoadingSpinner className={styles.loading} />
            ) : (
                <div>
                    {preview.preview.length === 0 ? (
                        <div className="mt-2 pt-2">
                            <div className={styles.empty}>
                                <p className="text-monospace">N/A</p>
                            </div>
                        </div>
                    ) : (
                        <>
                            <div className="mt-2 pt-2">
                                <div className={classNames('bg-dark text-light p-2', styles.container)}>
                                    {preview.preview.map(tag => (
                                        <p key={`${tag.name}`} className="text-monospace p-0 m-0">
                                            <span className="search-filter-keyword">repo:</span>
                                            <span>{tag.name}</span>
                                        </p>
                                    ))}
                                </div>
                            </div>
                        </>
                    )}
                </div>
            )}
        </div>
    )
}
