import React, { FunctionComponent, useMemo} from 'react';
import { ChartContent } from 'sourcegraph';
import * as H from 'history';
import { ParentSize } from '@visx/responsive';

import { createLinkClickHandler } from '../../../../shared/src/components/linkClickHandler'
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

    const linkHandler = useMemo(() => {
        const linkHandler = createLinkClickHandler(otherProps.history)
        return (event: DatumClickEvent): void => {
            console.log('Link handler chart view content')
            if (!event.link) {
                return
            }

            eventLogger.log('InsightDataPointClick', { insightType: otherProps.viewID.split('.')[0] })
            linkHandler(event.originEvent, event.link)
        }
    }, [otherProps.history, otherProps.viewID])

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

// interface CartesianChartViewContentProps extends ChartViewContentProps {
//     content: LineChartContent<any, string> | BarChartContent<any, string>;
// }
//
// export const CartesianChartViewContent: FunctionComponent<CartesianChartViewContentProps> = props => {
//     const { chart, ...content } = props.content;
//
//     const ChartComponent = chart === 'line' ? LineChart : BarChart
//
//     return (
//         <div className="chart-view-content" >
//             <ParentSize className='chart-view-content__chart'>
//                 {
//                     ({ width, height}) =>
//                         <ChartComponent
//                             {...content}
//                             width={width}
//                             height={height}
//                             onDatumClick={}
//                         />
//                 }
//             </ParentSize>
//         </div>
//     );
// };
//
// interface PieChartViewContentProps extends ChartViewContentProps {
//     content: PieChartContent<any>
// }
//
// export const PieChartViewContent: FunctionComponent<PieChartViewContentProps> = props => {
//     const { content } = props;
//
//     return (
//         <div className="chart-view-content" >
//             <ParentSize className='chart-view-content__chart'>
//                 { ({ width, height}) => <PieChart width={width} height={height} {...content} /> }
//             </ParentSize>
//         </div>
//     );
// };
