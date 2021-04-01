import React, { ReactElement } from 'react';
import { RenderTooltipParams } from '@visx/xychart/lib/components/Tooltip';
import { Accessors } from '../types'
import { LineChartContent } from 'sourcegraph'

export interface TooltipContentProps extends RenderTooltipParams<any>{
    accessors: Accessors<any, string>;
    series: LineChartContent<any, string>['series']
}

export function TooltipContent(props: TooltipContentProps): ReactElement {
  const { tooltipData, colorScale, accessors, series } = props;

  return (
      <>
          {/** date */}
          {(tooltipData?.nearestDatum?.datum &&
              new Date(accessors.x(tooltipData?.nearestDatum?.datum)).toDateString()) ||
          'No date'}
          <br/>
          <br/>
          {/** values */}
          {(Object.keys(tooltipData?.datumByKey ?? {}).filter(lineKey => lineKey) as any[]).map(lineKey => {
              const value =
                  tooltipData?.nearestDatum?.datum &&
                  accessors.y[lineKey](
                      tooltipData?.nearestDatum?.datum,
                  );

              const line = series.find(line => line.dataKey === lineKey)

              /* eslint-disable react/forbid-dom-props */
              return (
                  <div
                      className='line-chart__tooltip'
                      key={lineKey}>

                      <em
                          className='line-chart__tooltip-text'
                          style={{
                              color: colorScale?.(lineKey),
                              textDecoration:
                                  tooltipData?.nearestDatum?.key === lineKey ? 'underline' : undefined,
                          }}
                      >
                          {line?.name ?? 'unknown series'}
                      </em>{' '}
                      {
                          value === null || Number.isNaN(value)
                              ? '–'
                              : value
                      }
                  </div>
              );
          })}
      </>
  );
}

