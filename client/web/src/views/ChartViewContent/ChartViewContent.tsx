import React, { FunctionComponent } from 'react';
import {BarChartContent, ChartContent, LineChartContent, PieChartContent} from 'sourcegraph';
import * as H from 'history';
import { ParentSize } from '@visx/responsive';

import { LineChart } from './charts/line/LineChart';
import { PieChart } from './charts/pie/PieChart';

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

interface CartesianChartViewContentProps extends ChartViewContentProps {
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

interface PieChartViewContentProps extends ChartViewContentProps {
    content: PieChartContent<any>
}

export const PieChartViewContent: FunctionComponent<PieChartViewContentProps> = props => {
    const { content } = props;

    return (
        <div className="chart-view-content" >
            <ParentSize className='chart-view-content__chart'>
                { ({ width, height}) => <PieChart width={width} height={height} {...content} /> }
            </ParentSize>
        </div>
    );
};
