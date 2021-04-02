import React, { ReactElement } from 'react';
import { RenderTooltipParams } from '@visx/xychart/lib/components/Tooltip';
import { Accessors } from '../types'
import { LineChartContent } from 'sourcegraph'

export interface TooltipContentProps extends RenderTooltipParams<any>{
    accessors: Accessors<any, string>;
    series: LineChartContent<any, string>['series'];
    className?: string;
}

export function TooltipContent(props: TooltipContentProps): ReactElement | null {
  const { className = '', tooltipData, accessors, series } = props;
  const datum = tooltipData?.nearestDatum?.datum;

  if (!datum) {
      return null;
  }

  const dateString = new Date(accessors.x(datum)).toDateString();
  const lineKeys = Object.keys(tooltipData?.datumByKey ?? {}).filter(lineKey => lineKey);

  return (
      <div className={`line-chart__tooltip-content ${className}`}>

          <h3 className='line-chart__tooltip-date'>
              {dateString}
          </h3>

          {/** values */}
          <ul className='line-chart__tooltip-list'>
              {
                  lineKeys.map(lineKey => {
                      const value = accessors.y[lineKey](datum);
                      const line = series.find(line => line.dataKey === lineKey)
                      const datumKey = tooltipData?.nearestDatum?.key;

                      /* eslint-disable react/forbid-dom-props */
                      return (
                          <li
                              key={lineKey}
                              className='line-chart__tooltip-item'>

                              <em
                                  className='line-chart__tooltip-item-name'
                                  style={{
                                      color: line?.stroke,
                                      textDecoration: datumKey === lineKey ? 'underline' : undefined,
                                  }}
                              >
                                  {line?.name ?? 'unknown series'}
                              </em>
                              {' '}
                              <span className='line-chart__tooltip-item-value'>
                                  {
                                      value === null || Number.isNaN(value)
                                          ? '–'
                                          : value
                                  }
                              </span>

                          </li>
                      );
                  })
              }
          </ul>
      </div>
  );
}

