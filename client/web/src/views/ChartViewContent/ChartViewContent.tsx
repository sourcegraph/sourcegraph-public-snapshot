import React, { FunctionComponent, useMemo} from 'react';
import { ChartContent } from 'sourcegraph';
import * as H from 'history';
import { ParentSize } from '@visx/responsive';
import { useDebouncedCallback } from 'use-debounce'

import {
    createProgrammaticallyLinkHandler
} from '../../../../shared/src/components/linkClickHandler'
import { eventLogger } from '../../tracking/eventLogger';

import { LineChart } from './charts/line/LineChart';
import { PieChart } from './charts/pie/PieChart';
import { BarChart } from './charts/bar/BarChart';
import { DatumClickEvent } from './charts/types';

/**
 * Displays chart view content.
 */
export interface ChartViewContentProps {
    content: ChartContent
    history: H.History
    viewID: string
}

export const ChartViewContent: FunctionComponent<ChartViewContentProps> = props => {
    const { content, ...otherProps } = props;

    // Because xychart fires all consumer's handlers twice, we need to debounce our handler
    // Remove debounce when https://github.com/airbnb/visx/issues/1077 will be resolved
    const linkHandler = useDebouncedCallback(
        useMemo(() => {
            const linkHandler = createProgrammaticallyLinkHandler(otherProps.history)
            return (event: DatumClickEvent): void => {
                if (!event.link) {
                    return
                }

                eventLogger.log('InsightDataPointClick', { insightType: otherProps.viewID.split('.')[0] })
                linkHandler(event.originEvent, event.link)
            }
        }, [otherProps.history, otherProps.viewID])
    )

    return (
        <div className="chart-view-content" >
            <ParentSize className='chart-view-content__chart'>
                {
                    ({ width, height}) => {
                        if (content.chart === 'line' || content.chart === 'bar') {
                            const ChartComponent = content.chart === 'line' ? LineChart : BarChart

                            return (
                                <ChartComponent
                                    {...content}
                                    width={width}
                                    height={height}
                                    onDatumClick={linkHandler}
                                />
                            );
                        }

                        return (
                            <PieChart
                                {...content}
                                width={width}
                                height={height}
                                onDatumClick={linkHandler}
                            />
                        );
                    }
                }
            </ParentSize>
        </div>
    );
}
