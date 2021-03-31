import React, { FunctionComponent } from 'react';
import {BarChartContent, ChartContent, LineChartContent} from 'sourcegraph';
import * as H from 'history';
import { ParentSize } from '@visx/responsive';
import { LineChart } from './charts/line/LineChart';

/**
 * Displays chart view content.
 */
export interface ChartViewContentProps {
    content: ChartContent
    animate?: boolean
    history: H.History
    viewID: string
}

export const ChartViewContent: FunctionComponent<ChartViewContentProps> = ({ content, ...props }) => (
    <>
        {content.chart === 'line' || content.chart === 'bar' ? (
            <CartesianChartViewContent {...props} content={content} />
        ) : content.chart === 'pie' ? (
            <PieChartViewContent {...props} content={content} />
        ) : null}
    </>
)

export interface CartesianChartViewContentProps extends ChartViewContentProps {
    content: LineChartContent<any, string> | BarChartContent<any, string>;
}

export const CartesianChartViewContent: FunctionComponent<CartesianChartViewContentProps> = props => {
    const { content } = props;

    if (content.chart === 'bar') {
      return (<div>Bar chart mock</div>);
    }

    return (
        <div className="chart-view-content" >
            <ParentSize className='chart-view-content__chart'>
                { ({ width, height}) => <LineChart width={width} height={height} {...content} /> }
            </ParentSize>
        </div>
    );
};

export const PieChartViewContent: FunctionComponent<ChartViewContentProps> = props => {
    return (
        <div>Pie chart mock</div>
    );
};
