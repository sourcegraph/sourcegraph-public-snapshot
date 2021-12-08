import classNames from 'classnames'
import RefreshIcon from 'mdi-react/RefreshIcon'
import React, { PropsWithChildren, ReactElement, ReactNode } from 'react'
import { ChartContent } from 'sourcegraph'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { isErrorLike } from '@sourcegraph/shared/src/util/errors'

import { ErrorAlert } from '../../../../components/alerts'
import {
    LineChartLayoutOrientation,
    LineChartSettingsContext,
    ChartViewContent,
    ChartViewContentLayout,
} from '../../../../views'

import styles from './LivePreviewContainer.module.scss'

const LINE_CHART_SETTINGS = {
    zeroYAxisMin: false,
    layout: LineChartLayoutOrientation.Vertical,
}

export interface LivePreviewContainerProps {
    onUpdateClick: () => void
    loading: boolean
    disabled: boolean
    className?: string
    chartContentClassName?: string
    dataOrError: ChartContent | Error | undefined
    defaultMock: ChartContent
    mockMessage: ReactNode
    description?: ReactNode
}

export function LivePreviewContainer(props: PropsWithChildren<LivePreviewContainerProps>): ReactElement {
    const {
        disabled,
        loading,
        dataOrError,
        defaultMock,
        onUpdateClick,
        className,
        chartContentClassName,
        mockMessage,
        description,
    } = props

    return (
        <section className={classNames(styles.livePreview, className)}>
            <div className="d-flex align-items-center mb-1">
                Live preview
                <button type="button" disabled={disabled} className="btn btn-icon ml-1" onClick={onUpdateClick}>
                    <RefreshIcon size="1rem" />
                </button>
            </div>

            {loading && (
                <div
                    className={classNames(
                        styles.livePreviewLoader,
                        'flex-grow-1 d-flex flex-column align-items-center justify-content-center'
                    )}
                >
                    <LoadingSpinner /> Loading code insight
                </div>
            )}

            {isErrorLike(dataOrError) && <ErrorAlert className="m-0" error={dataOrError} />}

            {!loading && !isErrorLike(dataOrError) && (
                <div className={classNames(styles.livePreviewChartContainer, chartContentClassName, 'card')}>
                    <LineChartSettingsContext.Provider value={LINE_CHART_SETTINGS}>
                        <ChartViewContent
                            className={classNames(styles.livePreviewChart, 'card-body', {
                                [styles.livePreviewChartWithMock]: !dataOrError,
                            })}
                            viewID="search-insight-live-preview"
                            telemetryService={NOOP_TELEMETRY_SERVICE}
                            content={dataOrError ?? defaultMock}
                            layout={ChartViewContentLayout.ByContentSize}
                        />
                    </LineChartSettingsContext.Provider>

                    {!dataOrError && <p className={styles.livePreviewLoadingChartInfo}>{mockMessage}</p>}
                </div>
            )}

            {description && <span className="mt-2 text-muted">{description}</span>}
        </section>
    )
}
