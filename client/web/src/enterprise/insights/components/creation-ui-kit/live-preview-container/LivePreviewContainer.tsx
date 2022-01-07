import classNames from 'classnames'
import RefreshIcon from 'mdi-react/RefreshIcon'
import React, { PropsWithChildren, ReactElement, ReactNode } from 'react'
import { ChartContent } from 'sourcegraph'

import { isErrorLike } from '@sourcegraph/common'
import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { LineChartLayoutOrientation, LineChartSettingsContext, ChartViewContentLayout } from '../../../../../views'
import * as View from '../../../../../views'

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
        mockMessage,
        description,
        chartContentClassName,
    } = props

    return (
        <aside className={classNames(styles.livePreview, className)}>
            <div className="d-flex align-items-center mb-1">
                Live preview
                <button type="button" disabled={disabled} className="btn btn-icon ml-1" onClick={onUpdateClick}>
                    <RefreshIcon size="1rem" />
                </button>
            </div>

            <View.Root className={classNames(chartContentClassName, 'flex-grow-1')}>
                {loading ? (
                    <View.LoadingContent text="Loading code insight" />
                ) : isErrorLike(dataOrError) ? (
                    <View.ErrorContent error={dataOrError} title="" />
                ) : (
                    <LineChartSettingsContext.Provider value={LINE_CHART_SETTINGS}>
                        <View.Content
                            telemetryService={NOOP_TELEMETRY_SERVICE}
                            content={[dataOrError ?? defaultMock]}
                            layout={ChartViewContentLayout.ByContentSize}
                            className={classNames({ [styles.chartWithMock]: !dataOrError })}
                        />

                        {!dataOrError && <p className={styles.loadingChartInfo}>{mockMessage}</p>}
                    </LineChartSettingsContext.Provider>
                )}
            </View.Root>

            {description && <span className="mt-2 text-muted">{description}</span>}
        </aside>
    )
}
