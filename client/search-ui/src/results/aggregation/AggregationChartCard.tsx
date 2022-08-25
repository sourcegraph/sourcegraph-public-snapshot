import { HTMLAttributes, ReactElement, MouseEvent } from 'react'

import { ParentSize } from '@visx/responsive'
import classNames from 'classnames'

import { SearchAggregationMode } from '@sourcegraph/shared/src/graphql-operations'
import { Text, Link, Tooltip } from '@sourcegraph/wildcard'

import { SearchAggregationDatum } from '../../graphql-operations'

import { AggregationChart } from './AggregationChart'

import styles from './AggregationChartCard.module.scss'

const getName = (datum: SearchAggregationDatum): string => datum.label ?? ''
const getValue = (datum: SearchAggregationDatum): number => datum.count
const getColor = (datum: SearchAggregationDatum): string => (datum.label ? 'var(--primary)' : 'var(--text-muted)')
const getLink = (datum: SearchAggregationDatum): string => datum.query ?? ''

export enum AggregationCardMode {
    Data,
    Loading,
    Error,
}

type AggregationChartCardStateProps =
    | { type: AggregationCardMode.Data; data: SearchAggregationDatum[] }
    | { type: AggregationCardMode.Loading }
    | { type: AggregationCardMode.Error; errorMessage: string }

interface AggregationChartCardSharedProps extends HTMLAttributes<HTMLDivElement> {
    mode?: SearchAggregationMode | null
    missingCount?: number
    onBarLinkClick?: (query: string) => void
}

type AggregationChartCardProps = AggregationChartCardSharedProps & AggregationChartCardStateProps

export function AggregationChartCard(props: AggregationChartCardProps): ReactElement {
    const { type, mode, className, missingCount, onBarLinkClick, 'aria-label': ariaLabel } = props

    const handleDatumLinkClick = (event: MouseEvent, datum: SearchAggregationDatum): void => {
        event.preventDefault()
        onBarLinkClick?.(getLink(datum))
    }

    if (type === AggregationCardMode.Loading || type === AggregationCardMode.Error) {
        return (
            <div className={classNames(styles.container, className)}>
                <div className={styles.chartOverlay}>
                    {type === AggregationCardMode.Loading ? (
                        'Loading...'
                    ) : (
                        <div>
                            We couldn’t provide aggregation for this query. {props.errorMessage}.{' '}
                            <Link to="">Learn more</Link>
                        </div>
                    )}
                </div>
            </div>
        )
    }

    return (
        <ParentSize className={classNames(className, styles.container)}>
            {parent => (
                <>
                    <AggregationChart
                        aria-label={ariaLabel}
                        width={parent.width}
                        height={parent.height}
                        data={props.data}
                        mode={mode}
                        getDatumValue={getValue}
                        getDatumColor={getColor}
                        getDatumName={getName}
                        getDatumLink={getLink}
                        onDatumLinkClick={handleDatumLinkClick}
                    />

                    {!!missingCount && (
                        <Tooltip content={`Aggregation is not exhaustive, there are ${missingCount} groups more`}>
                            <Text size="small" className={styles.missingLabelCount}>
                                +{missingCount}
                            </Text>
                        </Tooltip>
                    )}
                </>
            )}
        </ParentSize>
    )
}
