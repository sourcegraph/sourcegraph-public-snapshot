import { HTMLAttributes, ReactElement, MouseEvent, FC, SVGProps, forwardRef } from 'react'

import { ParentSize } from '@visx/responsive'
import classNames from 'classnames'

import { ErrorAlert, ErrorMessage } from '@sourcegraph/branded/src/components/alerts'
import { SearchAggregationMode } from '@sourcegraph/shared/src/graphql-operations'
import { Text, Link, Tooltip, ForwardReferenceComponent } from '@sourcegraph/wildcard'

import { GetSearchAggregationResult, SearchAggregationDatum } from '../../graphql-operations'

import { AggregationChart } from './AggregationChart'

import styles from './AggregationChartCard.module.scss'

const getName = (datum: SearchAggregationDatum): string => datum.label ?? ''
const getValue = (datum: SearchAggregationDatum): number => datum.count
const getColor = (datum: SearchAggregationDatum): string => (datum.label ? 'var(--primary)' : 'var(--text-muted)')
const getLink = (datum: SearchAggregationDatum): string => datum.query ?? ''

/**
 * Nested aggregation results types from {@link AGGREGATION_SEARCH_QUERY} GQL query
 */
type SearchAggregationResult = GetSearchAggregationResult['searchQueryAggregate']['aggregations']

function getAggregationError(aggregation?: SearchAggregationResult): Error | undefined {
    if (aggregation?.__typename === 'SearchAggregationNotAvailable') {
        return new Error(aggregation.reason)
    }

    return
}

export function getAggregationData(aggregations: SearchAggregationResult): SearchAggregationDatum[] {
    switch (aggregations?.__typename) {
        case 'ExhaustiveSearchAggregationResult':
        case 'NonExhaustiveSearchAggregationResult':
            return aggregations.groups

        default:
            return []
    }
}

export function getOtherGroupCount(aggregations: SearchAggregationResult): number {
    switch (aggregations?.__typename) {
        case 'ExhaustiveSearchAggregationResult':
            return aggregations.otherGroupCount ?? 0
        case 'NonExhaustiveSearchAggregationResult':
            return aggregations.approximateOtherGroupCount ?? 0

        default:
            return 0
    }
}

interface AggregationChartCardProps extends HTMLAttributes<HTMLDivElement> {
    data?: SearchAggregationResult
    error?: Error
    loading: boolean
    mode?: SearchAggregationMode | null
    size?: 'sm' | 'md'
    onBarLinkClick?: (query: string) => void
}

export function AggregationChartCard(props: AggregationChartCardProps): ReactElement | null {
    const { data, error, loading, mode, className, size = 'sm', 'aria-label': ariaLabel, onBarLinkClick } = props

    if (loading) {
        return (
            <DataLayoutContainer size={size} className={classNames(styles.loading, className)}>
                Loading...
            </DataLayoutContainer>
        )
    }

    // Internal error
    if (error) {
        return (
            <DataLayoutContainer size={size} className={className}>
                <ErrorAlert error={error} className={styles.errorAlert} />
            </DataLayoutContainer>
        )
    }

    const aggregationError = getAggregationError(data)

    if (aggregationError) {
        return (
            <DataLayoutContainer
                data-error-layout={true}
                size={size}
                className={classNames(styles.aggregationErrorContainer, className)}
            >
                <ZeroStateBarsBackground />
                <div className={styles.errorMessageLayout}>
                    <div className={styles.errorMessage}>
                        We couldn’t provide an aggregation for this query. <ErrorMessage error={aggregationError} />.{' '}
                        <Link to="">Learn more</Link>
                    </div>
                </div>
            </DataLayoutContainer>
        )
    }

    if (!data) {
        return null
    }

    const missingCount = getOtherGroupCount(data)
    const handleDatumLinkClick = (event: MouseEvent, datum: SearchAggregationDatum): void => {
        event.preventDefault()
        onBarLinkClick?.(getLink(datum))
    }

    return (
        <ParentSize className={classNames(className, styles.container)}>
            {parent => (
                <>
                    <AggregationChart
                        aria-label={ariaLabel}
                        width={parent.width}
                        height={parent.height}
                        data={getAggregationData(data)}
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

interface DataLayoutContainerProps {
    size?: 'sm' | 'md'
}

const DataLayoutContainer = forwardRef((props, ref) => {
    const { as: Component = 'div', size = 'md', className, ...attributes } = props

    return (
        <Component
            {...attributes}
            ref={ref}
            className={classNames(className, styles.errorContainer, {
                [styles.errorContainerSmall]: size === 'sm',
            })}
        />
    )
}) as ForwardReferenceComponent<'div', DataLayoutContainerProps>

const ZeroStateBarsBackground: FC<SVGProps<SVGSVGElement>> = ({ className, ...attributes }) => (
    <svg
        {...attributes}
        className={classNames(className, styles.zeroStateBackground)}
        xmlns="http://www.w3.org/2000/svg"
    >
        <rect x="0%" y="5%" height="95%" width="5%" />
        <rect x="7%" y="20%" height="80%" width="5%" />
        <rect x="14%" y="25%" height="75%" width="5%" />
        <rect x="21%" y="30%" height="70%" width="5%" />
        <rect x="28%" y="32%" height="68%" width="5%" />
        <rect x="35%" y="32%" height="68%" width="5%" />
        <rect x="42%" y="38%" height="62%" width="5%" />
        <rect x="49%" y="45%" height="55%" width="5%" />
        <rect x="56%" y="60%" height="40%" width="5%" />
        <rect x="63%" y="62%" height="38%" width="5%" />
        <rect x="70%" y="67%" height="33%" width="5%" />
        <rect x="77%" y="70%" height="30%" width="5%" />
        <rect x="84%" y="75%" height="25%" width="5%" />
        <rect x="91%" y="85%" height="15%" width="5%" />
        <rect x="98%" y="93%" height="7%" width="5%" />
    </svg>
)
