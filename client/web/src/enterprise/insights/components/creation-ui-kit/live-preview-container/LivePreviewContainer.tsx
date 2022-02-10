import classNames from 'classnames'
import RefreshIcon from 'mdi-react/RefreshIcon'
import React, { PropsWithChildren, ReactElement, ReactNode } from 'react'
import { ChartContent } from 'sourcegraph'

import { isErrorLike } from '@sourcegraph/common'
import { Button } from '@sourcegraph/wildcard'

import { LineChartLayoutOrientation, LineChartSettingsContext, ChartViewContentLayout } from '../../../../../views'
import * as View from '../../../../../views'

import styles from './LivePreviewContainer.module.scss'

const LINE_CHART_SETTINGS = {
    zeroYAxisMin: false,
    layout: LineChartLayoutOrientation.Vertical,
}

export interface LivePreviewContainerProps {
    loading: boolean
    disabled: boolean
    livePreviewControls?: boolean
    chartContentClassName?: string
    dataOrError: ChartContent | Error | undefined
    defaultMock: ChartContent
    mockMessage: ReactNode
    title?: string
    description?: ReactNode
    className?: string
    onUpdateClick: () => void
    children?: (data: ChartContent) => ReactNode
}

export function LivePreviewContainer(props: PropsWithChildren<LivePreviewContainerProps>): ReactElement {
    const {
        title = '',
        livePreviewControls = true,
        disabled,
        loading,
        dataOrError,
        defaultMock,
        onUpdateClick,
        className,
        mockMessage,
        description,
        chartContentClassName,
        children,
    } = props

    return (
        <aside className={classNames(styles.livePreview, className)}>
            {livePreviewControls && (
                <div className="d-flex align-items-center mb-1">
                    Live preview
                    <Button disabled={disabled} variant="icon" className="ml-1" onClick={onUpdateClick}>
                        <RefreshIcon size="1rem" />
                    </Button>
                </div>
            )}

            <View.Root title={title} className={classNames(chartContentClassName, 'flex-grow-1')}>
                {loading ? (
                    <View.LoadingContent text="Loading code insight" />
                ) : isErrorLike(dataOrError) ? (
                    <View.ErrorContent error={dataOrError} title="" />
                ) : (
                    <LineChartSettingsContext.Provider value={LINE_CHART_SETTINGS}>
                        {dataOrError ? (
                            children ? (
                                children(dataOrError)
                            ) : (
                                <View.Content content={[dataOrError]} layout={ChartViewContentLayout.ByContentSize} />
                            )
                        ) : (
                            <View.Content
                                content={[defaultMock]}
                                layout={ChartViewContentLayout.ByContentSize}
                                className={styles.chartWithMock}
                            />
                        )}

                        {!dataOrError && <p className={styles.loadingChartInfo}>{mockMessage}</p>}
                    </LineChartSettingsContext.Provider>
                )}
            </View.Root>

            {description && <span className="mt-2 text-muted">{description}</span>}
        </aside>
    )
}
