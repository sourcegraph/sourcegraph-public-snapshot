import classnames from 'classnames'
import RefreshIcon from 'mdi-react/RefreshIcon'
import React, { ReactElement, ReactNode } from 'react'
import { ChartContent } from 'sourcegraph'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { isErrorLike } from '@sourcegraph/shared/src/util/errors'

import { ErrorAlert } from '../../../../components/alerts'
import { ChartViewContent } from '../../../../views/components/content/view-content/chart-view-content/ChartViewContent'

import styles from './LivePreviewContainer.module.scss'

export interface LivePreviewContainerProps {
    onUpdateClick: () => void
    loading: boolean
    disabled: boolean
    className?: string
    dataOrError: ChartContent | Error | undefined
    defaultMock: ChartContent
    mockMessage: ReactNode
    description?: ReactNode
}

export function LivePreviewContainer(props: LivePreviewContainerProps): ReactElement {
    const { disabled, loading, dataOrError, defaultMock, onUpdateClick, className, mockMessage, description } = props

    return (
        <section className={classnames(styles.livePreview, className)}>
            <div className="d-flex align-items-center mb-1">
                Live preview
                <button type="button" disabled={disabled} className="btn btn-icon ml-1" onClick={onUpdateClick}>
                    <RefreshIcon size="1rem" />
                </button>
            </div>

            {loading && (
                <div
                    className={classnames(
                        styles.livePreviewLoader,
                        'flex-grow-1 d-flex flex-column align-items-center justify-content-center'
                    )}
                >
                    <LoadingSpinner /> Loading code insight
                </div>
            )}

            {isErrorLike(dataOrError) && <ErrorAlert className="m-0" error={dataOrError} />}

            {!loading && !isErrorLike(dataOrError) && (
                <div className={classnames(styles.livePreviewChartContainer, 'card')}>
                    <ChartViewContent
                        className={classnames(styles.livePreviewChart, 'card-body', {
                            [styles.livePreviewChartWithMock]: !dataOrError,
                        })}
                        viewID="search-insight-live-preview"
                        telemetryService={NOOP_TELEMETRY_SERVICE}
                        content={dataOrError ?? defaultMock}
                    />

                    {!dataOrError && <p className={styles.livePreviewLoadingChartInfo}>{mockMessage}</p>}
                </div>
            )}

            {description && <span className="mt-2 text-muted">{description}</span>}
        </section>
    )
}
