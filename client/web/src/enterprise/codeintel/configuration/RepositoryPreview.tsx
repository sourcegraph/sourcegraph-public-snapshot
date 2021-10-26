import React, { FunctionComponent } from 'react'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { ErrorAlert } from '@sourcegraph/web/src/components/alerts'

import { usePreviewRepositoryFilter } from './useSearchRepositories'

interface RepositoryPreviewProps {
    pattern: string
}

const RepositoryHeader = <h3>Preview of Repository filter</h3>

export const RepositoryPreview: FunctionComponent<RepositoryPreviewProps> = ({ pattern }) => {
    const { previewResult: preview, isLoadingPreview: previewLoading, previewError } = usePreviewRepositoryFilter(
        pattern
    )

    return (
        <>
            {RepositoryHeader}

            {pattern === '' ? (
                <>
                    <small>Enter a pattern to preview matching repositories.</small>{' '}
                </>
            ) : (
                <div>
                    {/* className={styles.wrapper}> */}
                    {RepositoryHeader}
                    <small>
                        {preview.preview.length === 0 ? (
                            <>Configuration policy does not match any known commits.</>
                        ) : (
                            <>
                                Configuration policy will be applied to the following
                                {/* {typeText} */}
                            </>
                        )}
                    </small>

                    {previewError && (
                        <ErrorAlert prefix="Error fetching matching repository objects" error={previewError} />
                    )}

                    {previewLoading ? (
                        <LoadingSpinner />
                    ) : (
                        // className={styles.loading} />
                        <>
                            {preview.preview.length !== 0 ? (
                                <div className="mt-2 pt-2">
                                    {/* <div className={classNames('bg-dark text-light p-2', styles.container)}> */}
                                    {preview.preview.map(tag => (
                                        <p key={tag.name}>{tag.name}</p>
                                    ))}
                                    {/* </div> */}
                                </div>
                            ) : (
                                <div className="mt-2 pt-2">
                                    <div>
                                        {/* className={styles.empty}> */}
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
