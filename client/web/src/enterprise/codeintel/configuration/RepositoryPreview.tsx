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

    if (!pattern) {
        return <small>Enter a pattern to preview matching repositories.</small>
    }

    if (previewError) {
        return <ErrorAlert prefix="Error fetching matching repository objects" error={previewError} />
    }

    if (previewLoading) {
        return <LoadingSpinner className={styles.loading} />
    }

    return (
        <div className={styles.wrapper}>
            {preview.preview.length === 0 ? (
                <>
                    <small>Pattern does not match any repositories.</small>
                    <div>
                        <div className={styles.empty}>
                            <p className="text-monospace">N/A</p>
                        </div>
                    </div>
                </>
            ) : (
                <>
                    <small>Configuration policy will be applied to the following repositories.</small>
                    <div>
                        <div className={classNames('bg-dark text-light p-2', styles.container)}>
                            {preview.preview.map(tag => (
                                <p key={tag.name}>{tag.name}</p>
                            ))}
                        </div>
                    </div>
                </>
            )}
        </div>
    )
}
